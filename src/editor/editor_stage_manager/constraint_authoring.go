/******************************************************************************/
/* constraint_authoring.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"log/slog"
	"reflect"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine_entity_data/engine_entity_data_physics"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type ConstraintChainKind int

const (
	ConstraintChainDistance ConstraintChainKind = iota
	ConstraintChainRope
	ConstraintChainHinge
)

type ConstraintChainAttachment struct {
	Entity *StageEntity
	Data   *entity_data_binding.EntityDataEntry
}

func (m *StageManager) ConnectSelectedAsDistanceChain() []ConstraintChainAttachment {
	return m.ConnectSelectedAsConstraintChain(ConstraintChainDistance)
}

func (m *StageManager) ConnectSelectedAsRope() []ConstraintChainAttachment {
	return m.ConnectSelectedAsConstraintChain(ConstraintChainRope)
}

func (m *StageManager) ConnectSelectedAsHingeChain() []ConstraintChainAttachment {
	return m.ConnectSelectedAsConstraintChain(ConstraintChainHinge)
}

func (m *StageManager) ConnectSelectedAsConstraintChain(kind ConstraintChainKind) []ConstraintChainAttachment {
	defer tracing.NewRegion("StageManager.ConnectSelectedAsConstraintChain").End()
	selection := m.SelectedHierarchyOrder()
	if len(selection) < 2 {
		slog.Warn("at least two selected entities are required to create a constraint chain")
		return nil
	}
	attachments := make([]ConstraintChainAttachment, 0, len(selection)-1)
	for i := 0; i < len(selection)-1; i++ {
		entry := newConstraintChainBinding(kind, selection[i], selection[i+1])
		if entry == nil {
			continue
		}
		selection[i].AttachDataBinding(entry)
		attachments = append(attachments, ConstraintChainAttachment{
			Entity: selection[i],
			Data:   entry,
		})
	}
	return attachments
}

func (m *StageManager) SelectedHierarchyOrder() []*StageEntity {
	defer tracing.NewRegion("StageManager.SelectedHierarchyOrder").End()
	if len(m.selected) == 0 {
		return nil
	}
	out := make([]*StageEntity, 0, len(m.selected))
	selected := make(map[*StageEntity]struct{}, len(m.selected))
	for _, e := range m.selected {
		if e != nil && !e.IsDeleted() {
			selected[e] = struct{}{}
		}
	}
	var walk func(*StageEntity)
	walk = func(e *StageEntity) {
		if e == nil || e.IsDeleted() {
			return
		}
		if _, ok := selected[e]; ok {
			out = append(out, e)
		}
		for _, child := range e.Children {
			walk(EntityToStageEntity(child))
		}
	}
	for _, e := range m.entities {
		if e.IsRoot() {
			walk(e)
		}
	}
	return out
}

func newConstraintChainBinding(kind ConstraintChainKind, source, target *StageEntity) *entity_data_binding.EntityDataEntry {
	targetId := engine.EntityId(target.StageData.Description.Id)
	switch kind {
	case ConstraintChainDistance:
		return bindingEntryForEntityData(&engine_entity_data_physics.DistanceJointEntityData{
			ConnectedEntityId: targetId,
			Stiffness:         1,
			Bias:              0.2,
			Correction:        0.8,
			Slop:              0.001,
			MaxCorrection:     0.5,
			Enabled:           true,
			AutoRestLength:    true,
		})
	case ConstraintChainRope:
		return bindingEntryForEntityData(&engine_entity_data_physics.RopeJointEntityData{
			ConnectedEntityId: targetId,
			Stiffness:         1,
			Bias:              0.2,
			Correction:        0.8,
			Slop:              0.001,
			MaxCorrection:     0.5,
			Enabled:           true,
			AutoMaxLength:     true,
		})
	case ConstraintChainHinge:
		anchor := source.Transform.WorldPosition().Add(target.Transform.WorldPosition()).Scale(0.5)
		return bindingEntryForEntityData(&engine_entity_data_physics.HingeJointEntityData{
			ConnectedEntityId: targetId,
			LocalAnchorA:      source.Transform.InverseWorldMatrix().TransformPoint(anchor),
			TargetAnchorB:     target.Transform.InverseWorldMatrix().TransformPoint(anchor),
			Stiffness:         1,
			Bias:              0.2,
			Correction:        0.8,
			Slop:              0.001,
			MaxCorrection:     0.5,
			Enabled:           true,
			HingeAxis:         matrix.Vec3Backward(),
		})
	default:
		slog.Warn("unknown constraint chain kind", "kind", kind)
		return nil
	}
}

func bindingEntryForEntityData(target any) *entity_data_binding.EntityDataEntry {
	entry := entity_data_binding.ToDataBinding("", target)
	key := qualifiedNameForBinding(target)
	entry.Name = key
	entry.Gen.RegisterKey = key
	return &entry
}

func qualifiedNameForBinding(target any) string {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return pod.QualifiedNameForLayout(reflect.New(t).Elem().Interface())
}
