//go:build editor

/******************************************************************************/
/* entity.ed.go                                                               */
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
	"kaiju/editor/cache/project_cache"
	"kaiju/matrix"
	"kaiju/rendering"
)

const (
	editorDrawingBinding    = "drawing"
	editorDrawingDefinition = "drawingDefinition"
)

func init() {
	gob.Register(edDrawingDef{})
	gob.Register([]edDrawingDef(nil))
}

type edDrawingDef struct {
	CanvasId         string
	ShaderDefinition string
	MeshKey          string
	Textures         []string
	UseBlending      bool
	ShaderData       rendering.DrawInstance
}

type entityEditorBindings struct {
	data map[string]any
}

func (e *entityEditorBindings) init() {
	e.data = make(map[string]any)
}

func (e *entityEditorBindings) AddDrawing(drawing rendering.Drawing) {
	drawings := e.Drawings()
	defs := e.Data(editorDrawingDefinition)
	if drawings == nil {
		drawings = []rendering.Drawing{}
		e.Set(editorDrawingBinding, drawings)
		defs = []edDrawingDef{}
	}
	drawings = append(drawings, drawing)
	defs = append(defs.([]edDrawingDef), edDrawingDef{
		ShaderDefinition: drawing.Shader.Key,
		Textures:         rendering.TextureKeys(drawing.Textures),
		MeshKey:          drawing.Mesh.Key(),
		UseBlending:      drawing.UseBlending,
		ShaderData:       drawing.ShaderData,
		CanvasId:         drawing.CanvasId,
	})
	e.Set(editorDrawingBinding, drawings)
	e.Set(editorDrawingDefinition, defs)
}

func (e *entityEditorBindings) Drawings() []rendering.Drawing {
	if d, ok := e.data[editorDrawingBinding]; ok {
		return d.([]rendering.Drawing)
	} else {
		return nil
	}
}

func (e *entityEditorBindings) Set(key string, value any) {
	e.data[key] = value
}

func (e *entityEditorBindings) Data(key string) any {
	if d, ok := e.data[key]; ok {
		return d
	} else {
		return nil
	}
}

func (e *entityEditorBindings) Remove(key string) {
	delete(e.data, key)
}

func (e *Entity) EditorSerialize(stream io.Writer) error {
	if e.IsDestroyed() {
		return errors.New("destroyed entities cannot be serialized")
	}
	enc := gob.NewEncoder(stream)
	var p, r, s = e.Transform.Position(), e.Transform.Rotation(), e.Transform.Scale()
	cpyDrawings := e.EditorBindings.Drawings()
	e.EditorBindings.Remove(editorDrawingBinding)
	if err := enc.Encode(p); err != nil {
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
	} else if err = enc.Encode(e.EditorBindings.data); err != nil {
		return err
	}
	e.EditorBindings.Set(editorDrawingBinding, cpyDrawings)
	// TODO:  Serialize the parent id and all the child ids
	return nil
}

func (e *Entity) EditorDeserialize(stream io.Reader, host *Host) error {
	dec := gob.NewDecoder(stream)
	var p, r, s matrix.Vec3
	if err := dec.Decode(&p); err != nil {
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
	} else if err = dec.Decode(&e.EditorBindings.data); err != nil {
		return err
	}
	e.Transform.SetPosition(p)
	e.Transform.SetRotation(r)
	e.Transform.SetScale(s)
	if err := e.EditorBindings.setupDrawings(e, host); err != nil {
		return err
	}
	return nil
}

func (b *entityEditorBindings) setupDrawings(e *Entity, host *Host) error {
	defs := e.EditorBindings.Data(editorDrawingDefinition)
	if defs != nil {
		for _, d := range defs.([]edDrawingDef) {
			s := host.shaderCache.ShaderFromDefinition(d.ShaderDefinition)
			m, ok := host.MeshCache().FindMesh(d.MeshKey)
			if !ok {
				adi, err := asset_info.Lookup(d.MeshKey)
				if err != nil {
					return err
				}
				md, err := project_cache.LoadCachedMesh(adi)
				if err != nil {
					return err
				}
				m = host.MeshCache().Mesh(adi.ID, md.Verts, md.Indexes)
			}
			textures := make([]*rendering.Texture, len(d.Textures))
			for i, t := range d.Textures {
				tex, err := host.TextureCache().Texture(
					t, rendering.TextureFilterLinear)
				if err != nil {
					return err
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
			b.AddDrawing(drawing)
		}
	}
	return nil
}
