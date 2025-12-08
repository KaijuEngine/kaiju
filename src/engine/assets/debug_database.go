/******************************************************************************/
/* debug_database.go                                                          */
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

package assets

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/KaijuEngine/kaiju/platform/filesystem"
	"github.com/KaijuEngine/kaiju/platform/profiler/tracing"
)

type DebugContentDatabase struct{}

func (DebugContentDatabase) Cache(key string, data []byte) { /* No caching planned*/ }
func (DebugContentDatabase) CacheRemove(key string)        { /* No caching planned*/ }
func (DebugContentDatabase) CacheClear()                   { /* No caching planned*/ }
func (DebugContentDatabase) Close()                        {}

var cachedKeys = map[string]string{}

func findDebugDatabaseFile(key string) string {
	if path, ok := cachedKeys[key]; ok {
		return path
	}
	finalPath := ""
	paths := []string{"database/stock", "database/content", "database/debug"}
	for i := 0; i < len(paths) && finalPath == ""; i++ {
		filepath.Walk(paths[i], func(path string, info fs.FileInfo, err error) error {
			name := info.Name()
			cachedKeys[name] = path
			if finalPath != "" {
				return err
			}
			if name == key {
				finalPath = path
			}
			return err
		})
	}
	return finalPath
}

func (e DebugContentDatabase) Read(key string) ([]byte, error) {
	defer tracing.NewRegion("DebugContentDatabase.Read: " + key).End()
	if key[0] == absoluteFilePrefix {
		return filesystem.ReadFile(key[1:])
	}
	return os.ReadFile(findDebugDatabaseFile(key))
}

func (e DebugContentDatabase) ReadText(key string) (string, error) {
	defer tracing.NewRegion("DebugContentDatabase.ReadText: " + key).End()
	b, err := e.Read(key)
	return string(b), err
}

func (e DebugContentDatabase) Exists(key string) bool {
	defer tracing.NewRegion("DebugContentDatabase.Exists: " + key).End()
	if key[0] == absoluteFilePrefix {
		return filesystem.FileExists(key[1:])
	}
	_, err := os.Stat(findDebugDatabaseFile(key))
	return err == nil
}

func (DebugContentDatabase) PostWindowCreate(PostWindowCreateHandle) error { return nil }
