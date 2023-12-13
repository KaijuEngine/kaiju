package main

import (
	"kaiju/bootstrap"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/rendering"
	"runtime"
	"time"
	"unsafe"
)

const TriangleShaderDataSize = int(unsafe.Sizeof(TriangleShaderData{}))

type TriangleShaderData struct {
	rendering.ShaderDataBase
	Color matrix.Color
}

func init() {
	runtime.LockOSThread()
}

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
			Color:    matrix.ColorWhite(),
		}, {
			Position: matrix.Vec3{0.5, -0.5, 0.0},
			Color:    matrix.ColorWhite(),
		}, {
			Position: matrix.Vec3{0.0, 0.5, 0.0},
			Color:    matrix.ColorWhite(),
		},
	}
	mesh := rendering.Mesh{}
	host.Window.Renderer.CreateMesh(&mesh, verts, []uint32{0, 1, 2})
	drawGroup := rendering.NewDrawInstanceGroup(shader, &mesh, TriangleShaderDataSize)
	{
		t := TriangleShaderData{Color: matrix.ColorRed()}
		t.Model = matrix.Mat4Identity()
		t.Model.Translate(matrix.Vec3{-0.51, 0, 0})
		drawGroup.AddInstance(&t)
	}
	{
		t := TriangleShaderData{Color: matrix.ColorBlue()}
		t.Model = matrix.Mat4Identity()
		t.Model.Translate(matrix.Vec3{0.51, 0, 0})
		drawGroup.AddInstance(&t)
	}
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
		host.Window.Renderer.ReadyFrame(host.Camera, float32(host.Runtime()))
		host.Window.Renderer.Draw([]rendering.DrawInstanceGroup{drawGroup})
		host.Window.SwapBuffers()
	}
}
