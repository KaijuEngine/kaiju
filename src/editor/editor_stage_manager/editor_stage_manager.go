package editor_stage_manager

import (
	"kaiju/engine"
	"kaiju/engine/collision"
	"kaiju/klib"
	"kaiju/matrix"
	"weak"
)

// StageManager represents the current stage in the editor. It contains all of
// the entities on the stage.
type StageManager struct {
	host     *engine.Host
	entities []*engine.Entity
}

// StageEntityData is the structure holding all the uniquely identifiable and
// linking data about the entity on this stage. That will include things like
// content linkage, data bindings, etc.
type StageEntityData struct {
	Bvh       *collision.BVH
	Rendering struct {
		MeshId     string
		TextureIds []string
	}
}

func (m *StageManager) Initialize(host *engine.Host) { m.host = host }

// List will return all of the internally held entities for the stage
func (m *StageManager) List() []*engine.Entity { return m.entities }

// AddEntity will create a new entity for the stage. This entity will have a
// #StageEntityData automatically added to it as named data named "stage".
func (m *StageManager) AddEntity(point matrix.Vec3) (*engine.Entity, *StageEntityData) {
	e := m.host.NewEntity()
	e.Transform.SetPosition(point)
	m.entities = append(m.entities, e)
	sd := &StageEntityData{}
	e.AddNamedData("stage", sd)
	wm := weak.Make(m)
	we := weak.Make(e)
	e.OnDestroy.Add(func() {
		sm := wm.Value()
		if sm == nil {
			return
		}
		se := we.Value()
		for i := range sm.entities {
			if sm.entities[i] == se {
				sm.entities = klib.RemoveUnordered(sm.entities, i)
				return
			}
		}
	})
	return e, sd
}

// Clear will destroy all entities that are managed by this stage manager.
func (m *StageManager) Clear() {
	for i := range m.entities {
		m.entities[i].Destroy()
	}
}
