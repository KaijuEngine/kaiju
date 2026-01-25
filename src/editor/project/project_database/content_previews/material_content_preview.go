package content_previews

import (
	"encoding/json"
	"fmt"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
)

func (p *ContentPreviewer) updateSphereTransform() *matrix.Transform {
	host := p.ed.Host()
	cam := host.PrimaryCamera()
	ratio := cam.Width() / cam.Height()
	scaleOut := max(sphereRadius, ratio, 1/ratio) + 1.5
	spherePos := cam.Position().Add(cam.Forward().Scale(scaleOut))
	p.sphereTransform.SetPosition(spherePos)
	p.sphereTransform.SetScale(matrix.NewVec3(ratio, 1, ratio))
	p.sphereTransform.SetRotation(p.rotForMaterialTransform())
	return &p.sphereTransform
}

func (p *ContentPreviewer) renderMaterial(id string) {
	mat, err := readMaterial(id, p.ed)
	if err != nil {
		slog.Error("failed to generate a preview for material", "id", id, "error", err)
		p.completeProc()
		return
	}
	host := p.ed.Host()
	mesh := rendering.NewMeshSphere(host.MeshCache(), sphereRadius, sphereSegments, sphereSegments)
	sd := shader_data_registry.Create(mat.Shader.ShaderDataName())
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  p.updateSphereTransform(),
		ViewCuller: &host.Cameras.Primary,
	}
	host.Drawings.AddDrawing(draw)
	host.RunBeforeRender(func() {
		mat.Shader.DelayedCreate(host.Window.Renderer, host.AssetDatabase())
		host.RunAfterFrames(1, func() {
			p.readRenderPass(host, sd, id)
		})
	})
}

func (p *ContentPreviewer) rotForMaterialTransform() matrix.Vec3 {
	view := matrix.Mat4LookAt(p.sphereTransform.Position(),
		p.cam.Position(), matrix.Vec3Up())
	rot := matrix.Mat4Identity()
	rot[matrix.Mat4x0y0] = view[matrix.Mat4x0y0]
	rot[matrix.Mat4x0y1] = view[matrix.Mat4x1y0]
	rot[matrix.Mat4x0y2] = view[matrix.Mat4x2y0]
	rot[matrix.Mat4x1y0] = view[matrix.Mat4x0y1]
	rot[matrix.Mat4x1y1] = view[matrix.Mat4x1y1]
	rot[matrix.Mat4x1y2] = view[matrix.Mat4x2y1]
	rot[matrix.Mat4x2y0] = view[matrix.Mat4x0y2]
	rot[matrix.Mat4x2y1] = view[matrix.Mat4x2y1]
	rot[matrix.Mat4x2y2] = view[matrix.Mat4x2y2]
	rot[matrix.Mat4x0y2] *= -1
	rot[matrix.Mat4x1y2] *= -1
	rot[matrix.Mat4x2y2] *= -1
	q := matrix.QuaternionFromMat4(rot)
	return q.ToEuler()
}

func readMaterial(id string, ed EditorInterface) (*rendering.Material, error) {
	defer tracing.NewRegion("content_previews.readMaterial").End()
	cc, err := ed.Cache().Read(id)
	if err != nil {
		return nil, err
	}
	if cc.Config.Type != (content_database.Material{}).TypeName() {
		return nil, fmt.Errorf("can't generate a material preview image for content, the provided id '%s' is not a material", id)
	}
	matStr, err := ed.ProjectFileSystem().ReadFile(cc.ContentPath())
	if err != nil {
		return nil, err
	}
	key := "preview_" + id
	var materialData rendering.MaterialData
	if err := json.Unmarshal([]byte(matStr), &materialData); err != nil {
		slog.Error("failed to read the material", "material", key, "error", err)
		return nil, err
	}
	materialData.RenderPass = "ed_thumb_preview_mesh.renderpass"
	materialData.ShaderPipeline = "ed_thumb_preview_mesh.shaderpipeline"
	host := ed.Host()
	mat, err := materialData.CompileExt(host.AssetDatabase(), host.Window.Renderer, true)
	if err != nil {
		return nil, err
	}
	mat.Id = key
	return mat, nil
}
