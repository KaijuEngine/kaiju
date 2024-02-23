//go:build editor

package engine

import (
	"io"
	"kaiju/rendering"
)

const (
	editorDrawingBinding = "drawing"
)

type entityEditorBindings struct {
	data map[string]any
}

func (e *entityEditorBindings) AddDrawing(drawing rendering.Drawing) {
	drawings := e.Drawings()
	if drawings == nil {
		drawings = []rendering.Drawing{}
		e.Set(editorDrawingBinding, drawings)
	}
	drawings = append(drawings, drawing)
	e.Set(editorDrawingBinding, drawings)
}

func (e *entityEditorBindings) Drawings() []rendering.Drawing {
	if e.data == nil {
		e.data = make(map[string]any)
	}
	if d, ok := e.data[editorDrawingBinding]; ok {
		return d.([]rendering.Drawing)
	} else {
		return nil
	}
}

func (e *entityEditorBindings) Set(key string, value any) {
	if e.data == nil {
		e.data = make(map[string]any)
	}
	e.data[key] = value
}

func (e *entityEditorBindings) Data(key string) any {
	if e.data == nil {
		e.data = make(map[string]any)
	}
	if d, ok := e.data[key]; ok {
		return d
	} else {
		return nil
	}
}

func (e *entityEditorBindings) Remove(key string, value any) {
	if e.data == nil {
		return
	}
	delete(e.data, key)
}

func (e *Entity) EditorSerialize(stream io.Writer) error {
	//enc := gob.NewEncoder(stream)
	//return enc.Encode(mesh)
	return nil
}

func (e *Entity) EditorDeserialize(stream io.Reader) error {
	//dec := gob.NewDecoder(stream)
	//err = dec.Decode(&mesh)
	return nil
}
