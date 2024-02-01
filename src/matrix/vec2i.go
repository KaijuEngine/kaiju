package matrix

type Vec2i [2]int32

func (v Vec2i) X() int32 { return v[0] }
func (v Vec2i) Y() int32 { return v[1] }
