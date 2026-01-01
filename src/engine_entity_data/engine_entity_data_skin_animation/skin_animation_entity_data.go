/******************************************************************************/
/* skin_animation_entity_data.go                                              */
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

package engine_entity_data_skin_animation

import (
	"kaiju/engine"
	"kaiju/engine_entity_data/content_id"
	"kaiju/framework"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/rendering/loaders/load_result"
	"log/slog"
	"strings"
	"weak"
)

const BindingKey = "kaiju.SkinAnimationEntityData"

func init() {
	engine.RegisterEntityData(BindingKey, SkinAnimationEntityData{})
}

type SkinAnimationEntityData struct {
	MeshId   content_id.Mesh
	AnimName string `options:"animations"`
}

type MeshSkinningAnimation struct {
	frame     int
	animIdx   int
	anims     []kaiju_mesh.KaijuMeshAnimation
	joints    []kaiju_mesh.KaijuMeshJoint
	updateId  engine.UpdateId
	entity    weak.Pointer[engine.Entity]
	skin      weak.Pointer[rendering.SkinnedShaderDataHeader]
	current   framework.SkinAnimation
	isPlaying bool
}

func (c SkinAnimationEntityData) Init(e *engine.Entity, host *engine.Host) {
	km, err := kaiju_mesh.ReadMesh(string(c.MeshId), host)
	if err != nil {
		slog.Error("failed to deserialize kaiju mesh", "id", c.MeshId, "error", err)
		return
	}
	anim := &MeshSkinningAnimation{
		anims:  km.Animations,
		joints: km.Joints,
		entity: weak.Make(e),
	}
	anim.SetAnimation(anim.anims[0].Name)
	wh := weak.Make(host)
	e.OnDestroy.Add(func() {
		h := wh.Value()
		if h != nil {
			h.Updater.RemoveUpdate(&anim.updateId)
		}
	})
	e.AddNamedData(BindingKey, anim)
	// The shader data hasn't been assigned yet, wait until the next frame to setup
	host.RunNextFrame(func() { anim.setup(host) })
}

func (a *MeshSkinningAnimation) SetAnimation(name string) {
	for i := range a.anims {
		if strings.EqualFold(a.anims[i].Name, name) {
			a.animIdx = i
		}
	}
	a.current = framework.NewSkinAnimation(a.anims[a.animIdx])
}

func (a *MeshSkinningAnimation) setup(host *engine.Host) {
	e := a.entity.Value()
	sd := e.ShaderData()
	skin := sd.SkinningHeader()
	if skin == nil {
		e.RemoveNamedData(BindingKey, a)
		slog.Error("failed to find skinning shader data on entity for MeshSkinningAnimation", "entity", e.Id())
		return
	}
	if !skin.HasBones() {
		ids := klib.ExtractFromSlice(a.joints, func(i int) int32 {
			return a.joints[i].Id
		})
		skin.CreateBones(ids)
		for i := range a.joints {
			j := &a.joints[i]
			bone := skin.BoneByIndex(i)
			bone.Id = j.Id
			bone.Skin = j.Skin
			bone.Transform.Initialize(host.WorkGroup())
		}
		for i := range a.joints {
			bone := skin.BoneByIndex(i)
			j := &a.joints[i]
			parent := skin.FindBone(j.Parent)
			if parent != nil {
				bone.Transform.SetParent(&parent.Transform)
			} else {
				bone.Transform.SetParent(&e.Transform)
			}
			bone.Transform.SetLocalPosition(j.Position)
			bone.Transform.SetRotation(j.Rotation)
			bone.Transform.SetScale(j.Scale)
		}
	}
	if !a.updateId.IsValid() {
		a.updateId = host.Updater.AddUpdate(a.update)
	}
}

func (a *MeshSkinningAnimation) update(deltaTime float64) {
	if !a.isPlaying {
		return
	}
	skin := a.skin.Value()
	if skin == nil {
		return
	}
	a.current.Update(deltaTime)
	frame := a.current.CurrentFrame()
	for i := range frame.Key.Bones {
		frame.Bone = &frame.Key.Bones[i]
		bone := skin.FindBone(int32(frame.Bone.NodeIndex))
		if bone == nil {
			continue
		}
		nextFrame, ok := a.current.FindNextFrameForBone(bone.Id, frame.Bone.PathType)
		if !ok {
			nextFrame = frame
			nextFrame.Bone = frame.Bone
		}
		data := a.current.Interpolate(frame, nextFrame)
		switch frame.Bone.PathType {
		case load_result.AnimPathTranslation:
			bone.Transform.SetLocalPosition(matrix.Vec3FromSlice(data[:]))
		case load_result.AnimPathRotation:
			bone.Transform.SetRotation(matrix.Quaternion(data).ToEuler())
		case load_result.AnimPathScale:
			bone.Transform.SetScale(matrix.Vec3FromSlice(data[:]))
		}
	}
}
