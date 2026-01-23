package content_previews

import (
	"bytes"
	"image"
	"image/png"
	"kaiju/engine"
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
)

func GenerateMeshPreview(host *engine.Host, km kaiju_mesh.KaijuMesh, onComplete func(image.Image, error)) error {
	offset := matrix.Vec3Zero()
	points := make([]matrix.Vec3, len(km.Verts))
	for i := range km.Verts {
		points[i] = km.Verts[i].Position
		offset.AddAssign(km.Verts[i].Position)
	}
	offset = offset.Shrink(matrix.Float(len(km.Verts))).Negative()
	for i := range km.Verts {
		points[i].AddAssign(offset)
	}
	box := collision.AABBFromPoints(points)
	// TODO:  Create a camera view/projection that fits the size of the render-
	// pass texture size and render to that, rather than doing this thing where
	// we move the mesh to the primary camera and fit into it.
	offset = host.PrimaryCamera().Position()
	offset.AddAssign(host.PrimaryCamera().Forward().Scale(box.Size().Length() * 1.5))
	for i := range km.Verts {
		km.Verts[i].Position = points[i].Add(offset)
	}
	mesh := rendering.NewMesh("tmp", km.Verts, km.Indexes)
	mesh.DelayedCreate(host.Window.Renderer)
	mat, err := host.MaterialCache().Material("ed_thumb_preview_mesh.material")
	if err != nil {
		return err
	}
	sd := shader_data_registry.Create(mat.Shader.ShaderDataName())
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
	}
	host.Drawings.AddDrawing(draw)
	host.RunAfterFrames(1, func() {
		pixels, err := mat.RenderPass().Texture(0).ReadAllPixels(host.Window.Renderer)
		sd.Destroy()
		if err != nil {
			onComplete(nil, err)
		} else if len(pixels) > 0 {
			tex := mat.RenderPass().Texture(0)
			w, h := tex.Width, tex.Height
			img := image.NewRGBA(image.Rect(0, 0, w, h))
			copy(img.Pix, pixels)
			var buf bytes.Buffer
			if encErr := png.Encode(&buf, img); encErr != nil {
				onComplete(nil, encErr)
				return
			}
			decodedImg, decErr := png.Decode(&buf)
			if decErr != nil {
				onComplete(nil, decErr)
				return
			}
			onComplete(decodedImg, nil)
		}
	})
	return nil
}
