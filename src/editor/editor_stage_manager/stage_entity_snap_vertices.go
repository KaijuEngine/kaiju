/******************************************************************************/
/* stage_entity_snap_vertices.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"slices"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func SnapVerticesFromMesh(verts []rendering.Vertex) []matrix.Vec3 {
	out := make([]matrix.Vec3, 0, len(verts))
	seen := make(map[matrix.Vec3]struct{}, len(verts))
	for i := range verts {
		pos := verts[i].Position
		if _, ok := seen[pos]; ok {
			continue
		}
		seen[pos] = struct{}{}
		out = append(out, pos)
	}
	return out
}

func snapVerticesFromMesh(verts []rendering.Vertex) []matrix.Vec3 {
	return SnapVerticesFromMesh(verts)
}

func (m *StageManager) VertexSnapTargetEntities(excludedRoots []*StageEntity) []*StageEntity {
	out := make([]*StageEntity, 0, len(m.entities))
	for _, e := range m.entities {
		if e == nil || e.IsDeleted() || e.IsLocked() || e.StageData.Mesh == nil ||
			len(e.StageData.SnapVertices) == 0 || stageEntityInHierarchy(e, excludedRoots) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func stageEntityInHierarchy(e *StageEntity, roots []*StageEntity) bool {
	return slices.ContainsFunc(roots, func(root *StageEntity) bool {
		return e == root || root != nil && e.HasParent(&root.Entity)
	})
}
