package debug

import (
	"kaiju/assets"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"time"

	"github.com/KaijuEngine/uuid"
)

func DrawRay(host *engine.Host, from, to matrix.Vec3, duration time.Duration) {
	// TODO:  Return the handle to delete this thing
	grid := rendering.NewMeshGrid(host.MeshCache(),
		"raycast_"+uuid.New().String(),
		[]matrix.Vec3{from, to}, matrix.ColorWhite())
	shader := host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionGrid)
	sd := &rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.Color{0.5, 0.5, 0.5, 1},
	}
	host.Drawings.AddDrawing(&rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     shader,
		Mesh:       grid,
		ShaderData: sd,
	}, host.Window.Renderer.DefaultTarget())
	func() {
		time.Sleep(duration)
		sd.Destroy()
	}()
}
