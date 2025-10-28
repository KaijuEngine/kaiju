/******************************************************************************/
/* stage.go                                                                   */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package stages

import (
	"kaiju/build"
	"kaiju/debug"
	"kaiju/matrix"
	"reflect"
)

// //////////////////////////////////////////////////////////////////////////////
type Stage struct {
	Id       string
	Entities []EntityDescription
}

type StageJson struct {
	Id        string
	Meshes    []string                `json:",omitempty"`
	Materials []string                `json:",omitempty"`
	Textures  []string                `json:",omitempty"`
	Entities  []EntityDescriptionJson `json:",omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

// //////////////////////////////////////////////////////////////////////////////
type EntityDescription struct {
	Id          string
	Mesh        string
	Material    string
	Textures    []string
	Position    matrix.Vec3
	Rotation    matrix.Vec3
	Scale       matrix.Vec3
	DataBinding []EntityDataBinding
	Children    []EntityDescription
}

type EntityDescriptionJson struct {
	Id          string
	Mesh        int
	Material    int                     `json:"Mat"`
	Textures    []int                   `json:"Tex,omitempty"`
	Position    matrix.Vec3             `json:"P"`
	Rotation    matrix.Vec3             `json:"R"`
	Scale       matrix.Vec3             `json:"S"`
	DataBinding []EntityDataBinding     `json:"Data,omitempty"`
	Children    []EntityDescriptionJson `json:"Kids,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

// //////////////////////////////////////////////////////////////////////////////
type EntityDataBinding struct {
	Name   string
	Fields map[string]any `json:",omitempty"`
}

// //////////////////////////////////////////////////////////////////////////////
func debugEnsureStructsMatch() {
	if build.Debug {
		ra := reflect.TypeFor[Stage]()
		rb := reflect.TypeFor[StageJson]()
		debug.Assert(ra.NumField() == rb.NumField()-3,
			"the Stage field has been modified but the matching StageSerialized was not updated")
		ea := reflect.TypeFor[EntityDescription]()
		eb := reflect.TypeFor[EntityDescriptionJson]()
		debug.Assert(ea.NumField() == eb.NumField(),
			"the EntityDescription field has been modified but the matching EntityDescriptionSerialized was not updated")
	}
}

func (s *Stage) ToMinimized() StageJson {
	debugEnsureStructsMatch()
	ss := StageJson{
		Id:       s.Id,
		Entities: make([]EntityDescriptionJson, len(s.Entities)),
	}
	meshMap := map[string]int{}
	matMap := map[string]int{}
	texMap := map[string]int{}
	// Add a blank string into each for the case that they are not assigned
	meshMap[""] = 0
	matMap[""] = 0
	texMap[""] = 0
	for i := range s.Entities {
		meshMap[s.Entities[i].Mesh] = 0
		matMap[s.Entities[i].Material] = 0
		for j := range s.Entities[i].Textures {
			texMap[s.Entities[i].Textures[j]] = 0
		}
	}
	for k := range meshMap {
		meshMap[k] = len(ss.Meshes)
		ss.Meshes = append(ss.Meshes, k)
	}
	for k := range matMap {
		matMap[k] = len(ss.Materials)
		ss.Materials = append(ss.Materials, k)
	}
	for k := range texMap {
		texMap[k] = len(ss.Textures)
		ss.Textures = append(ss.Textures, k)
	}
	var proc func(from *EntityDescription, to *EntityDescriptionJson)
	proc = func(from *EntityDescription, to *EntityDescriptionJson) {
		to.Id = from.Id
		to.Position = from.Position
		to.Rotation = from.Rotation
		to.Scale = from.Scale
		to.DataBinding = from.DataBinding
		to.Mesh = meshMap[from.Mesh]
		to.Material = matMap[from.Material]
		to.Textures = make([]int, len(from.Textures))
		for i := range from.Textures {
			to.Textures[i] = texMap[from.Textures[i]]
		}
		to.Children = make([]EntityDescriptionJson, len(to.Children))
		for i := range from.Children {
			proc(&from.Children[i], &to.Children[i])
		}
	}
	for i := range s.Entities {
		proc(&s.Entities[i], &ss.Entities[i])
	}
	return ss
}

func (s *Stage) FromMinimized(ss StageJson) {
	debugEnsureStructsMatch()
	s.Id = ss.Id
	s.Entities = make([]EntityDescription, len(ss.Entities))
	var proc func(from *EntityDescriptionJson, to *EntityDescription)
	proc = func(from *EntityDescriptionJson, to *EntityDescription) {
		to.Id = from.Id
		to.Position = from.Position
		to.Rotation = from.Rotation
		to.Scale = from.Scale
		to.DataBinding = from.DataBinding
		to.Mesh = ss.Meshes[from.Mesh]
		to.Material = ss.Materials[from.Material]
		to.Textures = make([]string, len(from.Textures))
		for i := range from.Textures {
			to.Textures[i] = ss.Textures[from.Textures[i]]
		}
		to.Children = make([]EntityDescription, len(from.Children))
		for i := range from.Children {
			proc(&from.Children[i], &to.Children[i])
		}
	}
	for i := range ss.Entities {
		proc(&ss.Entities[i], &s.Entities[i])
	}
}
