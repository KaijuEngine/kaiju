package editor_stage_manager

import (
	"encoding/json"
	"kaiju/engine"
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/stages"
	"os"
	"weak"

	"github.com/KaijuEngine/uuid"
)

type StageEntity struct {
	engine.Entity
	StageData StageEntityEditorData
}

// StageManager represents the current stage in the editor. It contains all of
// the entities on the stage.
type StageManager struct {
	host     *engine.Host
	entities []*StageEntity
	selected []*StageEntity
}

// StageEntityEditorData is the structure holding all the uniquely identifiable
// and linking data about the entity on this stage. That will include things
// like content linkage, data bindings, etc.
type StageEntityEditorData struct {
	Bvh         *collision.BVH
	ShaderData  rendering.DrawInstance
	Description stages.EntityDescription
}

func (m *StageManager) Initialize(host *engine.Host) { m.host = host }

// List will return all of the internally held entities for the stage
func (m *StageManager) List() []*StageEntity { return m.entities }

func (m *StageManager) Selection() []*StageEntity { return m.selected }

// AddEntity will create a new entity for the stage. This entity will have a
// #StageEntityData automatically added to it as named data named "stage".
func (m *StageManager) AddEntity(point matrix.Vec3) *StageEntity {
	e := &StageEntity{}
	e.Init(m.host.WorkGroup())
	e.StageData.Description.Id = uuid.NewString()
	m.host.AddEntity(&e.Entity)
	e.Transform.SetPosition(point)
	m.entities = append(m.entities, e)
	e.AddNamedData("stage", e.StageData)
	wm := weak.Make(m)
	we := weak.Make(e)
	e.OnDestroy.Add(func() {
		sm := wm.Value()
		if sm == nil {
			return
		}
		if e.StageData.ShaderData != nil {
			e.StageData.ShaderData.Destroy()
		}
		se := we.Value()
		for i := range sm.entities {
			if sm.entities[i] == se {
				sm.entities = klib.RemoveUnordered(sm.entities, i)
				return
			}
		}
	})
	return e
}

// Clear will destroy all entities that are managed by this stage manager.
func (m *StageManager) Clear() {
	for i := range m.entities {
		m.entities[i].Destroy()
	}
}

func (m *StageManager) SaveStage() {
	s := stages.Stage{
		Id: uuid.NewString(),
	}
	rootCount := 0
	for i := range m.entities {
		if m.entities[i].IsRoot() {
			rootCount++
		}
	}
	s.Entities = make([]stages.EntityDescription, 0, rootCount)
	var readEntity func(parent *StageEntity)
	readEntity = func(parent *StageEntity) {
		desc := &parent.StageData.Description
		desc.Transform.Position = parent.Transform.Position()
		desc.Transform.Rotation = parent.Transform.Rotation()
		desc.Transform.Scale = parent.Transform.Scale()
		for i := range m.entities {
			if m.entities[i].Parent == &parent.Entity {
				desc.Children = append(desc.Children, m.entities[i].StageData.Description)
				readEntity(m.entities[i])
			}
		}
	}
	for i := range m.entities {
		if m.entities[i].IsRoot() {
			readEntity(m.entities[i])
			s.Entities = append(s.Entities, m.entities[i].StageData.Description)
		}
	}
	// TODO:  The code below is test code
	f, err := os.Create("tmp_stage.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(s)
}
