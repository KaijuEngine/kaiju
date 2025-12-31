/******************************************************************************/
/* skin_animation.go                                                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package framework

import (
	"kaiju/rendering/loaders/kaiju_mesh"
	"math"
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
		nextFrame: min(1, len(anim.Frames)-1),
	}
	s.absFrameTimes = make([]float64, len(s.Animation.Frames))
	for i := range s.Animation.Frames {
		s.absFrameTimes[i] = s.totalTime
		s.totalTime += float64(s.Animation.Frames[i].Time)
	}
	return s
}

func (a *SkinAnimation) FindNextFrameForBone(boneId int32, pathType kaiju_mesh.AnimationPathType) (SkinAnimationFrame, bool) {
	for i := a.frame + 1; i < len(a.Animation.Frames); i++ {
		for j := range a.Animation.Frames[i].Bones {
			if a.Animation.Frames[i].Bones[j].PathType == pathType {
				return SkinAnimationFrame{
					Key:     &a.Animation.Frames[i],
					Bone:    &a.Animation.Frames[i].Bones[j],
					AbsTime: a.absFrameTimes[i],
				}, true
			}
		}
	}
	for i := 0; i < a.frame; i++ {
		for j := range a.Animation.Frames[i].Bones {
			if a.Animation.Frames[i].Bones[j].PathType == pathType {
				return SkinAnimationFrame{
					Key:     &a.Animation.Frames[i],
					Bone:    &a.Animation.Frames[i].Bones[j],
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

func (a *SkinAnimation) Interpolate(from, to SkinAnimationFrame) [4]float32 {
	// if matrix.Vec4Approx(from.Bone.Data, to.Bone.Data) {
	return from.Bone.Data
	// }
	// t0 := from.AbsTime
	// t1 := to.AbsTime
	// if t1 < t0 {
	// 	t1 += t0
	// }
	// t := (a.time - t0) / (t1 - t0)
	// switch from.Bone.PathType {
	// case load_result.AnimPathRotation:
	// 	q0 := matrix.Quaternion(from.Bone.Data)
	// 	q1 := matrix.Quaternion(to.Bone.Data)
	// 	quat := matrix.QuaternionSlerp(q0, q1, float32(t))
	// 	return quat
	// case load_result.AnimPathTranslation:
	// 	fallthrough
	// case load_result.AnimPathScale:
	// 	p0 := matrix.Vec3FromSlice(from.Bone.Data[:])
	// 	p1 := matrix.Vec3FromSlice(to.Bone.Data[:])
	// 	out := matrix.Vec3Lerp(p0, p1, float32(t))
	// 	return [4]float32{out.X(), out.Y(), out.Z(), 1}
	// }
	// return from.Bone.Data
}
