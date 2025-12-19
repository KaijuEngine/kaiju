/******************************************************************************/
/* stage.go                                                                   */
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

package stages

import (
	"bytes"
	"kaiju/build"
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/runtime/encoding/gob"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"log/slog"
	"reflect"
)

const EntryPointAssetKey = "entryPointStage"

type Stage struct {
	Id       string
	Entities []EntityDescription
}

type StageJson struct {
	Id        string
	Meshes    []string                `json:",omitempty"`
	Materials []string                `json:",omitempty"`
	Textures  []string                `json:",omitempty"`
	Entities  []EntityDescriptionJson `json:",omitempty"`
}

type EntityDescriptionShaderDataField struct {
	Name  string
	Index int32
	Value any
}

type EntityDescription struct {
	Id             string
	TemplateId     string
	Name           string
	Mesh           string
	Material       string
	Textures       []string
	Position       matrix.Vec3
	Rotation       matrix.Vec3
	Scale          matrix.Vec3
	DataBinding    []EntityDataBinding
	Children       []EntityDescription
	ShaderData     []EntityDescriptionShaderDataField
	RawDataBinding []any
}

type EntityDescriptionJson struct {
	Id          string
	TemplateId  string
	Name        string
	Mesh        int
	Material    int                     `json:"Mat"`
	Textures    []int                   `json:"Tex,omitempty"`
	Position    matrix.Vec3             `json:"P"`
	Rotation    matrix.Vec3             `json:"R"`
	Scale       matrix.Vec3             `json:"S"`
	DataBinding []EntityDataBinding     `json:"Data,omitempty"`
	Children    []EntityDescriptionJson `json:"Kids,omitempty"`
	ShaderData  map[string]EntityDescriptionShaderDataField
}

type EntityDataBinding struct {
	RegistraionKey string
	Fields         map[string]any `json:",omitempty"`
}

func debugEnsureStructsMatch() {
	if build.Debug {
		ra := reflect.TypeFor[Stage]()
		rb := reflect.TypeFor[StageJson]()
		debug.Assert(ra.NumField() == rb.NumField()-3,
			"the Stage field has been modified but the matching StageSerialized was not updated")
		ea := reflect.TypeFor[EntityDescription]()
		eb := reflect.TypeFor[EntityDescriptionJson]()
		debug.Assert((ea.NumField()-1) == eb.NumField(), // -1 due to raw data field
			"the EntityDescription field has been modified but the matching EntityDescriptionSerialized was not updated")
	}
}

func (s *Stage) ToMinimized() StageJson {
	debugEnsureStructsMatch()
	ss := StageJson{
		Id:       s.Id,
		Entities: make([]EntityDescriptionJson, len(s.Entities)),
	}
	meshMap := map[string]int{}
	matMap := map[string]int{}
	texMap := map[string]int{}
	// Add a blank string into each for the case that they are not assigned
	meshMap[""] = 0
	matMap[""] = 0
	texMap[""] = 0
	for i := range s.Entities {
		meshMap[s.Entities[i].Mesh] = 0
		matMap[s.Entities[i].Material] = 0
		for j := range s.Entities[i].Textures {
			texMap[s.Entities[i].Textures[j]] = 0
		}
	}
	for k := range meshMap {
		meshMap[k] = len(ss.Meshes)
		ss.Meshes = append(ss.Meshes, k)
	}
	for k := range matMap {
		matMap[k] = len(ss.Materials)
		ss.Materials = append(ss.Materials, k)
	}
	for k := range texMap {
		texMap[k] = len(ss.Textures)
		ss.Textures = append(ss.Textures, k)
	}
	var proc func(from *EntityDescription, to *EntityDescriptionJson)
	proc = func(from *EntityDescription, to *EntityDescriptionJson) {
		to.Id = from.Id
		to.TemplateId = from.TemplateId
		to.Name = from.Name
		to.Position = from.Position
		to.Rotation = from.Rotation
		to.Scale = from.Scale
		to.DataBinding = from.DataBinding
		to.Mesh = meshMap[from.Mesh]
		to.Material = matMap[from.Material]
		to.Textures = make([]int, len(from.Textures))
		for i := range from.Textures {
			to.Textures[i] = texMap[from.Textures[i]]
		}
		to.ShaderData = make(map[string]EntityDescriptionShaderDataField)
		for i := range from.ShaderData {
			to.ShaderData[from.ShaderData[i].Name] = from.ShaderData[i]
		}
		to.Children = make([]EntityDescriptionJson, len(from.Children))
		for i := range from.Children {
			proc(&from.Children[i], &to.Children[i])
		}
	}
	for i := range s.Entities {
		proc(&s.Entities[i], &ss.Entities[i])
	}
	return ss
}

func (s *Stage) FromMinimized(ss StageJson) {
	debugEnsureStructsMatch()
	s.Id = ss.Id
	s.Entities = make([]EntityDescription, len(ss.Entities))
	var proc func(from *EntityDescriptionJson, to *EntityDescription)
	proc = func(from *EntityDescriptionJson, to *EntityDescription) {
		to.Id = from.Id
		to.TemplateId = from.TemplateId
		to.Name = from.Name
		to.Position = from.Position
		to.Rotation = from.Rotation
		to.Scale = from.Scale
		to.DataBinding = from.DataBinding
		to.Mesh = ss.Meshes[from.Mesh]
		to.Material = ss.Materials[from.Material]
		to.Textures = make([]string, len(from.Textures))
		for i := range from.Textures {
			to.Textures[i] = ss.Textures[from.Textures[i]]
		}
		for _, v := range from.ShaderData {
			to.ShaderData = append(to.ShaderData, v)
		}
		to.Children = make([]EntityDescription, len(from.Children))
		for i := range from.Children {
			proc(&from.Children[i], &to.Children[i])
		}
	}
	for i := range ss.Entities {
		proc(&ss.Entities[i], &s.Entities[i])
	}
}

func ArchiveDeserializer(rawData []byte) (Stage, error) {
	var s Stage
	err := gob.NewDecoder(bytes.NewReader(rawData)).Decode(&s)
	return s, err
}

func EntityDescriptionArchiveDeserializer(rawData []byte) (EntityDescription, error) {
	var desc EntityDescription
	err := gob.NewDecoder(bytes.NewReader(rawData)).Decode(&desc)
	return desc, err
}

func (s *Stage) Load(host *engine.Host) {
	entityBindings := []func(){}
	var proc func(se *EntityDescription, parent *engine.Entity)
	proc = func(se *EntityDescription, parent *engine.Entity) {
		e := host.NewEntity(host.WorkGroup())
		e.SetName(se.Name)
		if parent != nil {
			e.SetParent(parent)
		}
		e.Transform.SetPosition(se.Position)
		e.Transform.SetRotation(se.Rotation)
		e.Transform.SetScale(se.Scale)
		// TODO:  Entity data should have been serialized
		if build.Debug {
			for i := range se.DataBinding {
				b, ok := engine.DebugEntityDataRegistry[se.DataBinding[i].RegistraionKey]
				if ok {
					bi := reflect.ValueOf(b).Interface()
					nb := reflect.New(reflect.TypeOf(bi)).Elem()
					for k, v := range se.DataBinding[i].Fields {
						f := nb.FieldByName(k)
						engine.ReflectValueFromJson(v, f)
					}
					reflect.ValueOf(&b).Elem().Set(nb)
					entityBindings = append(entityBindings, func() {
						b.Init(e, host)
					})
				} else {
					slog.Error("failed to locate the registered key", "key", se.DataBinding[i].RegistraionKey)
				}
			}
		} else {
			for i := range se.RawDataBinding {
				entityBindings = append(entityBindings, func() {
					se.RawDataBinding[i].(engine.EntityData).Init(e, host)
				})
			}
		}
		if se.Mesh != "" {
			// TODO:  Handle error?
			SetupEntityFromDescription(e, host, se)
		}
		for i := range se.Children {
			proc(&se.Children[i], e)
		}
		// TODO:  Call the init for bound data after all have been created
	}
	for i := range s.Entities {
		proc(&s.Entities[i], nil)
	}
	for i := range entityBindings {
		entityBindings[i]()
	}
}

func SetupEntityFromDescription(e *engine.Entity, host *engine.Host, se *EntityDescription) (*engine.Entity, error) {
	ad := host.AssetDatabase()
	meshId := se.Mesh
	materialId := se.Material
	textureIds := se.Textures
	var km kaiju_mesh.KaijuMesh
	var err error
	// TODO:  Should these just be in the content database (no switch)?
	switch meshId {
	case "quad":
		km.Verts, km.Indexes = rendering.MeshQuadData()
	case "plane":
		km.Verts, km.Indexes = rendering.MeshPlaneData()
	default:
		var kmData []byte
		kmData, err = ad.Read(meshId)
		if err != nil {
			slog.Error("failed to load the mesh data", "id", meshId, "error", err)
			return nil, err
		}
		km, err = kaiju_mesh.Deserialize(kmData)
	}
	if err != nil {
		slog.Error("failed to deserialize the mesh data", "id", meshId, "error", err)
		return nil, err
	}
	mesh := host.MeshCache().Mesh(meshId, km.Verts, km.Indexes)
	var mat *rendering.Material
	if materialId == "" {
		slog.Warn("no material provided for SpawnMesh, will use fallback material")
		materialId = assets.MaterialDefinitionBasic
	}
	mat, err = host.MaterialCache().Material(materialId)
	if err != nil {
		slog.Error("failed to create the standard material", "error", err)
		return nil, err
	}
	texs := make([]*rendering.Texture, 0, len(textureIds))
	for i := range textureIds {
		texData, err := ad.Read(textureIds[i])
		if err != nil {
			slog.Error("failed to read the texture file", "id", textureIds[i], "error", err)
			return nil, err
		}
		// TODO:  Should be reading the filter from the configuration file
		tex, err := rendering.NewTextureFromMemory(textureIds[i],
			texData, 0, 0, rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the texture from it's data", "id", textureIds[i], "error", err)
			return nil, err
		}
		texs = append(texs, tex)
	}
	// TODO:  This should be based on the rendering.MaterialData texture count
	if len(textureIds) == 0 {
		slog.Warn("missing textures for mesh, using a fallback one")
		tex, err := host.TextureCache().Texture(assets.TextureSquare,
			rendering.TextureFilterLinear)
		if err != nil {
			slog.Error("failed to create the default texture", "error", err)
		}
		texs = append(texs, tex)
	}
	mat = mat.CreateInstance(texs)
	sd := shader_data_registry.Create(mat.Shader.ShaderDataName())
	for i := range texs {
		texs[i].DelayedCreate(host.Window.Renderer)
	}
	draw := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &e.Transform,
		ViewCuller: &host.Cameras.Primary,
	}
	// TODO:  Keeping this simple reflection for now so that this is flexible
	// for the future. I want to think this through and not be locked into any
	// one way of doing things.
	if len(se.ShaderData) > 0 {
		v := reflect.ValueOf(sd)
		for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			v = v.Elem()
		}
		if build.Debug {
			for i := range se.ShaderData {
				f := v.Field(int(se.ShaderData[i].Index))
				engine.ReflectValueFromJson(se.ShaderData[i].Value, f)
			}
		} else {
			for i := range se.ShaderData {
				f := v.Field(int(se.ShaderData[i].Index))
				f.Set(reflect.ValueOf(se.ShaderData[i].Value))
			}
		}
	}
	host.Drawings.AddDrawing(draw)
	e.OnDestroy.Add(func() { sd.Destroy() })
	return e, nil
}
