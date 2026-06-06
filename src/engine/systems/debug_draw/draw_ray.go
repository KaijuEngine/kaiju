/******************************************************************************/
/* draw_ray.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package debug

import (
	"log/slog"
	"time"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"

	"github.com/KaijuEngine/uuid"
)

func DrawRay(host *engine.Host, from, to matrix.Vec3, duration time.Duration) {
	// TODO:  Return the handle to delete this thing
	grid := rendering.NewMeshGrid(host.MeshCache(),
		"raycast_"+uuid.NewString(),
		[]matrix.Vec3{from, to}, matrix.ColorWhite())
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionGrid)
	if err != nil {
		slog.Error("failed to load the grid material for drawing raycast", "error", err)
		return
	}
	sd := &shader_data_registry.ShaderDataEdTransformWire{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.Color{0.5, 0.5, 0.5, 1},
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       grid,
		ShaderData: sd,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: &host.Cameras.Primary,
	})
	func() {
		time.Sleep(duration)
		sd.Destroy()
	}()
}
