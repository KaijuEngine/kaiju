package stages

import "kaiju/matrix"

type EntityDescription struct {
	Id        string
	Rendering struct {
		Mesh     string
		Material string
		Textures []string `json:",omitempty"`
	}
	Transform struct {
		Position matrix.Vec3
		Rotation matrix.Vec3
		Scale    matrix.Vec3
	}
	DataBinding []EntityDataBinding `json:",omitempty"`
	Children    []EntityDescription `json:",omitempty"`
}

type EntityDataBinding struct {
	Name   string
	Fields map[string]any `json:",omitempty"`
}

type Stage struct {
	Id       string
	Entities []EntityDescription `json:",omitempty"`
}
