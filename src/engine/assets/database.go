/******************************************************************************/
/* database.go                                                                */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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

// Database defines the contract for asset storage back‑ends used by the engine.
// Implementations may store assets on disk, in memory, inside an archive, or
// provide a debug view of the file system. The interface abstracts common
// operations required by the engine and editor.
type Database interface {
	// PostWindowCreate is a hook that is called after a window has been
	// created. Implementations can use the provided handle to perform any
	// platform‑specific initialisation. Most implementations are no‑ops.
	PostWindowCreate(windowHandle PostWindowCreateHandle) error

	// Cache stores the raw byte slice `data` under `key` for fast subsequent
	// reads. Implementations may choose to ignore this (e.g. DebugContentDatabase)
	// or provide a real cache (e.g. FileDatabase).
	Cache(key string, data []byte)

	// CacheRemove removes a cached entry identified by `key`. If the
	// implementation does not maintain a cache this is a no‑op.
	CacheRemove(key string)

	// CacheClear clears all cached entries. Implementations without a cache
	// should simply return.
	CacheClear()

	// Read returns the raw bytes for the asset identified by `key`. An error is
	// returned if the asset cannot be found or read.
	Read(key string) ([]byte, error)

	// ReadText is a convenience wrapper around Read that returns the asset as a
	// string. It returns any error from Read.
	ReadText(key string) (string, error)

	// Exists reports whether an asset with the given `key` is available in the
	// underlying storage. Implementations should check both the cache and the
	// backing store.
	Exists(key string) bool

	// Close releases any resources held by the implementation. Implementations
	// that have no resources may implement this as a no‑op.
	Close()
}
