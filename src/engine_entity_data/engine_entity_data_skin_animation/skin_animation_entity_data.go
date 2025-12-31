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
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/rendering/loaders/load_result"
	"log/slog"
	"weak"
)

const BindingKey = "kaiju.SkinAnimationEntityData"

func init() {
	engine.RegisterEntityData(BindingKey, SkinAnimationEntityData{})
}

type SkinAnimationEntityData struct {
	MeshId content_id.Mesh
}

type MeshSkinningAnimation struct {
	frame     int
	animIdx   int
	animTime  float64
	anims     []kaiju_mesh.KaijuMeshAnimation
	joints    []kaiju_mesh.KaijuMeshJoint
	updateId  engine.UpdateId
	entity    weak.Pointer[engine.Entity]
	skin      weak.Pointer[rendering.SkinnedShaderDataHeader]
	isPlaying bool
}

func (c SkinAnimationEntityData) Init(e *engine.Entity, host *engine.Host) {
	data, err := host.AssetDatabase().Read(string(c.MeshId))
	if err != nil {
		slog.Error("failed to read the mesh", "id", c.MeshId, "error", err)
		return
	}
	km, err := kaiju_mesh.Deserialize(data)
	if err != nil {
		slog.Error("failed to deserialize kaiju mesh", "id", c.MeshId, "error", err)
		return
	}
	anim := &MeshSkinningAnimation{
		anims:  km.Animations,
		joints: km.Joints,
		entity: weak.Make(e),
	}
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

func (a *MeshSkinningAnimation) setup(host *engine.Host) {
	e := a.entity.Value()
	sd := e.ShaderData()
	header := sd.SkinningHeader()
	if header == nil {
		e.RemoveNamedData(BindingKey, a)
		slog.Error("failed to find skinning shader data on entity for MeshSkinningAnimation", "entity", e.Id())
		return
	}
	a.updateId = host.Updater.AddUpdate(a.update)
}

func (a *MeshSkinningAnimation) update(deltaTime float64) {
	if !a.isPlaying {
		return
	}
	skin := a.skin.Value()
	if skin == nil {
		return
	}
	a.animTime += deltaTime
	if a.animTime >= float64(a.anims[a.animIdx].Frames[a.frame].Time) {
		a.frame++
		a.animTime = 0
		if a.frame >= len(a.anims[a.animIdx].Frames) {
			a.frame = 0
		}
	}
	for i := range a.anims[a.animIdx].Frames[a.frame].Bones {
		b := &a.anims[a.animIdx].Frames[a.frame].Bones[i]
		bone := skin.FindBone(int32(b.NodeIndex))
		if bone == nil {
			continue
		}
		switch b.PathType {
		case load_result.AnimPathTranslation:
			bone.Transform.SetPosition(matrix.Vec3FromSlice(b.Data[:]))
		case load_result.AnimPathRotation:
			bone.Transform.SetRotation(matrix.Quaternion(b.Data).ToEuler())
		case load_result.AnimPathScale:
			bone.Transform.SetScale(matrix.Vec3FromSlice(b.Data[:]))
		}
	}
}
