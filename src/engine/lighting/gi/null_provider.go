/******************************************************************************/
/* null_provider.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import (
	"fmt"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type NullProvider struct{}

func (*NullProvider) ID() string                                               { return ProviderNull }
func (*NullProvider) Supports(Capabilities) bool                               { return true }
func (*NullProvider) Initialize(ProviderContext) error                         { return nil }
func (*NullProvider) Configure(Settings) error                                 { return nil }
func (*NullProvider) SyncScene(SceneDelta) error                               { return nil }
func (*NullProvider) AddUpdatePasses(*rendering.FrameGraph, FrameInputs) error { return nil }
func (*NullProvider) ProbeField(ViewID) ProbeFieldBinding                      { return ProbeFieldBinding{} }
func (*NullProvider) ShaderData(matrix.Vec3, float32) rendering.GlobalIlluminationForRender {
	return rendering.GlobalIlluminationForRender{}
}
func (*NullProvider) Invalidate(Invalidation)  {}
func (*NullProvider) ResetHistory(ViewID)      {}
func (*NullProvider) SetScenario(string) error { return nil }
func (*NullProvider) DebugViews() []DebugView  { return nil }
func (*NullProvider) Shutdown()                {}
func (*NullProvider) Stats() Stats {
	return Stats{Provider: ProviderNull, Converged: true}
}

func (*NullProvider) AddResolvePasses(graph *rendering.FrameGraph, inputs FrameInputs) (Outputs, error) {
	name := fmt.Sprintf("gi.null.diffuse.%d", inputs.View)
	resource, err := graph.AddResource(rendering.FrameGraphResourceDescription{
		Name: name,
		Kind: rendering.FrameGraphResourceImage,
	})
	if err != nil {
		return Outputs{}, err
	}
	_, err = graph.AddPass(rendering.FrameGraphPassDescription{
		Name:  fmt.Sprintf("GI Null Resolve [%d]", inputs.View),
		Queue: rendering.FrameGraphQueueCompute,
		Uses: []rendering.FrameGraphResourceUse{{
			Resource: resource,
			Access:   rendering.FrameGraphAccessWrite,
		}},
		Execute: func(context *rendering.FrameGraphExecutionContext) error {
			context.Values[name] = [4]float32{}
			return nil
		},
	})
	if err != nil {
		return Outputs{}, err
	}
	return Outputs{DiffuseIrradiance: resource, Valid: true}, nil
}
