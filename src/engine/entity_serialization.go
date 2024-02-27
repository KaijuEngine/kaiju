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

type drawingDef struct {
	CanvasId         string
	ShaderDefinition string
	MeshKey          string
	Textures         []string
	UseBlending      bool
	ShaderData       rendering.DrawInstance
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

func (e *Entity) Serialize(stream io.Writer) error {
	if e.IsDestroyed() {
		return errors.New("destroyed entities cannot be serialized")
	}
	enc := gob.NewEncoder(stream)
	var p, r, s = e.Transform.Position(), e.Transform.Rotation(), e.Transform.Scale()
	if err := enc.Encode(e.id); err != nil {
		return err
	} else if err := enc.Encode(p); err != nil {
		return err
	} else if err = enc.Encode(r); err != nil {
		return err
	} else if err = enc.Encode(s); err != nil {
		return err
	} else if err = enc.Encode(e.name); err != nil {
		return err
	} else if err = enc.Encode(e.isActive); err != nil {
		return err
	} else if err = enc.Encode(e.deactivatedFromParent); err != nil {
		return err
	} else if err = enc.Encode(e.orderedChildren); err != nil {
		return err
	}
	return e.EditorBindings.serialize(enc)
}

func (e *Entity) Deserialize(stream io.Reader, host *Host) error {
	dec := gob.NewDecoder(stream)
	var p, r, s matrix.Vec3
	var drawingDefs []drawingDef
	if err := dec.Decode(&e.id); err != nil {
		return err
	} else if err := dec.Decode(&p); err != nil {
		return err
	} else if err = dec.Decode(&r); err != nil {
		return err
	} else if err = dec.Decode(&s); err != nil {
		return err
	} else if err = dec.Decode(&e.name); err != nil {
		return err
	} else if err = dec.Decode(&e.isActive); err != nil {
		return err
	} else if err = dec.Decode(&e.deactivatedFromParent); err != nil {
		return err
	} else if err = dec.Decode(&e.orderedChildren); err != nil {
		return err
	} else if err = dec.Decode(&drawingDefs); err != nil {
		return err
	}
	e.Transform.SetPosition(p)
	e.Transform.SetRotation(r)
	e.Transform.SetScale(s)
	if drawings, err := setupDrawings(e, host, drawingDefs); err != nil {
		return err
	} else {
		return e.EditorBindings.deserialize(e, dec, host, drawings)
	}
}
