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

func (t TriangleShaderData) Size() int {
	const size = int(unsafe.Sizeof(TriangleShaderData{}) - rendering.ShaderBaseDataStart)
	return size
}

func init() {
	runtime.LockOSThread()
}

func testDrawing(host *engine.Host) {
	positions := []matrix.Vec3{
		{-1, 0.0, 0.0},
		{1, 0.0, 0.0},
	}
	colors := []matrix.Color{
		{1.0, 0.0, 0.0, 1.0},
		{0.0, 1.0, 0.0, 1.0},
	}
	rots := []matrix.Float{45, -45}
	for i := 0; i < 2; i++ {
		shader := host.ShaderCache.CreateShader("content/basic.vert", "content/basic.frag", "", "", "")
		mesh := rendering.NewMeshQuad(&host.MeshCache)
		droidTex, _ := host.TextureCache.Texture("content/android.png", rendering.TextureFilterNearest)
		tsd := TriangleShaderData{Color: colors[i]}
		m := matrix.Mat4Identity()
		m.Rotate(matrix.Vec3{0.0, rots[i], 0.0})
		m.Translate(positions[i])
		tsd.SetModel(m)
		host.Drawings.AddDrawing(shader, rendering.Drawing{
			Shader:     shader,
			Mesh:       mesh,
			Textures:   []*rendering.Texture{droidTex},
			ShaderData: &tsd,
			Transform:  nil,
		})
	}
}

func main() {
	lastTime := time.Now()
	host, err := engine.NewHost()
	if err != nil {
		panic(err)
	}
	bootstrap.Main(&host)
	host.Camera.SetPosition(matrix.Vec3{0.0, 0.0, 2.0})
	testDrawing(&host)
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
		host.Render()
	}
}
