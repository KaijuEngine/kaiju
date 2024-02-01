package matrix

type Vec4i [4]int32

func (v Vec4i) X() int32 { return v[0] }
func (v Vec4i) Y() int32 { return v[1] }
func (v Vec4i) Z() int32 { return v[2] }
func (v Vec4i) W() int32 { return v[3] }
