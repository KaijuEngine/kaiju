/******************************************************************************/
/* sample_stages_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package samples

import (
	"bytes"
	"encoding/json"
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestConstraintSampleStagesLoadArchiveAndRemainStable(t *testing.T) {
	tests := []struct {
		name            string
		stage           stages.Stage
		wantConstraints int
		assert          func(*testing.T, stages.LoadResult)
	}{
		{
			name:            SampleStageDistanceChain,
			stage:           DistanceChainSampleStage(),
			wantConstraints: 6,
			assert: func(t *testing.T, res stages.LoadResult) {
				center := res.EntitiesById["chain_03"].Transform.WorldPosition()
				if center.Length() > 0.7 {
					t.Fatalf("expected chain center to remain bounded, got %v", center)
				}
			},
		},
		{
			name:            SampleStageRope,
			stage:           RopeSampleStage(),
			wantConstraints: 4,
			assert: func(t *testing.T, res stages.LoadResult) {
				tail := res.EntitiesById["rope_04"].Transform.WorldPosition()
				if tail.Y() > -3.4 || tail.Y() < -4.8 {
					t.Fatalf("expected rope tail to stay near its span, got %v", tail)
				}
			},
		},
		{
			name:            SampleStageBridge,
			stage:           BridgeSampleStage(),
			wantConstraints: 6,
			assert: func(t *testing.T, res stages.LoadResult) {
				center := res.EntitiesById["bridge_03"].Transform.WorldPosition()
				if center.Y() > -0.02 || center.Y() < -1.25 {
					t.Fatalf("expected bridge center to sag without drifting away, got %v", center)
				}
			},
		},
		{
			name:            SampleStageHingePendulum,
			stage:           HingePendulumSampleStage(),
			wantConstraints: 1,
			assert: func(t *testing.T, res stages.LoadResult) {
				arm := res.EntitiesById["hinge_arm"].Transform.WorldPosition()
				if arm.Distance(matrix.Vec3Zero()) > 2.25 {
					t.Fatalf("expected hinge arm to stay anchored, got %v", arm)
				}
			},
		},
		{
			name:            SampleStageBodyWorld,
			stage:           BodyWorldAnchorSampleStage(),
			wantConstraints: 1,
			assert: func(t *testing.T, res stages.LoadResult) {
				body := res.EntitiesById["anchored_body"].Transform.WorldPosition()
				if body.Distance(matrix.Vec3Zero()) > 1.7 {
					t.Fatalf("expected body-world anchor to remain bounded, got %v", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSampleStageMinimizedJSON(t, tt.stage)
			assertSampleStageArchive(t, tt.stage)
			host := engine.NewHost("sample-stage-test", nil, nil)
			res := tt.stage.Load(host)
			if got := len(host.Physics().World().Constraints()); got != tt.wantConstraints {
				t.Fatalf("expected %d constraints, got %d", tt.wantConstraints, got)
			}
			stepSampleStage(t, host, 180)
			for _, entity := range res.Entities {
				pos := entity.Transform.WorldPosition()
				if !isFiniteVec3(pos) {
					t.Fatalf("expected finite entity position for %s, got %v", entity.Name(), pos)
				}
			}
			tt.assert(t, res)
		})
	}
}

func assertSampleStageMinimizedJSON(t *testing.T, stage stages.Stage) {
	t.Helper()
	data, err := json.Marshal(stage.ToMinimized())
	if err != nil {
		t.Fatalf("failed to marshal minimized stage: %v", err)
	}
	minimized := stages.StageJson{}
	if err := json.Unmarshal(data, &minimized); err != nil {
		t.Fatalf("failed to unmarshal minimized stage: %v", err)
	}
	loaded := stages.Stage{}
	loaded.FromMinimized(minimized)
	if len(loaded.Entities) != len(stage.Entities) {
		t.Fatalf("expected %d loaded editor entities, got %d", len(stage.Entities), len(loaded.Entities))
	}
	for i := range loaded.Entities {
		if len(loaded.Entities[i].DataBinding) == 0 {
			t.Fatalf("expected editor data bindings on entity %s", loaded.Entities[i].Id)
		}
	}
}

func assertSampleStageArchive(t *testing.T, stage stages.Stage) {
	t.Helper()
	archiveStage := cloneSampleStage(stage)
	for i := range archiveStage.Entities {
		stripEditorBindingsForArchive(&archiveStage.Entities[i])
	}
	buf := bytes.Buffer{}
	if err := pod.NewEncoder(&buf).Encode(archiveStage); err != nil {
		t.Fatalf("failed to encode sample stage archive: %v", err)
	}
	loaded, err := stages.ArchiveDeserializer(buf.Bytes())
	if err != nil {
		t.Fatalf("failed to deserialize sample stage archive: %v", err)
	}
	if len(loaded.Entities) != len(archiveStage.Entities) {
		t.Fatalf("expected %d archived entities, got %d", len(archiveStage.Entities), len(loaded.Entities))
	}
	for i := range loaded.Entities {
		if len(loaded.Entities[i].RawDataBinding) == 0 {
			t.Fatalf("expected raw archive data bindings on entity %s", loaded.Entities[i].Id)
		}
	}
}

func cloneSampleStage(stage stages.Stage) stages.Stage {
	stage.Entities = cloneSampleEntities(stage.Entities)
	return stage
}

func cloneSampleEntities(entities []stages.EntityDescription) []stages.EntityDescription {
	out := make([]stages.EntityDescription, len(entities))
	copy(out, entities)
	for i := range out {
		out[i].DataBinding = append([]stages.EntityDataBinding(nil), entities[i].DataBinding...)
		out[i].RawDataBinding = append([]any(nil), entities[i].RawDataBinding...)
		out[i].Children = cloneSampleEntities(entities[i].Children)
	}
	return out
}

func stripEditorBindingsForArchive(desc *stages.EntityDescription) {
	desc.DataBinding = nil
	for i := range desc.Children {
		stripEditorBindingsForArchive(&desc.Children[i])
	}
}

func stepSampleStage(t *testing.T, host *engine.Host, steps int) {
	t.Helper()
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()
	host.Physics().SetMaxSubSteps(1)
	for range steps {
		host.Physics().Update(&workGroup, &threads, host.Physics().FixedTimeStep())
	}
}

func isFiniteVec3(v matrix.Vec3) bool {
	return !v.IsNaN() && !v.IsInf(0)
}
