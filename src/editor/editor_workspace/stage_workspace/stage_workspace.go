/******************************************************************************/
/* stage_workspace.go                                                         */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package stage_workspace

import (
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/rendering"
	"log/slog"
)

type Workspace struct {
	common_workspace.CommonWorkspace
	camera     editor_controls.EditorCamera
	updateId   int
	gridShader *rendering.ShaderDataBasic
}

func (w *Workspace) Initialize(host *engine.Host) {
	w.CommonWorkspace.InitializeWithUI(host, "editor/ui/workspace/stage_workspace.go.html", nil, nil)
	w.createViewportGrid()
	w.setupCamera()
}

func (w *Workspace) Open() {
	w.CommonOpen()
	w.gridShader.Activate()
	w.updateId = w.Host.Updater.AddUpdate(w.update)
}

func (w *Workspace) Close() {
	w.Host.Updater.RemoveUpdate(w.updateId)
	w.updateId = 0
	w.gridShader.Deactivate()
	w.CommonClose()
}

func (w *Workspace) update(deltaTime float64) {
	if w.IsBlurred {
		return
	}
	w.camera.Update(w.Host, deltaTime)
}

func (w *Workspace) createViewportGrid() {
	const gridCount = 20
	const halfGridCount = gridCount / 2
	material, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionGrid)
	if err != nil {
		slog.Error("failed to load the grid material", "error", err)
		return
	}
	points := make([]matrix.Vec3, 0, gridCount*4)
	for i := -halfGridCount; i <= halfGridCount; i++ {
		fi := float32(i)
		points = append(points, matrix.Vec3{fi, 0, -halfGridCount})
		points = append(points, matrix.Vec3{fi, 0, halfGridCount})
		points = append(points, matrix.Vec3{-halfGridCount, 0, fi})
		points = append(points, matrix.Vec3{halfGridCount, 0, fi})
	}
	grid := rendering.NewMeshGrid(w.Host.MeshCache(), "viewport_grid",
		points, matrix.Color{0.5, 0.5, 0.5, 1})
	w.gridShader = &rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.Color{0.5, 0.5, 0.5, 1},
	}
	w.Host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   w.Host.Window.Renderer,
		Material:   material,
		Mesh:       grid,
		ShaderData: w.gridShader,
	})
}

func (w *Workspace) setupCamera() {
	w.camera.OnModeChange.Add(func() {
		m := matrix.Mat4Identity()
		switch w.camera.Mode() {
		case editor_controls.EditorCameraMode3d:
			// Identity matrix is fine
		case editor_controls.EditorCameraMode2d:
			m.RotateX(90)
		}
		w.gridShader.SetModel(m)
	})
	w.camera.SetMode(editor_controls.EditorCameraMode3d, w.Host)
}
