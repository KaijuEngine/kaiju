/******************************************************************************/
/* draw_ray.go                                                                */
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
		CanvasId:   "default",
	})
	func() {
		time.Sleep(duration)
		sd.Destroy()
	}()
}
