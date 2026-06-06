/******************************************************************************/
/* skin_animation_entity_data.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_skin_animation

import (
	"log/slog"
	"strings"
	"weak"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/framework"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
	"kaijuengine.com/rendering/loaders/load_result"
)

var bindingKey = ""

func init() {
	engine.RegisterEntityData(SkinAnimationEntityData{})
}

func BindingKey() string {
	if bindingKey == "" {
		bindingKey = pod.QualifiedNameForLayout(SkinAnimationEntityData{})
	}
	return bindingKey
}

type SkinAnimationEntityData struct {
	MeshId   content_id.Mesh
	AnimName string `options:"animations"`
}

type MeshSkinningAnimation struct {
	frame          int
	animIdx        int
	anims          []kaiju_mesh.KaijuMeshAnimation
	joints         []kaiju_mesh.KaijuMeshJoint
	updateId       engine.UpdateId
	entity         weak.Pointer[engine.Entity]
	skin           weak.Pointer[rendering.SkinnedShaderDataHeader]
	shaderDataBase weak.Pointer[rendering.ShaderDataBase]
	current        framework.SkinAnimation
	isPlaying      bool
}

func (c SkinAnimationEntityData) Init(e *engine.Entity, host *engine.Host) {
	km, err := kaiju_mesh.ReadMesh(string(c.MeshId), host)
	if err != nil {
		slog.Error("failed to deserialize kaiju mesh", "id", c.MeshId, "error", err)
		return
	}
	sd := e.ShaderData()
	anim := &MeshSkinningAnimation{
		anims:          km.Animations,
		joints:         km.Joints,
		entity:         weak.Make(e),
		skin:           weak.Make(sd.SkinningHeader()),
		shaderDataBase: weak.Make(sd.Base()),
	}
	anim.SetAnimation(c.AnimName)
	wh := weak.Make(host)
	e.OnDestroy.Add(func() {
		h := wh.Value()
		if h != nil {
			h.Updater.RemoveUpdate(&anim.updateId)
		}
	})
	e.AddNamedData(bindingKey, anim)
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
	a.isPlaying = true
}

func (a *MeshSkinningAnimation) setup(host *engine.Host) {
	e := a.entity.Value()
	sd := e.ShaderData()
	skin := sd.SkinningHeader()
	if skin == nil {
		e.RemoveNamedData(bindingKey, a)
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
			bone.Transform.SetLocalPosition(j.Position)
			bone.Transform.SetRotation(j.Rotation)
			bone.Transform.SetScale(j.Scale)
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
	sd := a.shaderDataBase.Value()
	skin := a.skin.Value()
	if skin == nil || (sd != nil && !sd.IsInView()) {
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
