//go:build amd64

package matrix

//go:noescape
func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4
