/******************************************************************************/
/* render_caches.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "kaijuengine.com/engine/assets"

type RenderCaches interface {
	ShaderCache() *ShaderCache
	TextureCache() *TextureCache
	MeshCache() *MeshCache
	FontCache() *FontCache
	MaterialCache() *MaterialCache
	AssetDatabase() assets.Database
}
