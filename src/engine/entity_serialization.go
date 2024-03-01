/******************************************************************************/
/* entity_serialization.go                                                    */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package engine

import (
	"encoding/gob"
	"errors"
	"io"
	"kaiju/assets/asset_info"
	"kaiju/cache/project_cache"
	"kaiju/matrix"
	"kaiju/rendering"
)

func init() {
	gob.Register(drawingDef{})
	gob.Register([]drawingDef(nil))
}

type entityStorage struct {
	Id                    string
	Position              matrix.Vec3
	Rotation              matrix.Vec3
	Scale                 matrix.Vec3
	Name                  string
	IsActive              bool
	DeactivatedFromParent bool
	OrderedChildren       bool
	Data                  []EntityData
}

type drawingDef struct {
	CanvasId         string
	ShaderDefinition string
	MeshKey          string
	Textures         []string
	UseBlending      bool
	ShaderData       rendering.DrawInstance
}

// Serialize will write the entity to the given stream and is reversed using
// #Deserialize. This will not serialize the children of the entity, that is
// the responsibility of the caller. All errors returned will be related to
// decoding the binary stream
func (e *Entity) Serialize(stream io.Writer) error {
	if e.IsDestroyed() {
		return errors.New("destroyed entities cannot be serialized")
	}
	enc := gob.NewEncoder(stream)
	var store entityStorage
	store.fromEntity(e)
	if err := enc.Encode(store); err != nil {
		return err
	}
	return e.EditorBindings.serialize(enc)
}

// Deserialize will read the entity from the given stream and is reversed using
// #Serialize. This will not deserialize the children of the entity, that is
// the responsibility of the caller. All errors returned will be related to
// decoding the binary stream
func (e *Entity) Deserialize(stream io.Reader, host *Host) error {
	dec := gob.NewDecoder(stream)
	var drawingDefs []drawingDef
	var store entityStorage
	if err := dec.Decode(&store); err != nil {
		return err
	} else if err = dec.Decode(&drawingDefs); err != nil {
		return err
	}
	store.toEntity(e)
	if drawings, err := setupDrawings(e, host, drawingDefs); err != nil {
		return err
	} else {
		return e.EditorBindings.deserialize(e, dec, host, drawings)
	}
}

func setupDrawings(e *Entity, host *Host, defs []drawingDef) ([]rendering.Drawing, error) {
	drawings := []rendering.Drawing{}
	for _, d := range defs {
		s := host.shaderCache.ShaderFromDefinition(d.ShaderDefinition)
		m, ok := host.MeshCache().FindMesh(d.MeshKey)
		if !ok {
			adi, err := asset_info.Lookup(d.MeshKey)
			if err != nil {
				return drawings, err
			}
			md, err := project_cache.LoadCachedMesh(adi)
			if err != nil {
				return drawings, err
			}
			m = host.MeshCache().Mesh(adi.ID, md.Verts, md.Indexes)
		}
		textures := make([]*rendering.Texture, len(d.Textures))
		for i, t := range d.Textures {
			tex, err := host.TextureCache().Texture(
				t, rendering.TextureFilterLinear)
			if err != nil {
				return drawings, err
			}
			textures[i] = tex
		}
		drawing := rendering.Drawing{
			Renderer:    host.Window.Renderer,
			Shader:      s,
			Mesh:        m,
			Textures:    textures,
			ShaderData:  d.ShaderData,
			Transform:   &e.Transform,
			CanvasId:    d.CanvasId,
			UseBlending: d.UseBlending,
		}
		host.Drawings.AddDrawing(&drawing)
		drawings = append(drawings, drawing)
	}
	return drawings, nil
}

func (s *entityStorage) fromEntity(e *Entity) {
	s.Id = string(e.id)
	s.Position = e.Transform.Position()
	s.Rotation = e.Transform.Rotation()
	s.Scale = e.Transform.Scale()
	s.Name = e.name
	s.IsActive = e.isActive
	s.DeactivatedFromParent = e.deactivatedFromParent
	s.OrderedChildren = e.orderedChildren
	s.Data = e.data
}

func (s *entityStorage) toEntity(e *Entity) {
	e.id = EntityId(s.Id)
	e.Transform.SetPosition(s.Position)
	e.Transform.SetRotation(s.Rotation)
	e.Transform.SetScale(s.Scale)
	e.name = s.Name
	e.isActive = s.IsActive
	e.deactivatedFromParent = s.DeactivatedFromParent
	e.orderedChildren = s.OrderedChildren
	e.data = s.Data
}
