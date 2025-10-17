package stage_workspace

import (
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/games/editor/editor_controls"
	"kaiju/games/editor/editor_workspace/common_workspace"
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
	w.gridShader.Deactivate()
	w.CommonClose()
}

func (w *Workspace) update(deltaTime float64) {
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
	w.camera.SetMode(editor_controls.EditorCameraMode2d, w.Host)
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
}
