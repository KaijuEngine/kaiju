/******************************************************************************/
/* vertex_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestVertexFaceNormal(t *testing.T) {
	cases := []struct {
		name  string
		verts [3]Vertex
		want  matrix.Vec3
	}{
		{
			name: "counter-clockwise xy plane",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{1, 0, 0}},
				{Position: matrix.Vec3{0, 1, 0}},
			},
			want: matrix.Vec3Forward().Negative(),
		},
		{
			name: "clockwise xy plane",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{0, 1, 0}},
				{Position: matrix.Vec3{1, 0, 0}},
			},
			want: matrix.Vec3Forward(),
		},
		{
			name: "scaled triangle normalizes",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{2, 0, 0}},
				{Position: matrix.Vec3{0, 3, 0}},
			},
			want: matrix.Vec3Forward().Negative(),
		},
		{
			name: "degenerate",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{1, 1, 1}},
				{Position: matrix.Vec3{2, 2, 2}},
			},
			want: matrix.Vec3Zero(),
		},
	}
	for _, c := range cases {
		got := VertexFaceNormal(c.verts)
		if !matrix.Vec3Approx(got, c.want) {
			t.Fatalf("%s normal = %v, want %v", c.name, got, c.want)
		}
	}
}
