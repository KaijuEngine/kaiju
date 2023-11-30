package main

import (
	"kaiju/bootstrap"
	"kaiju/engine"
	"kaiju/gl"
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
	verts := []float32{
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
		0.0, 0.5, 0.0,
	}
	var vao, vbo gl.Handle
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ArrayBuffer, vbo)
	gl.BufferData(gl.ArrayBuffer, verts, gl.StaticDraw)
	gl.VertexAttribPointer(0, 3, gl.Float, false, 0, nil)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ArrayBuffer, 0)
	gl.BindVertexArray(0)
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
		gl.ClearScreen()
		gl.UseProgram(shader.RenderId.(gl.Handle))
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.Triangles, 0, 3)
		host.Window.SwapBuffers()
	}
}
