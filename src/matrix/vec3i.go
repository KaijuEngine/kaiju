package matrix

type Vec3i [3]int32

func (v Vec3i) X() int32 { return v[0] }
func (v Vec3i) Y() int32 { return v[1] }
func (v Vec3i) Z() int32 { return v[2] }
