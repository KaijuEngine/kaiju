/******************************************************************************/
/* cache_database.go                                                          */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package content_database

import (
	"io/fs"
	"kaiju/build"
	"kaiju/debug"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/systems/events"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
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
	return filepath.Base(c.Path)
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
	if idx, ok := c.lookup[id]; !ok {
		return CachedContent{}, NotInCacheError{Id: id}
	} else {
		return c.cache[idx], nil
	}
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

// Build will run through the project's file system config folder and scan all
// of the content configurations in the folder and load them into memory as part
// of the cache. While the build is running, searches and filters will work,
// but [Read] will not (due to it's mapping nature). You can use
// [OnBuildFinished] to know when the build has completed.
func (c *Cache) Build(pfs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Cache.Build").End()
	c.isBuilding.Store(true)
	if cap(c.cache) == 0 {
		c.cache = make([]CachedContent, 0, 1024)
		c.lookup = make(map[string]int, 1024)
	} else {
		klib.WipeSlice(c.cache)
		clear(c.lookup)
	}
	root := pfs.FullPath(project_file_system.ContentConfigFolder)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}
		if filepath.Base(info.Name()) == ".gitignore" {
			return err
		}
		p := filepath.Join(project_file_system.ContentConfigFolder, strings.TrimPrefix(path, root))
		return c.Index(p, pfs)
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
	if idx, ok := c.lookup[cc.Id()]; ok {
		c.cache[idx] = cc
	} else {
		c.cache = append(c.cache, cc)
		c.lookup[cc.Id()] = len(c.cache) - 1
	}
	return nil
}

// Remove will delete an entry from the cache (not the config), it is useful
// when content is being deleted from the project. This will delete the entry
// from the lookup as well. Deleting from the lookup makes it unstable since
// removing items from the cache will swap the deleted entry with the last entry
// and resize the length. This once last item will have the index of the removed
// entry and it's lookup will be updated.
func (c *Cache) Remove(id string) {
	defer tracing.NewRegion("Cache.Remove").End()
	if idx, ok := c.lookup[id]; ok {
		lastId := c.cache[len(c.cache)-1].Id()
		c.cache = klib.RemoveUnordered(c.cache, idx)
		delete(c.lookup, id)
		if len(c.cache) > 0 {
			c.lookup[lastId] = idx
			if build.Debug {
				debug.Assert(c.cache[idx].Id() == lastId,
					"the behavior of klib.RemoveUnordered must have changed!")
			}
		}
	}
}
