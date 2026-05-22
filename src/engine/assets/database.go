/******************************************************************************/
/* database.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
