package main

import (
	"kaiju/bootstrap"
	"kaiju/engine"
	"kaiju/gl"
	"kaiju/matrix"
	"kaiju/rendering"
	"time"
)

func main() {
	lastTime := time.Now()
	host, err := engine.NewHost()
	if err != nil {
		panic(err)
	}
	bootstrap.Main(&host)
	shader := host.ShaderCache.CreateShader("content/basic.vert", "content/basic.frag", "", "", "")
	verts := []rendering.Vertex{
		{
			Position: matrix.Vec3{-0.5, -0.5, 0.0},
			Color:    matrix.ColorGreen(),
		}, {
			Position: matrix.Vec3{0.5, -0.5, 0.0},
			Color:    matrix.ColorGreen(),
		}, {
			Position: matrix.Vec3{0.0, 0.5, 0.0},
			Color:    matrix.ColorGreen(),
		},
	}
	indices := []uint32{0, 1, 2}
	mesh := rendering.Mesh{}
	host.Window.Renderer.CreateMesh(&mesh, verts, indices)
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
		gl.ClearScreen()
		gl.UseProgram(shader.RenderId.(gl.Handle))
		gl.BindVertexArray(mesh.MeshId.(rendering.MeshIdGL).VAO)
		gl.DrawArrays(gl.Triangles, 0, 3)
		host.Window.SwapBuffers()
	}
}
