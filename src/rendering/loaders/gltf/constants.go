package gltf

const (
	POSITION   = "POSITION"
	NORMAL     = "NORMAL"
	TANGENT    = "TANGENT"
	TEXCOORD_0 = "TEXCOORD_0"
	TEXCOORD_1 = "TEXCOORD_1"
	COLOR_0    = "COLOR_0"
	COLOR_1    = "COLOR_1"
	JOINTS_0   = "JOINTS_0"
	JOINTS_1   = "JOINTS_1"
	WEIGHTS_0  = "WEIGHTS_0"
	WEIGHTS_1  = "WEIGHTS_1"
)

type ComponentType = int32

const (
	BYTE           ComponentType = 5120
	UNSIGNED_BYTE  ComponentType = 5121
	SHORT          ComponentType = 5122
	UNSIGNED_SHORT ComponentType = 5123
	UNSIGNED_INT   ComponentType = 5125
	FLOAT          ComponentType = 5126
)

type AccessorType = string

const (
	SCALAR AccessorType = "SCALAR"
	VEC2   AccessorType = "VEC2"
	VEC3   AccessorType = "VEC3"
	VEC4   AccessorType = "VEC4"
	MAT2   AccessorType = "MAT2"
	MAT3   AccessorType = "MAT3"
	MAT4   AccessorType = "MAT4"
)
