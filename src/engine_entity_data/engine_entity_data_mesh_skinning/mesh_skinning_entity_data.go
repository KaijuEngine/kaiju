package engine_entity_data_mesh_skinning

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

const BindingKey = "kaiju.MeshSkinningEntityData"

func init() {
	engine.RegisterEntityData(BindingKey, MeshSkinningEntityData{})
}

type MeshSkinningEntityData struct {
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

func (c MeshSkinningEntityData) Init(e *engine.Entity, host *engine.Host) {
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
