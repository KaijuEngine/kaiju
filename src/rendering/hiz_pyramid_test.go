/******************************************************************************/
/* hiz_pyramid_test.go                                                        */
/******************************************************************************/

package rendering

import "testing"

func TestHiZPyramidLevelDimensions(t *testing.T) {
	got := hiZPyramidLevelDimensions(9, 5)
	want := [][2]int{{4, 2}, {2, 1}, {1, 1}}
	if len(got) != len(want) {
		t.Fatalf("level count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Width() != int32(want[i][0]) || got[i].Height() != int32(want[i][1]) {
			t.Fatalf("level %d = %dx%d, want %dx%d", i, got[i].Width(), got[i].Height(), want[i][0], want[i][1])
		}
	}
}

func TestHiZDispatchGroups(t *testing.T) {
	if got := hiZDispatchGroups(17, 9); got != ([3]uint32{3, 2, 1}) {
		t.Fatalf("dispatch groups = %v", got)
	}
}

func TestHiZReduceDepthUsesNearestDepth(t *testing.T) {
	if got := hiZReduceDepth(0.7, 0.2, 1.0, 0.4); got != 0.2 {
		t.Fatalf("Hi-Z reduction = %f, want nearest/min depth", got)
	}
	if got := hiZReduceDepth(); got != 1 {
		t.Fatalf("empty Hi-Z reduction = %f, want far clear depth", got)
	}
}
