/******************************************************************************/
/* cache_database.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

// Cache is an in-memory cache of all of the content in the project. It
// keeps an internal lookup so that it's quick to look up cached information on
// an asset by it's id. The structure is designed so that it keeps track of when
// it is building internally atomically which is set when the [Build] function
// is called. Removing items from the cache will cause a swap removal which
// means that the ids inside of the lookup are not stable and change with
// removals.
type Cache struct {
	cache           []CachedContent
	lookup          map[string]int
	isBuilding      atomic.Bool
	OnBuildFinished events.Event
	mutex           sync.RWMutex
}

// CachedContent is the content entry in the cache that is returned from lookups
// and searches.
type CachedContent struct {
	// Path is the location in the file system for this cached configuration.
	// You will typically want to use [content_database.ToContentPath] with this
	// path to get the content's location in the content folder.
	Path   string
	Config ContentConfig
}

func (cc CachedContent) ContentPath() string { return ToContentPath(cc.Path) }

// New will return a new instance of the cache database with it's members
// pre-sized to an arbitrary amount to speed up initial loading
func New() Cache {
	return Cache{
		cache:  make([]CachedContent, 0, 1024),
		lookup: make(map[string]int, 1024),
	}
}

// CachedContent will read the id from the file name and return it as a string
func (c *CachedContent) Id() string {
	contentPath := ToContentPath(c.Path)
	return filepath.Base(contentPath)
}

// List will return the internally held cached content slice.
func (c *Cache) List() []CachedContent { return c.cache }

// Read will try and locate the cached content data by id. This can fail if the
// content is not in the cache, in which case the caller should call [Index] to
// index the file. This can also fail if the cache is currently in the process
// of being built, in which case the caller should wait until it's done
// building and try again, or bind to the [OnBuildFinished] event.
func (c *Cache) Read(id string) (CachedContent, error) {
	defer tracing.NewRegion("Cache.Read").End()
	if c.isBuilding.Load() {
		return CachedContent{}, ReadDuringBuildError{}
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if idx, ok := c.lookup[id]; !ok {
		return CachedContent{}, NotInCacheError{Id: id}
	} else {
		return c.cache[idx], nil
	}
}

func (c *Cache) Rename(id, newName string, pfs *project_file_system.FileSystem) (CachedContent, error) {
	cc, err := c.Read(id)
	if err != nil {
		return cc, err
	}
	if cc.Config.Name == newName {
		return cc, CacheContentNameEqual
	}
	cc.Config.Name = newName
	if err := WriteConfig(cc.Path, cc.Config, pfs); err != nil {
		slog.Error("failed to update the content config file", "id", id, "error", err)
		return cc, err
	}
	c.IndexCachedContent(cc)
	return cc, nil
}

func (c *Cache) ListByType(typeName string) []CachedContent {
	defer tracing.NewRegion("Cache.ListByType").End()
	out := []CachedContent{}
	for i := range c.cache {
		if c.cache[i].Config.Type == typeName {
			out = append(out, c.cache[i])
		}
	}
	return out
}

// ReadLinked will return all of the linked content for the given id. This will
// also return the content for the id itself.
func (c *Cache) ReadLinked(id string) ([]CachedContent, error) {
	defer tracing.NewRegion("Cache.ReadLinked").End()
	cc, err := c.Read(id)
	if err != nil {
		return []CachedContent{}, err
	}
	if cc.Config.LinkedId == "" {
		return []CachedContent{cc}, nil
	}
	linked := []CachedContent{}
	for i := range c.cache {
		if c.cache[i].Config.LinkedId == cc.Config.LinkedId {
			linked = append(linked, c.cache[i])
		}
	}
	return linked, nil
}

// TagFilter will filter all content to that which matches the given tags. This
// is an OR comparison so that any content that has at least one of the tags
// will be selected by the filter.
func (c *Cache) TagFilter(tags []string) []CachedContent {
	defer tracing.NewRegion("Cache.TagFilter").End()
	out := []CachedContent{}
	for i := range c.cache {
		for j := range tags {
			if c.cache[i].Config.Tags.Contains(tags[j]) {
				out = append(out, c.cache[i])
			}
		}
	}
	return out
}

// TypeFilter will filter all content to that which matches the given types.
// This is an OR comparison so that any content that has at least one of the
// types will be selected by the filter.
func (c *Cache) TypeFilter(types []string) []CachedContent {
	defer tracing.NewRegion("Cache.TypeFilter").End()
	out := []CachedContent{}
	for i := range c.cache {
		for j := range types {
			if strings.EqualFold(c.cache[i].Config.Type, types[j]) {
				out = append(out, c.cache[i])
			}
		}
	}
	return out
}

// Search will filter all content by the given query and return any content that
// matches the query. Currently the only thing that is matched by this query
// is the developer-given name of the content. This is an exact match on part or
// all of the name (case-insensitive), fuzzy search may be introduced later.
func (c *Cache) Search(query string) []CachedContent {
	defer tracing.NewRegion("Cache.Search").End()
	out := []CachedContent{}
	q := strings.ToLower(query)
	for i := range c.cache {
		p := strings.ToLower(c.cache[i].Config.Name)
		if strings.Contains(p, q) {
			out = append(out, c.cache[i])
		}
	}
	return out
}

func (c *Cache) SearchSources(typeName, src string) []CachedContent {
	defer tracing.NewRegion("Cache.SearchSources").End()
	out := []CachedContent{}
	q := strings.ToLower(src)
	for i := range c.cache {
		if c.cache[i].Config.Type == typeName {
			p := strings.ToLower(c.cache[i].Config.SrcPath)
			if p == q {
				out = append(out, c.cache[i])
			}
		}
	}
	return out
}

// Build will run through the project's file system config folder and scan all
// of the content configurations in the folder and load them into memory as part
// of the cache. While the build is running, searches and filters will work,
// but [Read] will not (due to it's mapping nature). You can use
// [OnBuildFinished] to know when the build has completed.
func (c *Cache) Build(pfs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Cache.Build").End()
	c.isBuilding.Store(true)
	c.mutex.Lock()
	if cap(c.cache) == 0 {
		c.cache = make([]CachedContent, 0, 1024)
		c.lookup = make(map[string]int, 1024)
	} else {
		klib.WipeSlice(c.cache)
		clear(c.lookup)
	}
	c.mutex.Unlock()
	root := pfs.FullPath(project_file_system.ContentConfigFolder)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(info.Name())
		// Skip hidden / OS-droppings files (.DS_Store, ._*, .Trashes,
		// .Spotlight-V100, .gitignore, ...). They are never content
		// configs and the upgrade-rename in ReadConfig would otherwise
		// permanently append ".json" to them and trap the project on
		// every subsequent open (recovered from 2026-05-26 .DS_Store
		// lockout).
		if strings.HasPrefix(base, ".") {
			return nil
		}
		// Only .json files are content configs. Stray scratch / backup
		// files ("notes.txt", "scratch.bak") are silently ignored so a
		// single dropped file can not abort the entire project open.
		if filepath.Ext(base) != ".json" {
			return nil
		}
		p := filepath.Join(project_file_system.ContentConfigFolder, strings.TrimPrefix(path, root))
		if err := c.Index(p, pfs); err != nil {
			// Per-file decode errors warn + continue so a single
			// corrupt config can not lock the user out of the
			// project. The path surfaces in the log so the next
			// "blocked at startup" mystery is a one-line diagnosis.
			slog.Warn("cache build: skipping unreadable config",
				"path", p,
				"error", err)
			return nil
		}
		return nil
	})
	c.isBuilding.Store(false)
	c.OnBuildFinished.Execute()
	return err
}

// Index will insert the given content configuration into the cache. If the
// cache already contains the id for the path, then the cache will replace it's
// currently held values with the new values. This should be called when
// building the cache, importing new content to the project, or when the
// developer changes settings for content that alters the configuration.
func (c *Cache) Index(path string, pfs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Cache.Index").End()
	cfg, err := ReadConfig(path, pfs)
	if err != nil {
		return err
	}
	cc := CachedContent{
		Path:   path,
		Config: cfg,
	}
	c.IndexCachedContent(cc)
	return nil
}

func (c *Cache) IndexCachedContent(cc CachedContent) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if idx, ok := c.lookup[cc.Id()]; ok {
		c.cache[idx] = cc
	} else {
		c.cache = append(c.cache, cc)
		c.lookup[cc.Id()] = len(c.cache) - 1
	}
}

// Remove will delete an entry from the cache (not the config), it is useful
// when content is being deleted from the project. This will delete the entry
// from the lookup as well. Deleting from the lookup makes it unstable since
// removing items from the cache will swap the deleted entry with the last entry
// and resize the length. This once last item will have the index of the removed
// entry and it's lookup will be updated.
func (c *Cache) Remove(id string) {
	defer tracing.NewRegion("Cache.Remove").End()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if idx, ok := c.lookup[id]; ok {
		lastCacheIdx := len(c.cache) - 1
		delete(c.lookup, id)
		if lastCacheIdx == idx {
			c.cache = c.cache[:idx]
		} else {
			c.cache = klib.RemoveUnordered(c.cache, idx)
			if len(c.cache) > 0 {
				c.lookup[c.cache[idx].Id()] = idx
			}
		}
	}
}

func (c *Cache) ChangeGuid(from, to string, pfs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Cache.ChangeGuid").End()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// Check if the new id already exists in the cache
	if _, ok := c.lookup[to]; ok {
		return DuplicateIdError{Id: to}
	}

	// Read the current cached content without calling Read (re-locks)
	idx, ok := c.lookup[from]
	if !ok {
		err := NotInCacheError{Id: from}
		slog.Error("failed to read cached content for guid change", "from", from, "to", to, "error", err)
		return err
	}
	cc := c.cache[idx]

	// Build new paths with the new id
	dir := filepath.Dir(cc.Path)
	newConfigPath := filepath.Join(dir, to) + filepath.Ext(cc.Path)
	oldContentPath := cc.ContentPath()
	newContentPath := ToContentPath(newConfigPath)

	// Rename the config file
	if err := pfs.Rename(cc.Path, newConfigPath); err != nil {
		slog.Error("failed to rename config file", "from", cc.Path, "to", newConfigPath, "error", err)
		return err
	}

	// Rename the content file
	if err := pfs.Rename(oldContentPath, newContentPath); err != nil {
		slog.Error("failed to rename content file", "from", oldContentPath, "to", newContentPath, "error", err)
		// Attempt to rollback the config file rename
		if rollbackErr := pfs.Rename(newConfigPath, cc.Path); rollbackErr != nil {
			slog.Error("failed to rollback config file rename during content file rename error", "error", rollbackErr)
		}
		return err
	}

	// Update the cache inline without methods that re-lock
	delete(c.lookup, from)
	cc.Path = newConfigPath
	c.cache[idx] = cc
	c.lookup[to] = idx
	return nil
}
