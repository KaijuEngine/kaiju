/******************************************************************************/
/* skin_animation_entity_data_renderer.go                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package data_binding_renderer

import (
	"strings"
	"weak"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/engine_entity_data/engine_entity_data_skin_animation"
	"kaijuengine.com/framework"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
	"kaijuengine.com/rendering/loaders/load_result"
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
		meshAsset := kaiju_mesh.ParseMeshRef(meshId).Asset
		if !host.AssetDatabase().Exists(meshAsset) {
			return
		}
		km, err := kaiju_mesh.ReadMesh(meshId, host)
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
				bone.Transform.SetLocalPosition(j.Position)
				bone.Transform.SetRotation(j.Rotation)
				bone.Transform.SetScale(j.Scale)
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
