/******************************************************************************/
/* editor_power_test.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"testing"

	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/platform/power"
)

func TestEditorEffectiveRefreshRateForPowerStatus(t *testing.T) {
	tests := []struct {
		name     string
		settings editor_settings.Settings
		status   power.Status
		want     int32
	}{
		{
			name: "AC uses default refresh rate",
			settings: editor_settings.Settings{
				RefreshRate:           144,
				UseBatteryRefreshRate: true,
				BatteryRefreshRate:    30,
			},
			status: power.Status{Source: power.SourceAC, HasBattery: true},
			want:   144,
		},
		{
			name: "battery uses battery refresh rate",
			settings: editor_settings.Settings{
				RefreshRate:           144,
				UseBatteryRefreshRate: true,
				BatteryRefreshRate:    30,
			},
			status: power.Status{Source: power.SourceBattery, HasBattery: true},
			want:   30,
		},
		{
			name: "unknown uses default refresh rate",
			settings: editor_settings.Settings{
				RefreshRate:           144,
				UseBatteryRefreshRate: true,
				BatteryRefreshRate:    30,
			},
			status: power.Status{Source: power.SourceUnknown, HasBattery: true},
			want:   144,
		},
		{
			name: "no battery uses default refresh rate",
			settings: editor_settings.Settings{
				RefreshRate:           144,
				UseBatteryRefreshRate: true,
				BatteryRefreshRate:    30,
			},
			status: power.Status{Source: power.SourceAC, HasBattery: false},
			want:   144,
		},
		{
			name: "disabled setting uses default refresh rate on battery",
			settings: editor_settings.Settings{
				RefreshRate:           144,
				UseBatteryRefreshRate: false,
				BatteryRefreshRate:    30,
			},
			status: power.Status{Source: power.SourceBattery, HasBattery: true},
			want:   144,
		},
		{
			name: "refresh rate is clamped",
			settings: editor_settings.Settings{
				RefreshRate:           600,
				UseBatteryRefreshRate: false,
			},
			status: power.Status{Source: power.SourceAC},
			want:   320,
		},
		{
			name: "battery refresh rate is clamped",
			settings: editor_settings.Settings{
				RefreshRate:           144,
				UseBatteryRefreshRate: true,
				BatteryRefreshRate:    -1,
			},
			status: power.Status{Source: power.SourceBattery, HasBattery: true},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := &Editor{settings: tt.settings}
			if got := ed.effectiveRefreshRate(tt.status); got != tt.want {
				t.Fatalf("effectiveRefreshRate() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEditorPowerPollQueriesAndCachesSourceChanges(t *testing.T) {
	statuses := []power.Status{
		{Source: power.SourceAC, HasBattery: true},
		{Source: power.SourceBattery, HasBattery: true},
	}
	queryCount := 0
	ed := &Editor{
		settings: editor_settings.Settings{
			RefreshRate:           60,
			UseBatteryRefreshRate: true,
			BatteryRefreshRate:    30,
		},
	}
	ed.power.query = func() (power.Status, error) {
		status := statuses[min(queryCount, len(statuses)-1)]
		queryCount++
		return status, nil
	}
	ed.updatePowerState(editorPowerPollInterval)
	if queryCount != 1 || ed.power.lastStatus.Source != power.SourceAC {
		t.Fatalf("expected initial AC status to be cached, count=%d status=%#v", queryCount, ed.power.lastStatus)
	}
	ed.updatePowerState(editorPowerPollInterval)
	if queryCount != 2 || ed.power.lastStatus.Source != power.SourceBattery {
		t.Fatalf("expected battery status change to be cached, count=%d status=%#v", queryCount, ed.power.lastStatus)
	}
}
