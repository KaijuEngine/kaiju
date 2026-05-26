/******************************************************************************/
/* content_id.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_id

import (
	"kaijuengine.com/engine/encoding/pod"
)

type Css string
type Font string
type Html string
type Material string
type Mesh string
type Music string
type ParticleSystem string
type RenderPass string
type ShaderPipeline string
type Shader string
type Sound string
type TableOfContents string
type Template string
type Terrain string
type Texture string
type Stage string

func init() {
	pod.Register(Css(""))
	pod.Register(Font(""))
	pod.Register(Html(""))
	pod.Register(Material(""))
	pod.Register(Mesh(""))
	pod.Register(Music(""))
	pod.Register(ParticleSystem(""))
	pod.Register(RenderPass(""))
	pod.Register(ShaderPipeline(""))
	pod.Register(Shader(""))
	pod.Register(Sound(""))
	pod.Register(TableOfContents(""))
	pod.Register(Template(""))
	pod.Register(Terrain(""))
	pod.Register(Texture(""))
	pod.Register(Stage(""))
}
