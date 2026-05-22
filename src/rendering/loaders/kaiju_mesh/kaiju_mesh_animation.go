/******************************************************************************/
/* kaiju_mesh_animation.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package kaiju_mesh

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/loaders/load_result"
)

type KaijuMeshJoint struct {
	Id       int32
	Parent   int32
	Skin     matrix.Mat4
	Position matrix.Vec3
	Rotation matrix.Vec3
	Scale    matrix.Vec3
}

type KaijuMeshAnimation struct {
	Name   string
	Frames []AnimKeyFrame
}

type AnimKeyFrame struct {
	Bones []AnimBone
	Time  float32
}

type AnimBone struct {
	NodeIndex     int
	PathType      AnimationPathType
	Interpolation AnimationInterpolation
	// Could be Vec3 or Quaternion, doing this because Go doesn't have a union
	Data [4]matrix.Float
}

func (j *KaijuMeshJoint) fromLoadResult(res *load_result.Result, r *load_result.Joint) {
	j.Id = r.Id
	j.Skin = r.Skin
	n := &res.Nodes[j.Id]
	j.Parent = int32(n.Parent)
	j.Position = n.Position
	j.Rotation = n.Rotation.ToEuler()
	j.Scale = n.Scale
}

func (a *KaijuMeshAnimation) fromLoadResult(r *load_result.Animation) {
	a.Name = r.Name
	a.Frames = make([]AnimKeyFrame, len(r.Frames))
	for i := range r.Frames {
		a.Frames[i].fromLoadResult(&r.Frames[i])
	}
}

func (f *AnimKeyFrame) fromLoadResult(r *load_result.AnimKeyFrame) {
	f.Time = r.Time
	f.Bones = make([]AnimBone, len(r.Bones))
	for i := range r.Bones {
		f.Bones[i].fromLoadResult(&r.Bones[i])
	}
}

func (b *AnimBone) fromLoadResult(r *load_result.AnimBone) {
	b.Data = r.Data
	b.Interpolation = r.Interpolation
	b.NodeIndex = r.NodeIndex
	b.PathType = r.PathType
}
