package render_graph_workspace

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestShaderGraphSplineUpdateMarksTransformDirty(t *testing.T) {
	host := engine.NewHost("test", nil, assets.Database(nil))
	*host.MeshCache() = rendering.NewMeshCache(nil, nil)

	verts := make([]rendering.Vertex, (shaderGraphSplineSegments+1)*2)
	mesh := host.MeshCache().DynamicMesh("test-spline", verts, []uint32{0, 1, 2})
	spline := shaderGraphSpline{
		host:   host,
		mesh:   mesh,
		shader: &ui.ShaderData{ShaderDataBase: rendering.NewShaderDataBase()},
		verts:  verts,
		points: make([]matrix.Vec2, shaderGraphSplineSegments+1),
	}
	spline.transform.SetupRawTransform()
	spline.transform.ResetDirty()

	spline.Update(matrix.NewVec2(10, 10), matrix.NewVec2(160, 48))

	if !spline.transform.IsDirty() {
		t.Fatal("spline Update should mark the transform dirty after changing mesh bounds")
	}
}
