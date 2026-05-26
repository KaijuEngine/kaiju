/******************************************************************************/
/* codgen_registry.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package codegen

import (
	"reflect"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/matrix"
)

var (
	registry = make(map[string]reflect.Type)
)

func init() {
	RegisterTypeName("matrix.Float", matrix.Float(0))
	RegisterType[matrix.Color]()
	RegisterType[matrix.Color]()
	RegisterType[matrix.Mat3]()
	RegisterType[matrix.Mat4]()
	RegisterType[matrix.Mat4]()
	RegisterType[matrix.Quaternion]()
	RegisterType[matrix.Quaternion]()
	RegisterType[matrix.Transform]()
	RegisterType[matrix.Transform]()
	RegisterType[matrix.Vec2]()
	RegisterType[matrix.Vec2]()
	RegisterType[matrix.Vec2i]()
	RegisterType[matrix.Vec2i]()
	RegisterType[matrix.Vec3]()
	RegisterType[matrix.Vec3]()
	RegisterType[matrix.Vec3i]()
	RegisterType[matrix.Vec3i]()
	RegisterType[matrix.Vec4]()
	RegisterType[matrix.Vec4]()
	RegisterType[matrix.Vec4i]()
	RegisterType[engine.Entity]()
	RegisterType[engine.EntityId]()
	RegisterType[engine.Host]()
	RegisterType[engine.UpdateId]()
	RegisterType[content_id.Css]()
	RegisterType[content_id.Font]()
	RegisterType[content_id.Html]()
	RegisterType[content_id.Material]()
	RegisterType[content_id.Mesh]()
	RegisterType[content_id.Music]()
	RegisterType[content_id.ParticleSystem]()
	RegisterType[content_id.RenderPass]()
	RegisterType[content_id.ShaderPipeline]()
	RegisterType[content_id.Shader]()
	RegisterType[content_id.Sound]()
	RegisterType[content_id.TableOfContents]()
	RegisterType[content_id.Template]()
	RegisterType[content_id.Terrain]()
	RegisterType[content_id.Texture]()
	RegisterType[content_id.Stage]()
}

func RegisterType[T any]() {
	t := reflect.TypeFor[T]()
	registry[reflect.TypeFor[T]().String()] = t
}

func RegisterTypeName(name string, t any) {
	registry[name] = reflect.TypeOf(t)
}
