/******************************************************************************/
/* render_target.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"fmt"
	"sync"

	"kaijuengine.com/platform/profiler/tracing"
)

const (
	RenderTargetOutputColor = "color"
	RenderTargetOutputDepth = "depth"
)

var (
	ErrRenderTargetNotFound    = errors.New("render target not found")
	ErrRenderTargetNotRealized = errors.New("render target texture is not realized")
	ErrRenderTargetDestroyed   = errors.New("render target is destroyed")
)

type RenderTargetResizeMode int

const (
	RenderTargetResizeModeFixed RenderTargetResizeMode = iota
	RenderTargetResizeModeMatchWindow
)

type RenderTargetOptions struct {
	Name        string
	Width       int
	Height      int
	ResizeMode  RenderTargetResizeMode
	ColorFormat GPUFormat
	Depth       bool
}

type renderTargetPendingOp uint8

const (
	renderTargetPendingResize renderTargetPendingOp = iota
	renderTargetPendingDestroy
)

type RenderTarget struct {
	manager     *RenderTargetManager
	options     RenderTargetOptions
	width       int
	height      int
	resizeDirty bool
	destroyed   bool
	outputs     map[string]*Texture
	mutex       sync.RWMutex
}

func newRenderTarget(options RenderTargetOptions, manager *RenderTargetManager) *RenderTarget {
	outputs := map[string]*Texture{
		RenderTargetOutputColor: nil,
	}
	if options.Depth {
		outputs[RenderTargetOutputDepth] = nil
	}
	return &RenderTarget{
		manager: manager,
		options: options,
		width:   options.Width,
		height:  options.Height,
		outputs: outputs,
	}
}

func (t *RenderTarget) Name() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.options.Name
}

func (t *RenderTarget) Options() RenderTargetOptions {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.options
}

func (t *RenderTarget) Width() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.width
}

func (t *RenderTarget) Height() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.height
}

func (t *RenderTarget) Size() (int, int) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.width, t.height
}

func (t *RenderTarget) ResizeDirty() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.resizeDirty
}

func (t *RenderTarget) Destroyed() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.destroyed
}

func (t *RenderTarget) Resize(width, height int) bool {
	defer tracing.NewRegion("RenderTarget.Resize").End()
	if width <= 0 || height <= 0 {
		return false
	}
	t.mutex.Lock()
	if t.destroyed || t.width == width && t.height == height {
		t.mutex.Unlock()
		return false
	}
	t.width = width
	t.height = height
	t.resizeDirty = true
	manager := t.manager
	t.mutex.Unlock()
	if manager != nil {
		manager.queuePending(t, renderTargetPendingResize)
	}
	return true
}

func (t *RenderTarget) Texture(name string) (*Texture, error) {
	defer tracing.NewRegion("RenderTarget.Texture").End()
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if t.destroyed {
		return nil, fmt.Errorf("%w: %s", ErrRenderTargetDestroyed, t.options.Name)
	}
	tex, ok := t.outputs[name]
	if !ok {
		return nil, fmt.Errorf("render target %q has no output texture %q", t.options.Name, name)
	}
	if tex == nil || !tex.RenderId.IsValid() {
		return nil, fmt.Errorf("%w: %s.%s", ErrRenderTargetNotRealized, t.options.Name, name)
	}
	return tex, nil
}

func (t *RenderTarget) setTexture(name string, texture *Texture) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.destroyed {
		return fmt.Errorf("%w: %s", ErrRenderTargetDestroyed, t.options.Name)
	}
	if _, ok := t.outputs[name]; !ok {
		return fmt.Errorf("render target %q has no output texture %q", t.options.Name, name)
	}
	t.outputs[name] = texture
	if texture != nil && texture.RenderId.IsValid() {
		t.resizeDirty = false
	}
	return nil
}

func (t *RenderTarget) processPending(device *GPUDevice, op renderTargetPendingOp) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	switch op {
	case renderTargetPendingDestroy:
		t.destroyed = true
		t.releaseOutputsLocked(device)
	case renderTargetPendingResize:
		t.releaseOutputsLocked(device)
	}
}

func (t *RenderTarget) releaseOutputsLocked(device *GPUDevice) {
	for name, tex := range t.outputs {
		if tex != nil {
			if tex.RenderId.IsValid() {
				if device == nil {
					continue
				}
				device.LogicalDevice.FreeTexture(&tex.RenderId)
			}
			t.outputs[name] = nil
		}
	}
}

type RenderTargetManager struct {
	targets map[string]*RenderTarget
	pending map[*RenderTarget]renderTargetPendingOp
	mutex   sync.RWMutex
}

func NewRenderTargetManager() RenderTargetManager {
	return RenderTargetManager{
		targets: make(map[string]*RenderTarget),
		pending: make(map[*RenderTarget]renderTargetPendingOp),
	}
}

func (m *RenderTargetManager) Create(options RenderTargetOptions) (*RenderTarget, error) {
	defer tracing.NewRegion("RenderTargetManager.Create").End()
	if err := validateRenderTargetOptions(options); err != nil {
		return nil, err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ensureLocked()
	if _, ok := m.targets[options.Name]; ok {
		return nil, fmt.Errorf("render target %q already exists", options.Name)
	}
	target := newRenderTarget(options, m)
	m.targets[options.Name] = target
	return target, nil
}

func (m *RenderTargetManager) Target(name string) (*RenderTarget, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	target, ok := m.targets[name]
	return target, ok
}

func (m *RenderTargetManager) Destroy(name string) error {
	defer tracing.NewRegion("RenderTargetManager.Destroy").End()
	m.mutex.Lock()
	m.ensureLocked()
	target, ok := m.targets[name]
	if ok {
		delete(m.targets, name)
	}
	m.mutex.Unlock()
	if !ok {
		return fmt.Errorf("%w: %s", ErrRenderTargetNotFound, name)
	}
	m.queuePending(target, renderTargetPendingDestroy)
	return nil
}

func (m *RenderTargetManager) ProcessPending(device *GPUDevice) {
	defer tracing.NewRegion("RenderTargetManager.ProcessPending").End()
	m.mutex.Lock()
	m.ensureLocked()
	pending := make(map[*RenderTarget]renderTargetPendingOp, len(m.pending))
	for target, op := range m.pending {
		pending[target] = op
	}
	for target := range m.pending {
		delete(m.pending, target)
	}
	m.mutex.Unlock()
	for target, op := range pending {
		target.processPending(device, op)
	}
}

func (m *RenderTargetManager) DestroyAll(device *GPUDevice) {
	defer tracing.NewRegion("RenderTargetManager.DestroyAll").End()
	m.mutex.Lock()
	m.ensureLocked()
	targets := make([]*RenderTarget, 0, len(m.targets)+len(m.pending))
	seen := make(map[*RenderTarget]struct{}, len(m.targets)+len(m.pending))
	for _, target := range m.targets {
		targets = append(targets, target)
		seen[target] = struct{}{}
	}
	for target := range m.pending {
		if _, ok := seen[target]; !ok {
			targets = append(targets, target)
		}
	}
	m.targets = make(map[string]*RenderTarget)
	m.pending = make(map[*RenderTarget]renderTargetPendingOp)
	m.mutex.Unlock()
	for _, target := range targets {
		target.processPending(device, renderTargetPendingDestroy)
	}
}

func (m *RenderTargetManager) queuePending(target *RenderTarget, op renderTargetPendingOp) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ensureLocked()
	if existing, ok := m.pending[target]; ok && existing == renderTargetPendingDestroy {
		return
	}
	m.pending[target] = op
}

func (m *RenderTargetManager) ensureLocked() {
	if m.targets == nil {
		m.targets = make(map[string]*RenderTarget)
	}
	if m.pending == nil {
		m.pending = make(map[*RenderTarget]renderTargetPendingOp)
	}
}

func validateRenderTargetOptions(options RenderTargetOptions) error {
	if options.Name == "" {
		return errors.New("render target name is required")
	}
	if options.Width <= 0 || options.Height <= 0 {
		return fmt.Errorf("render target %q has invalid size %dx%d", options.Name, options.Width, options.Height)
	}
	switch options.ResizeMode {
	case RenderTargetResizeModeFixed, RenderTargetResizeModeMatchWindow:
	default:
		return fmt.Errorf("render target %q has invalid resize mode %d", options.Name, options.ResizeMode)
	}
	return nil
}
