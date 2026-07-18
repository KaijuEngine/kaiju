/******************************************************************************/
/* settings.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import "fmt"

type Mode uint8

const (
	ModeDisabled Mode = iota
	ModeAuto
	ModeBaked
	ModeDynamicDDGI
)

type QualityPreset uint8

const (
	QualityPresetOff QualityPreset = iota
	QualityPresetLow
	QualityPresetMedium
	QualityPresetHigh
	QualityPresetUltra
	QualityPresetCustom
)

type FallbackPolicy uint8

const (
	FallbackAllow FallbackPolicy = iota
	FallbackRequireExact
)

type ContactDetailMode uint8

const (
	ContactDetailOff ContactDetailMode = iota
	ContactDetailGTAO
)

type DynamicGeometryMode uint8

const (
	DynamicGeometryStaticOnly DynamicGeometryMode = iota
	DynamicGeometryStaticAndRigid
)

type EmissiveParticipationMode uint8

const (
	EmissiveParticipationOff EmissiveParticipationMode = iota
	EmissiveParticipationStatic
	EmissiveParticipationDynamic
)

// Settings is the complete runtime configuration for global illumination.
// SettingsForPreset returns fully populated values that callers may override.
type Settings struct {
	Mode                      Mode
	Preset                    QualityPreset
	Fallback                  FallbackPolicy
	GPUTimeBudgetMS           float32
	MemoryBudgetMB            uint32
	CoverageDistance          float32
	ProbeSpacing              float32
	CascadeCount              uint32
	RaysPerProbe              uint32
	MaxProbeUpdatesPerFrame   uint32
	ResolveScale              float32
	UpdateHz                  float32
	HistoryWeight             float32
	ContactDetail             ContactDetailMode
	DynamicGeometry           DynamicGeometryMode
	EmissiveParticipation     EmissiveParticipationMode
	ScenarioTransitionSeconds float32
	AdaptiveBudget            bool
}

func DefaultSettings() Settings { return SettingsForPreset(QualityPresetMedium) }

func SettingsForPreset(preset QualityPreset) Settings {
	base := Settings{
		Mode:                      ModeAuto,
		Preset:                    preset,
		Fallback:                  FallbackAllow,
		CoverageDistance:          192,
		ProbeSpacing:              2,
		ResolveScale:              0.5,
		UpdateHz:                  60,
		HistoryWeight:             0.97,
		ContactDetail:             ContactDetailGTAO,
		DynamicGeometry:           DynamicGeometryStaticAndRigid,
		EmissiveParticipation:     EmissiveParticipationStatic,
		ScenarioTransitionSeconds: 1,
		AdaptiveBudget:            true,
	}
	switch preset {
	case QualityPresetOff:
		base.Mode = ModeDisabled
		base.MemoryBudgetMB = 0
		base.GPUTimeBudgetMS = 0
		base.ResolveScale = 0
		base.ContactDetail = ContactDetailOff
	case QualityPresetLow:
		base.MemoryBudgetMB = 48
		base.GPUTimeBudgetMS = 0.5
		base.ProbeSpacing = 4
		base.ResolveScale = 0.5
		base.UpdateHz = 15
	case QualityPresetMedium:
		base.MemoryBudgetMB = 96
		base.GPUTimeBudgetMS = 1
		base.ProbeSpacing = 2
		base.ResolveScale = 0.5
		base.UpdateHz = 30
	case QualityPresetHigh:
		base.MemoryBudgetMB = 160
		base.GPUTimeBudgetMS = 1.5
		base.ProbeSpacing = 2
		base.CascadeCount = 3
		base.RaysPerProbe = 64
		base.MaxProbeUpdatesPerFrame = 384
		base.ResolveScale = 0.5
		base.UpdateHz = 60
		base.EmissiveParticipation = EmissiveParticipationDynamic
	case QualityPresetUltra:
		base.MemoryBudgetMB = 256
		base.GPUTimeBudgetMS = 3
		base.CoverageDistance = 336
		base.ProbeSpacing = 1.5
		base.CascadeCount = 4
		base.RaysPerProbe = 128
		base.MaxProbeUpdatesPerFrame = 768
		base.ResolveScale = 1
		base.UpdateHz = 60
		base.HistoryWeight = 0.95
		base.EmissiveParticipation = EmissiveParticipationDynamic
	case QualityPresetCustom:
		base.MemoryBudgetMB = 96
		base.GPUTimeBudgetMS = 1
	}
	return base
}

func (s Settings) Validate() error {
	if s.Mode > ModeDynamicDDGI {
		return fmt.Errorf("invalid GI mode %d", s.Mode)
	}
	if s.Preset > QualityPresetCustom {
		return fmt.Errorf("invalid GI quality preset %d", s.Preset)
	}
	if s.Fallback > FallbackRequireExact {
		return fmt.Errorf("invalid GI fallback policy %d", s.Fallback)
	}
	if s.Mode == ModeDisabled || s.Preset == QualityPresetOff {
		return nil
	}
	if s.MemoryBudgetMB == 0 {
		return fmt.Errorf("GI memory budget must be greater than zero")
	}
	if s.GPUTimeBudgetMS <= 0 {
		return fmt.Errorf("GI GPU time budget must be greater than zero")
	}
	if s.CoverageDistance <= 0 {
		return fmt.Errorf("GI coverage distance must be greater than zero")
	}
	if s.ProbeSpacing <= 0 {
		return fmt.Errorf("GI probe spacing must be greater than zero")
	}
	if s.ResolveScale <= 0 || s.ResolveScale > 1 {
		return fmt.Errorf("GI resolve scale must be in (0, 1]")
	}
	if s.UpdateHz <= 0 {
		return fmt.Errorf("GI update frequency must be greater than zero")
	}
	if s.HistoryWeight < 0 || s.HistoryWeight >= 1 {
		return fmt.Errorf("GI history weight must be in [0, 1)")
	}
	if s.ScenarioTransitionSeconds < 0 {
		return fmt.Errorf("GI scenario transition cannot be negative")
	}
	if s.Mode == ModeDynamicDDGI || s.Preset == QualityPresetHigh || s.Preset == QualityPresetUltra {
		if s.CascadeCount == 0 {
			return fmt.Errorf("dynamic GI requires at least one cascade")
		}
		if s.RaysPerProbe < 32 {
			return fmt.Errorf("dynamic GI requires at least 32 rays per probe")
		}
		if s.MaxProbeUpdatesPerFrame == 0 {
			return fmt.Errorf("dynamic GI requires a non-zero probe update budget")
		}
	}
	return nil
}
