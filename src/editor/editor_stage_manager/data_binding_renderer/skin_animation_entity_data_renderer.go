/******************************************************************************/
/* skin_animation_entity_data_renderer.go                                     */
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

package data_binding_renderer

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine_entity_data/content_id"
	"kaiju/engine_entity_data/engine_entity_data_skin_animation"
	"kaiju/framework"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
	"kaiju/rendering/loaders/load_result"
	"strings"
	"weak"
)

type skinAnimationGizmo struct {
	skin       weak.Pointer[rendering.SkinnedShaderDataHeader]
	anim       framework.SkinAnimation
	animations []kaiju_mesh.KaijuMeshAnimation
	meshId     string
	animName   string
}

type SkinAnimationEntityDataRenderer struct {
	Skins    map[*editor_stage_manager.StageEntity]*skinAnimationGizmo
	updateId engine.UpdateId
}

func init() {
	AddRenderer(engine_entity_data_skin_animation.BindingKey(), &SkinAnimationEntityDataRenderer{
		Skins: make(map[*editor_stage_manager.StageEntity]*skinAnimationGizmo),
	})
}

func (c *SkinAnimationEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("SkinAnimationEntityDataRenderer.Attached").End()
	if !c.updateId.IsValid() {
		c.updateId = host.Updater.AddUpdate(c.update)
	}
	skin := target.StageData.ShaderData.SkinningHeader()
	g := &skinAnimationGizmo{
		skin: weak.Make(skin),
	}
	c.Skins[target] = g
	target.OnDestroy.Add(func() {
		c.Detatched(host, manager, target, data)
	})
	if skin != nil {
		c.bindSkin(host, target, data)
		c.Update(host, target, data)
	}
}

func (c *SkinAnimationEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("SkinAnimationEntityDataRenderer.Detatched").End()
	delete(c.Skins, target)
}

func (c *SkinAnimationEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	// defer tracing.NewRegion("SkinAnimationEntityDataRenderer.Show").End()
}

func (c *SkinAnimationEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	// defer tracing.NewRegion("SkinAnimationEntityDataRenderer.Hide").End()
}

func (c *SkinAnimationEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if g, ok := c.Skins[target]; ok {
		meshId := string(data.FieldValueByName("MeshId").(content_id.Mesh))
		name := data.FieldValueByName("AnimName").(string)
		skin := target.StageData.ShaderData.SkinningHeader()
		if skin != nil {
			if g.meshId != meshId {
				c.bindSkin(host, target, data)
			}
			if !strings.EqualFold(g.animName, name) {
				for i := range g.animations {
					if strings.EqualFold(g.animations[i].Name, name) {
						g.anim = framework.NewSkinAnimation(g.animations[i])
						break
					}
				}
			}
		}
		g.meshId = meshId
		g.animName = name
	}
}

func (c *SkinAnimationEntityDataRenderer) bindSkin(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("SkinAnimationEntityDataRenderer.bindSkin").End()
	skin := target.StageData.ShaderData.SkinningHeader()
	if skin == nil {
		return
	}
	meshId := string(data.FieldValueByName("MeshId").(content_id.Mesh))
	g := c.Skins[target]
	if len(g.animations) == 0 {
		if !host.AssetDatabase().Exists(meshId) {
			return
		}
		meshData, err := host.AssetDatabase().Read(meshId)
		if err != nil {
			return
		}
		km, err := kaiju_mesh.Deserialize(meshData)
		if err != nil {
			return
		}
		ids := klib.ExtractFromSlice(km.Joints, func(i int) int32 {
			return km.Joints[i].Id
		})
		if !skin.HasBones() {
			skin.CreateBones(ids)
			for i := range km.Joints {
				j := &km.Joints[i]
				bone := skin.BoneByIndex(i)
				bone.Id = j.Id
				bone.Skin = j.Skin
				bone.Transform.Initialize(host.WorkGroup())
			}
			for i := range km.Joints {
				bone := skin.BoneByIndex(i)
				j := &km.Joints[i]
				parent := skin.FindBone(j.Parent)
				if parent != nil {
					bone.Transform.SetParent(&parent.Transform)
				} else {
					bone.Transform.SetParent(&target.Transform)
				}
				bone.Transform.SetLocalPosition(j.Position)
				bone.Transform.SetRotation(j.Rotation)
				bone.Transform.SetScale(j.Scale)
			}
		}
		g.animations = km.Animations
	}
	if len(g.animations) > 0 {
		if g.animName == "" {
			g.anim = framework.SkinAnimation{}
		} else {
			for i := range g.animations {
				if strings.EqualFold(g.animations[i].Name, g.animName) {
					g.anim = framework.NewSkinAnimation(g.animations[i])
					break
				}
			}
		}
	}
}

func (c *SkinAnimationEntityDataRenderer) update(deltaTime float64) {
	defer tracing.NewRegion("SkinAnimationEntityDataRenderer.update").End()
	for k, v := range c.Skins {
		if !k.IsActive() || len(v.animations) == 0 {
			continue
		}
		skin := v.skin.Value()
		if skin == nil || !v.anim.IsValid() || !k.StageData.ShaderData.IsInView() {
			continue
		}
		v.anim.Update(deltaTime)
		frame := v.anim.CurrentFrame()
		for i := range frame.Key.Bones {
			frame.Bone = &frame.Key.Bones[i]
			bone := skin.FindBone(int32(frame.Bone.NodeIndex))
			if bone == nil {
				continue
			}
			nextFrame, ok := v.anim.FindNextFrameForBone(bone.Id, frame.Bone.PathType)
			if !ok {
				nextFrame = frame
				nextFrame.Bone = frame.Bone
			}
			data := v.anim.Interpolate(frame, nextFrame)
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
}
