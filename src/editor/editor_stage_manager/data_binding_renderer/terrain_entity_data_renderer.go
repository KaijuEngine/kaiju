/******************************************************************************/
/* terrain_entity_data_renderer.go                                            */
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
	"log/slog"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/terrain"
	"kaijuengine.com/engine_entity_data/content_id"
	"kaijuengine.com/engine_entity_data/engine_entity_data_terrain"
	"kaijuengine.com/platform/profiler/tracing"
)

type terrainGizmo struct {
	id      string
	terrain *terrain.Terrain
}

type TerrainEntityDataRenderer struct {
	Terrains map[*editor_stage_manager.StageEntity]*terrainGizmo
}

func init() {
	AddRenderer(engine_entity_data_terrain.BindingKey(), &TerrainEntityDataRenderer{
		Terrains: make(map[*editor_stage_manager.StageEntity]*terrainGizmo),
	})
}

func (r *TerrainEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("TerrainEntityDataRenderer.Attached").End()
	r.Terrains[target] = &terrainGizmo{}
	target.OnDestroy.Add(func() {
		r.Detatched(host, manager, target, data)
	})
	target.OnActivate.Add(func() {
		if g, ok := r.Terrains[target]; ok && g.terrain != nil {
			for i := range g.terrain.ShaderData {
				g.terrain.ShaderData[i].Activate()
			}
		}
	})
	target.OnDeactivate.Add(func() {
		if g, ok := r.Terrains[target]; ok && g.terrain != nil {
			for i := range g.terrain.ShaderData {
				g.terrain.ShaderData[i].Deactivate()
			}
		}
	})
	r.Update(host, target, data)
}

func (r *TerrainEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("TerrainEntityDataRenderer.Detatched").End()
	if g, ok := r.Terrains[target]; ok {
		if g.terrain != nil {
			g.terrain.Destroy(nil)
		}
		delete(r.Terrains, target)
	}
}

func (r *TerrainEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("TerrainEntityDataRenderer.Show").End()
	// Terrain visuals are persistent and do not depend on selection state.
	// Update will only reload if the terrain ID has changed.
	r.Update(host, target, data)
}

func (r *TerrainEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("TerrainEntityDataRenderer.Hide").End()
	// Do not destroy the terrain on deselect - it should show at all times in the stage.
	// Hide is only for selection-based gizmos/overlays (none for terrain itself).
	// The terrain is deactivated only via entity OnDeactivate if the entity is disabled.
	if g, ok := r.Terrains[target]; ok && g.terrain != nil {
		g.terrain.ClearBrushPreview()
	}
}

func (r *TerrainEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("TerrainEntityDataRenderer.Update").End()
	g, ok := r.Terrains[target]
	if !ok {
		g = &terrainGizmo{}
		r.Terrains[target] = g
	}
	id, ok := data.FieldValueByName("Id").(content_id.Terrain)
	if !ok {
		slog.Error("terrain id failure", "id", id)
		return
	}
	sid := string(id)
	if sid == "" {
		return
	}
	if g.id == sid && g.terrain != nil {
		return // no change
	}
	if g.terrain != nil {
		g.terrain.Destroy(nil)
		g.terrain = nil
	}
	model, err := terrain.LoadForEntity(host, sid, &target.Entity)
	if err != nil {
		slog.Error("invalid terrain id specified", "id", id, "error", err)
		return
	}
	g.id = sid
	g.terrain = model
}
