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

func testDrawing(host *engine.Host) {
	positions := []matrix.Vec3{
		{-0.6, 0.0, 0.0},
		{0.6, 0.0, 0.0},
	}
	colors := []matrix.Color{
		{1.0, 0.0, 0.0, 1.0},
		{0.0, 1.0, 0.0, 1.0},
	}
	for i := 0; i < 2; i++ {
		shader := host.ShaderCache.CreateShader("content/basic.vert", "content/basic.frag", "", "", "")
		mesh := rendering.NewMeshQuad(&host.MeshCache)
		drawGroup := rendering.NewDrawInstanceGroup(mesh, TriangleShaderDataSize)
		droidTex, _ := host.TextureCache.Texture("content/android.png", rendering.TextureFilterNearest)
		drawGroup.Textures = []*rendering.Texture{droidTex}
		tsd := TriangleShaderData{Color: colors[i]}
		tsd.Model = matrix.Mat4Identity()
		tsd.Model.Translate(positions[i])
		drawGroup.AddInstance(&tsd)
		host.Drawings.AddDrawing(shader, drawGroup)
	}
}

func main() {
	lastTime := time.Now()
	host, err := engine.NewHost()
	if err != nil {
		panic(err)
	}
	bootstrap.Main(&host)
	testDrawing(&host)
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
		host.Render()
	}
}
