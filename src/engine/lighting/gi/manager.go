/******************************************************************************/
/* manager.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import (
	"errors"
	"fmt"
	"sync"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type Manager struct {
	mutex           sync.RWMutex
	capabilities    Capabilities
	defaultSettings Settings
	settings        Settings
	factories       map[string]ProviderFactory
	provider        Provider
	lastError       error
	assets          AssetReader
	fallbackReason  string
}

func NewManager(capabilities Capabilities) *Manager {
	m := &Manager{
		capabilities:    capabilities,
		defaultSettings: SettingsForPreset(QualityPresetOff),
		settings:        SettingsForPreset(QualityPresetOff),
		factories:       make(map[string]ProviderFactory),
	}
	m.factories[ProviderNull] = func() Provider { return &NullProvider{} }
	m.factories[ProviderBakedProbe] = func() Provider { return &BakedProbeProvider{} }
	provider := m.factories[ProviderNull]()
	_ = provider.Initialize(ProviderContext{Capabilities: capabilities, Assets: m.assets})
	_ = provider.Configure(m.settings)
	m.provider = provider
	return m
}

// SetDefaultSettings stores the project-wide GI settings and applies them to
// the active provider. Stages may temporarily replace these settings through
// ApplyStageSettings without losing the project baseline.
func (m *Manager) SetDefaultSettings(settings Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if err := m.configureLocked(settings); err != nil {
		return err
	}
	m.defaultSettings = settings
	return nil
}

func (m *Manager) DefaultSettings() Settings {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.defaultSettings
}

// ApplyStageSettings selects either a stage override or the project defaults,
// clears the previous stage's scenario, and then loads the requested probe
// asset. A scenario load failure therefore cannot leak lighting from the
// previously loaded stage.
func (m *Manager) ApplyStageSettings(override *Settings, scenarioAsset string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	settings := m.defaultSettings
	if override != nil {
		settings = *override
	}
	if err := settings.Validate(); err != nil {
		return err
	}
	if err := m.configureLocked(settings); err != nil {
		return err
	}
	if err := m.provider.SetScenario(""); err != nil {
		m.lastError = err
		return err
	}
	if scenarioAsset == "" || settings.Mode == ModeDisabled || settings.Preset == QualityPresetOff {
		m.lastError = nil
		return nil
	}
	if err := m.provider.SetScenario(scenarioAsset); err != nil {
		m.lastError = err
		return err
	}
	m.lastError = nil
	return nil
}

func (m *Manager) ClearScenario() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.provider == nil {
		return nil
	}
	return m.provider.SetScenario("")
}

// SetAssetReader supplies the content source used by asset-backed providers.
// It should normally be called before selecting a baked provider.
func (m *Manager) SetAssetReader(reader AssetReader) {
	m.mutex.Lock()
	m.assets = reader
	m.mutex.Unlock()
}

func (m *Manager) RegisterProvider(id string, factory ProviderFactory) error {
	if id == "" {
		return errors.New("GI provider id is empty")
	}
	if factory == nil {
		return fmt.Errorf("GI provider %q has a nil factory", id)
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, exists := m.factories[id]; exists {
		return fmt.Errorf("GI provider %q is already registered", id)
	}
	m.factories[id] = factory
	return nil
}

func (m *Manager) SetCapabilities(capabilities Capabilities) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.capabilities = capabilities
	return m.configureLocked(m.settings)
}

func (m *Manager) Capabilities() Capabilities {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.capabilities
}

func (m *Manager) Configure(settings Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.configureLocked(settings)
}

func (m *Manager) configureLocked(settings Settings) error {
	candidates := m.providerCandidates(settings)
	var failures []error
	for _, id := range candidates {
		factory, exists := m.factories[id]
		if !exists {
			failures = append(failures, fmt.Errorf("GI provider %q is not registered", id))
			continue
		}
		if m.provider != nil && m.provider.ID() == id && m.provider.Supports(m.capabilities) {
			if err := m.provider.Configure(settings); err == nil {
				m.settings = settings
				m.lastError = nil
				m.fallbackReason = joinedErrorString(failures)
				return nil
			} else {
				failures = append(failures, fmt.Errorf("configure GI provider %q: %w", id, err))
				if settings.Fallback == FallbackRequireExact {
					break
				}
			}
		}
		candidate := factory()
		if candidate == nil {
			failures = append(failures, fmt.Errorf("GI provider factory %q returned nil", id))
			continue
		}
		if !candidate.Supports(m.capabilities) {
			failures = append(failures, fmt.Errorf("GI provider %q is unsupported", id))
			candidate.Shutdown()
			continue
		}
		if err := candidate.Initialize(ProviderContext{Capabilities: m.capabilities, Assets: m.assets}); err != nil {
			failures = append(failures, fmt.Errorf("initialize GI provider %q: %w", id, err))
			candidate.Shutdown()
			continue
		}
		if err := candidate.Configure(settings); err != nil {
			failures = append(failures, fmt.Errorf("configure GI provider %q: %w", id, err))
			candidate.Shutdown()
			continue
		}
		previous := m.provider
		m.provider = candidate
		m.settings = settings
		m.lastError = nil
		m.fallbackReason = joinedErrorString(failures)
		if previous != nil {
			previous.Shutdown()
		}
		return nil
	}
	err := errors.Join(failures...)
	if err == nil {
		err = errors.New("no GI provider candidates are available")
	}
	m.lastError = err
	return err
}

func joinedErrorString(failures []error) string {
	if len(failures) == 0 {
		return ""
	}
	return errors.Join(failures...).Error()
}

func (m *Manager) providerCandidates(settings Settings) []string {
	if settings.Mode == ModeDisabled || settings.Preset == QualityPresetOff {
		return []string{ProviderNull}
	}
	var candidates []string
	switch settings.Mode {
	case ModeAuto:
		if m.capabilities.SupportsDynamicDDGI() &&
			(settings.Preset == QualityPresetHigh || settings.Preset == QualityPresetUltra || settings.Preset == QualityPresetCustom) {
			candidates = append(candidates, ProviderDDGI)
		}
		candidates = append(candidates, ProviderBakedProbe, ProviderNull)
	case ModeBaked:
		candidates = append(candidates, ProviderBakedProbe)
	case ModeDynamicDDGI:
		candidates = append(candidates, ProviderDDGI)
		if settings.Fallback == FallbackAllow {
			candidates = append(candidates, ProviderBakedProbe, ProviderNull)
		}
	}
	if settings.Fallback == FallbackRequireExact && len(candidates) > 1 {
		candidates = candidates[:1]
	}
	return candidates
}

func (m *Manager) Settings() Settings {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings
}

func (m *Manager) ActiveProvider() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.provider == nil {
		return ""
	}
	return m.provider.ID()
}

func (m *Manager) LastError() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.lastError
}

func (m *Manager) SyncScene(delta SceneDelta) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.provider.SyncScene(delta)
}

func (m *Manager) AddFramePasses(graph *rendering.FrameGraph, inputs FrameInputs) (Outputs, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if err := m.provider.AddUpdatePasses(graph, inputs); err != nil {
		return Outputs{}, err
	}
	return m.provider.AddResolvePasses(graph, inputs)
}

func (m *Manager) ProbeField(view ViewID) ProbeFieldBinding {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.provider.ProbeField(view)
}

func (m *Manager) ShaderData(position matrix.Vec3, runtimeSeconds float32) rendering.GlobalIlluminationForRender {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.provider.ShaderData(position, runtimeSeconds)
}

func (m *Manager) Invalidate(invalidation Invalidation) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.provider.Invalidate(invalidation)
}

func (m *Manager) ResetHistory(view ViewID) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	m.provider.ResetHistory(view)
}

func (m *Manager) SetScenario(id string) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.provider.SetScenario(id)
}

func (m *Manager) Stats() Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	stats := m.provider.Stats()
	if m.fallbackReason != "" {
		if stats.FallbackReason == "" {
			stats.FallbackReason = m.fallbackReason
		} else {
			stats.FallbackReason = m.fallbackReason + "; " + stats.FallbackReason
		}
	}
	return stats
}

func (m *Manager) DebugViews() []DebugView {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.provider.DebugViews()
}

func (m *Manager) Shutdown() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.provider != nil {
		m.provider.Shutdown()
		m.provider = nil
	}
}
