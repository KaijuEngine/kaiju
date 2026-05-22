/******************************************************************************/
/* terrain_entity_data.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_terrain

import (
	"log/slog"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/engine_entity_data/content_id"
)

var bindingKey = ""

func init() {
	engine.RegisterEntityData(TerrainEntityData{})
}

func BindingKey() string {
	if bindingKey == "" {
		bindingKey = pod.QualifiedNameForLayout(TerrainEntityData{})
	}
	return bindingKey
}

type TerrainEntityData struct {
	Id content_id.Terrain `visible:"false"`
}

func (d TerrainEntityData) Init(e *engine.Entity, host *engine.Host) {
	model, err := terrain.LoadForEntity(host, string(d.Id), e)
	if err != nil {
		slog.Error("failed to load terrain", "id", d.Id, "error", err)
		return
	}
	e.AddNamedData("Terrain", model)
	e.OnDestroy.Add(func() {
		model.Destroy(nil)
		e.RemoveNamedData("Terrain", model)
	})
}

func (d TerrainEntityData) EntityDataInitPhase() engine.EntityDataPhase {
	return engine.EntityDataPhasePhysicsBody - 1
}
