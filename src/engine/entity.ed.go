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
	"kaiju/rendering"
	"log/slog"
	"slices"

	"github.com/KaijuEngine/uuid"
)

const (
	editorDrawingBinding    = "drawing"
	editorDrawingDefinition = "drawingDefinition"
)

type entityEditorBindings struct {
	data map[string]any
}

func (e *entityEditorBindings) init() {
	e.data = make(map[string]any)
}

// AddDrawing will add a drawing to the entity
//
// `EDITOR ONLY`
func (e *entityEditorBindings) AddDrawing(drawing rendering.Drawing) {
	drawings := e.Drawings()
	defs := e.Data(editorDrawingDefinition)
	if drawings == nil {
		drawings = []rendering.Drawing{}
		e.Set(editorDrawingBinding, drawings)
		defs = []drawingDef{}
	}
	drawings = append(drawings, drawing)
	defs = append(defs.([]drawingDef), drawingDef{
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

// Drawings will return the drawings associated with this entity
//
// `EDITOR ONLY`
func (e *entityEditorBindings) Drawings() []rendering.Drawing {
	if d, ok := e.data[editorDrawingBinding]; ok {
		return d.([]rendering.Drawing)
	} else {
		return nil
	}
}

// Set will set the data associated with the key
//
// `EDITOR ONLY`
func (e *entityEditorBindings) Set(key string, value any) {
	e.data[key] = value
}

// Data will return the data associated with the key, if it does not exist
// it will return nil
//
// `EDITOR ONLY`
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

// GenerateId will create a unique ID for this entity, if one already exists
// it will log an error and return the existing ID
//
// `EDITOR ONLY`
func (e *Entity) GenerateId() string {
	if e.id == "" {
		e.id = uuid.New().String()
	} else {
		slog.Error("Generating entity ID when one already exists")
	}
	return e.id
}

// AddData will add the entity data to the entity
//
// `EDITOR ONLY`
func (e *Entity) AddData(data EntityData) {
	e.data = append(e.data, data)
}

// RemoveData will remove the entity data from the entity
//
// `EDITOR ONLY`
func (e *Entity) RemoveData(idx int) {
	e.data = slices.Delete(e.data, idx, idx+1)
}

// ListData will return the entity data
//
// `EDITOR ONLY`
func (e *Entity) ListData() []EntityData { return e.data }

func (e *entityEditorBindings) serialize(enc *gob.Encoder) error {
	cpyDrawings := e.Drawings()
	drawingDefs := e.Data(editorDrawingDefinition).([]drawingDef)
	e.Remove(editorDrawingBinding)
	e.Remove(editorDrawingDefinition)
	if err := enc.Encode(drawingDefs); err != nil {
		return err
	} else if err := enc.Encode(e.data); err != nil {
		return err
	}
	e.Set(editorDrawingBinding, cpyDrawings)
	e.Set(editorDrawingDefinition, drawingDefs)
	return nil
}

func (e *entityEditorBindings) deserialize(entity *Entity,
	dec *gob.Decoder, host *Host, drawings []rendering.Drawing) error {
	if err := dec.Decode(&e.data); err != nil {
		return err
	}
	for i := range drawings {
		e.AddDrawing(drawings[i])
	}
	return nil
}

func (e *Entity) initialize(host *Host) {}
