/******************************************************************************/
/* shader_data_pbr_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"testing"

	"kaijuengine.com/rendering"
)

func TestPBRSelectsPreexistingLightsOnFirstUpdate(t *testing.T) {
	data := Create("pbr").(*ShaderDataPBR)
	light := rendering.NewLight(&rendering.GPUDevice{}, nil, nil, rendering.LightTypeDirectional)

	data.SelectLights(rendering.LightsForRender{
		Lights: []rendering.Light{light},
		// This is deliberately false: stage reload can add a drawing after the
		// light collection's change flag was consumed by an earlier frame.
		HasChanges: false,
	})

	if data.LightIds[0] != 0 {
		t.Fatalf("first PBR light id = %d, want 0", data.LightIds[0])
	}
}

func TestPBRRefreshesLightsWhenCollectionChanges(t *testing.T) {
	data := Create("pbr").(*ShaderDataPBR)
	light := rendering.NewLight(&rendering.GPUDevice{}, nil, nil, rendering.LightTypeDirectional)
	data.SelectLights(rendering.LightsForRender{Lights: []rendering.Light{light}})
	data.SelectLights(rendering.LightsForRender{HasChanges: true})

	for i, id := range data.LightIds {
		if id != -1 {
			t.Fatalf("light id %d after removal = %d, want -1", i, id)
		}
	}
}
