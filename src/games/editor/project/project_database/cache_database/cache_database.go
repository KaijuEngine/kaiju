package cache_database

import (
	"io/fs"
	"kaiju/build"
	"kaiju/debug"
	"kaiju/engine/systems/events"
	"kaiju/games/editor/project/project_database/content_database"
	"kaiju/games/editor/project/project_file_system"
	"kaiju/klib"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
)

// CacheDatabase is an in-memory cache of all of the content in the project. It
// keeps an internal lookup so that it's quick to look up cached information on
// an asset by it's id. The structure is designed so that it keeps track of when
// it is building internally atomically which is set when the [Build] function
// is called. Removing items from the cache will cause a swap removal which
// means that the ids inside of the lookup are not stable and change with
// removals.
type CacheDatabase struct {
	cache           []CachedContent
	lookup          map[string]int
	isBuilding      atomic.Bool
	OnBuildFinished events.Event
}

// CachedContent is the content entry in the cache that is returned from lookups
// and searches.
type CachedContent struct {
	Path   string
	Config content_database.ContentConfig
}

// New will return a new instance of the cache database with it's members
// pre-sized to an arbitrary amount to speed up initial loading
func New() CacheDatabase {
	return CacheDatabase{
		cache:  make([]CachedContent, 0, 1024),
		lookup: make(map[string]int, 1024),
	}
}

// CachedContent will read the id from the file name and return it as a string
func (c *CachedContent) Id() string {
	return filepath.Base(c.Path)
}

// Read will try and locate the cached content data by id. This can fail if the
// content is not in the cache, in which case the caller should call [Index] to
// index the file. This can also fail if the cache is currently in the process
// of being built, in which case the caller should wait until it's done
// building and try again, or bind to the [OnBuildFinished] event.
func (c *CacheDatabase) Read(id string) (CachedContent, error) {
	if c.isBuilding.Load() {
		return CachedContent{}, ReadDuringBuildError{}
	}
	if idx, ok := c.lookup[id]; !ok {
		return CachedContent{}, NotInCacheError{Id: id}
	} else {
		return c.cache[idx], nil
	}
}

// TagFilter will filter all content to that which matches the given tags. This
// is an OR comparison so that any content that has at least one of the tags
// will be selected by the filter.
func (c *CacheDatabase) TagFilter(tags []string) []CachedContent {
	out := []CachedContent{}
	for i := range c.cache {
		for j := range tags {
			if slices.Contains(c.cache[i].Config.Tags, tags[j]) {
				out = append(out, c.cache[i])
			}
		}
	}
	return out
}

// TypeFilter will filter all content to that which matches the given types.
// This is an OR comparison so that any content that has at least one of the
// types will be selected by the filter.
func (c *CacheDatabase) TypeFilter(types []string) []CachedContent {
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
func (c *CacheDatabase) Search(query string) []CachedContent {
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

// Build will run through the project's file system config folder and scan all
// of the content configurations in the folder and load them into memory as part
// of the cache. While the build is running, searches and filters will work,
// but [Read] will not (due to it's mapping nature). You can use
// [OnBuildFinished] to know when the build has completed.
func (c *CacheDatabase) Build(pfs *project_file_system.FileSystem) error {
	c.isBuilding.Store(true)
	if cap(c.cache) == 0 {
		c.cache = make([]CachedContent, 0, 1024)
		c.lookup = make(map[string]int, 1024)
	} else {
		klib.WipeSlice(c.cache)
		clear(c.lookup)
	}
	err := filepath.Walk(pfs.FullPath(project_file_system.ContentConfigFolder), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}
		return c.Index(path, pfs)
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
func (c *CacheDatabase) Index(path string, pfs *project_file_system.FileSystem) error {
	cfg, err := content_database.ReadConfig(path, pfs)
	if err != nil {
		return err
	}
	cc := CachedContent{
		Path:   path,
		Config: cfg,
	}
	if idx, ok := c.lookup[cc.Id()]; ok {
		c.cache[idx] = cc
	} else {
		c.cache = append(c.cache, cc)
		c.lookup[cc.Id()] = len(c.cache)
	}
	return nil
}

// Remove will delete an entry from the cache (not the config), it is useful
// when content is being deleted from the project. This will delete the entry
// from the lookup as well. Deleting from the lookup makes it unstable since
// removing items from the cache will swap the deleted entry with the last entry
// and resize the length. This once last item will have the index of the removed
// entry and it's lookup will be updated.
func (c *CacheDatabase) Remove(id string) {
	if idx, ok := c.lookup[id]; ok {
		lastId := c.cache[len(c.cache)-1].Id()
		c.cache = klib.RemoveUnordered(c.cache, idx)
		c.lookup[lastId] = idx
		if build.Debug {
			debug.Assert(c.cache[c.lookup[lastId]].Id() == lastId,
				"the behavior of klib.RemoveUnordered must have changed!")
		}
		delete(c.lookup, id)
	}
}
