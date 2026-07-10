/******************************************************************************/
/* skin_animation.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package framework

import (
	"math"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
	"kaijuengine.com/rendering/loaders/load_result"
)

type SkinAnimation struct {
	Animation     kaiju_mesh.KaijuMeshAnimation
	frame         int
	nextFrame     int
	time          float64
	totalTime     float64
	absFrameTimes []float64
}

type SkinAnimationFrame struct {
	Key     *kaiju_mesh.AnimKeyFrame
	Bone    *kaiju_mesh.AnimBone
	AbsTime float64
}

func NewSkinAnimation(anim kaiju_mesh.KaijuMeshAnimation) SkinAnimation {
	s := SkinAnimation{
		Animation: anim,
		frame:     -1,
	}
	s.absFrameTimes = make([]float64, len(s.Animation.Frames))
	for i := range s.Animation.Frames {
		s.absFrameTimes[i] = s.totalTime
		s.totalTime += float64(s.Animation.Frames[i].Time)
	}
	return s
}

func (a *SkinAnimation) IsValid() bool { return len(a.Animation.Frames) > 0 }

func (a *SkinAnimation) FindNextFrameForBone(boneId int32, pathType kaiju_mesh.AnimationPathType) (SkinAnimationFrame, bool) {
	for i := a.frame + 1; i < len(a.Animation.Frames); i++ {
		for j := range a.Animation.Frames[i].Bones {
			bone := &a.Animation.Frames[i].Bones[j]
			if bone.NodeIndex == int(boneId) && bone.PathType == pathType {
				return SkinAnimationFrame{
					Key:     &a.Animation.Frames[i],
					Bone:    bone,
					AbsTime: a.absFrameTimes[i],
				}, true
			}
		}
	}
	for i := 0; i < a.frame; i++ {
		for j := range a.Animation.Frames[i].Bones {
			bone := &a.Animation.Frames[i].Bones[j]
			if bone.NodeIndex == int(boneId) && bone.PathType == pathType {
				return SkinAnimationFrame{
					Key:     &a.Animation.Frames[i],
					Bone:    bone,
					AbsTime: a.absFrameTimes[i],
				}, true
			}
		}
	}
	return SkinAnimationFrame{}, false
}

func (a *SkinAnimation) CurrentFrame() SkinAnimationFrame {
	return SkinAnimationFrame{
		Key:     &a.Animation.Frames[a.frame],
		AbsTime: a.absFrameTimes[a.frame],
	}
}

func (a *SkinAnimation) NextFrame() SkinAnimationFrame {
	return SkinAnimationFrame{
		Key:     &a.Animation.Frames[a.nextFrame],
		AbsTime: a.absFrameTimes[a.nextFrame],
	}
}

func (a *SkinAnimation) Update(deltaTime float64) {
	if len(a.Animation.Frames) <= 1 {
		return
	}
	// Setting first frame is crutial, it shouldn't be skipped by large delta time
	if a.frame < 0 {
		a.frame = 0
		a.nextFrame = min(a.frame+1, len(a.Animation.Frames)-1)
	} else {
		a.time += deltaTime
		nextTime := a.absFrameTimes[a.nextFrame]
		for a.time >= nextTime {
			a.frame++
			if a.time > a.totalTime {
				a.time = math.Mod(a.time, a.totalTime)
				a.frame = 0
			}
			a.nextFrame = min(a.frame+1, len(a.Animation.Frames)-1)
			nextTime = a.absFrameTimes[a.nextFrame]
		}
	}
}

func (a *SkinAnimation) Interpolate(from, to SkinAnimationFrame) [4]matrix.Float {
	if matrix.Vec4Approx(from.Bone.Data, to.Bone.Data) {
		return from.Bone.Data
	}
	t0 := from.AbsTime
	t1 := to.AbsTime
	if t1 < t0 {
		t1 += a.totalTime
	}
	t := (a.time - t0) / (t1 - t0)
	switch from.Bone.PathType {
	case load_result.AnimPathRotation:
		q0 := matrix.Quaternion(from.Bone.Data)
		q1 := matrix.Quaternion(to.Bone.Data)
		quat := matrix.QuaternionSlerp(q0, q1, matrix.Float(t))
		return quat
	case load_result.AnimPathTranslation:
		fallthrough
	case load_result.AnimPathScale:
		p0 := matrix.Vec3FromSlice(from.Bone.Data[:])
		p1 := matrix.Vec3FromSlice(to.Bone.Data[:])
		out := matrix.Vec3Lerp(p0, p1, matrix.Float(t))
		return [4]matrix.Float{out.X(), out.Y(), out.Z(), 1}
	}
	return from.Bone.Data
}
