package common_workspace

import (
	"fmt"
	"strconv"

	"kaijuengine.com/engine/lighting/gi"
)

type GISettingsUIData struct {
	Preset                    string
	Mode                      string
	Fallback                  string
	GPUTimeBudgetMS           string
	MemoryBudgetMB            string
	CoverageDistance          string
	ProbeSpacing              string
	CascadeCount              string
	RaysPerProbe              string
	MaxProbeUpdatesPerFrame   string
	ResolveScale              string
	UpdateHz                  string
	HistoryWeight             string
	ContactDetail             string
	DynamicGeometry           string
	EmissiveParticipation     string
	ScenarioTransitionSeconds string
	AdaptiveBudget            bool
}

func NewGISettingsUIData(s gi.Settings) GISettingsUIData {
	f := func(v float32) string { return strconv.FormatFloat(float64(v), 'f', -1, 32) }
	u := func(v uint32) string { return strconv.FormatUint(uint64(v), 10) }
	return GISettingsUIData{
		Preset:                    strconv.Itoa(int(s.Preset)),
		Mode:                      strconv.Itoa(int(s.Mode)),
		Fallback:                  strconv.Itoa(int(s.Fallback)),
		GPUTimeBudgetMS:           f(s.GPUTimeBudgetMS),
		MemoryBudgetMB:            u(s.MemoryBudgetMB),
		CoverageDistance:          f(s.CoverageDistance),
		ProbeSpacing:              f(s.ProbeSpacing),
		CascadeCount:              u(s.CascadeCount),
		RaysPerProbe:              u(s.RaysPerProbe),
		MaxProbeUpdatesPerFrame:   u(s.MaxProbeUpdatesPerFrame),
		ResolveScale:              f(s.ResolveScale),
		UpdateHz:                  f(s.UpdateHz),
		HistoryWeight:             f(s.HistoryWeight),
		ContactDetail:             strconv.Itoa(int(s.ContactDetail)),
		DynamicGeometry:           strconv.Itoa(int(s.DynamicGeometry)),
		EmissiveParticipation:     strconv.Itoa(int(s.EmissiveParticipation)),
		ScenarioTransitionSeconds: f(s.ScenarioTransitionSeconds),
		AdaptiveBudget:            s.AdaptiveBudget,
	}
}

func ApplyGISettingsField(current gi.Settings, field, value string, checked bool) (gi.Settings, error) {
	next := current
	parseInt := func() (int, error) { return strconv.Atoi(value) }
	parseUint := func() (uint32, error) {
		v, err := strconv.ParseUint(value, 10, 32)
		return uint32(v), err
	}
	parseFloat := func() (float32, error) {
		v, err := strconv.ParseFloat(value, 32)
		return float32(v), err
	}
	var err error
	switch field {
	case "Preset":
		var v int
		if v, err = parseInt(); err == nil && v >= int(gi.QualityPresetOff) && v <= int(gi.QualityPresetCustom) {
			next = gi.SettingsForPreset(gi.QualityPreset(v))
		} else if err == nil {
			err = fmt.Errorf("invalid GI preset %d", v)
		}
	case "Mode":
		var v int
		v, err = parseInt()
		next.Mode = gi.Mode(v)
	case "Fallback":
		var v int
		v, err = parseInt()
		next.Fallback = gi.FallbackPolicy(v)
	case "GPUTimeBudgetMS":
		next.GPUTimeBudgetMS, err = parseFloat()
	case "MemoryBudgetMB":
		next.MemoryBudgetMB, err = parseUint()
	case "CoverageDistance":
		next.CoverageDistance, err = parseFloat()
	case "ProbeSpacing":
		next.ProbeSpacing, err = parseFloat()
	case "CascadeCount":
		next.CascadeCount, err = parseUint()
	case "RaysPerProbe":
		next.RaysPerProbe, err = parseUint()
	case "MaxProbeUpdatesPerFrame":
		next.MaxProbeUpdatesPerFrame, err = parseUint()
	case "ResolveScale":
		next.ResolveScale, err = parseFloat()
	case "UpdateHz":
		next.UpdateHz, err = parseFloat()
	case "HistoryWeight":
		next.HistoryWeight, err = parseFloat()
	case "ContactDetail":
		var v int
		v, err = parseInt()
		next.ContactDetail = gi.ContactDetailMode(v)
	case "DynamicGeometry":
		var v int
		v, err = parseInt()
		next.DynamicGeometry = gi.DynamicGeometryMode(v)
	case "EmissiveParticipation":
		var v int
		v, err = parseInt()
		next.EmissiveParticipation = gi.EmissiveParticipationMode(v)
	case "ScenarioTransitionSeconds":
		next.ScenarioTransitionSeconds, err = parseFloat()
	case "AdaptiveBudget":
		next.AdaptiveBudget = checked
	default:
		err = fmt.Errorf("unknown GI settings field %q", field)
	}
	if err != nil {
		return current, err
	}
	if field != "Preset" {
		next.Preset = gi.QualityPresetCustom
	}
	if err := next.Validate(); err != nil {
		return current, err
	}
	return next, nil
}
