/******************************************************************************/
/* render_view.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

const DefaultRenderViewName = "default"

type RenderViewMode int

const (
	RenderViewModeNormal RenderViewMode = iota
	RenderViewModeWireframe
	RenderViewModeUnlit
	RenderViewModeProfile

	RenderViewModeDefault = RenderViewModeNormal
)

type RenderViewOptions struct {
	Name      string
	Target    *RenderTarget
	Camera    any
	LayerMask RenderLayerMask
	Clear     bool
	Sort      int
	ViewMode  RenderViewMode
}

type RenderView struct {
	options            RenderViewOptions
	order              uint64
	enabled            bool
	destroyed          bool
	previousView       matrix.Mat4
	previousProjection matrix.Mat4
	historyValid       bool
	historyReset       bool
	mutex              sync.RWMutex
}

type RenderViewFrame struct {
	View               *RenderView
	Options            RenderViewOptions
	Order              uint64
	Enabled            bool
	Destroyed          bool
	CurrentView        matrix.Mat4
	CurrentProjection  matrix.Mat4
	PreviousView       matrix.Mat4
	PreviousProjection matrix.Mat4
	HistoryValid       bool
	HistoryReset       bool
}

type renderViewMatrixCamera interface {
	View() matrix.Mat4
	Projection() matrix.Mat4
}

func newRenderView(options RenderViewOptions, order uint64) *RenderView {
	options.LayerMask = normalizeRenderLayerMask(options.LayerMask)
	return &RenderView{
		options:      options,
		order:        order,
		enabled:      true,
		historyReset: true,
	}
}

func newRenderViewFrame(view *RenderView) RenderViewFrame {
	if view == nil {
		return RenderViewFrame{}
	}
	view.mutex.Lock()
	defer view.mutex.Unlock()
	frame := RenderViewFrame{
		View:      view,
		Options:   view.options,
		Order:     view.order,
		Enabled:   view.enabled,
		Destroyed: view.destroyed,
	}
	if camera, ok := view.options.Camera.(renderViewMatrixCamera); ok {
		frame.CurrentView = camera.View()
		frame.CurrentProjection = camera.Projection()
		frame.HistoryValid = view.historyValid && !view.historyReset
		frame.HistoryReset = !frame.HistoryValid
		if frame.HistoryValid {
			frame.PreviousView = view.previousView
			frame.PreviousProjection = view.previousProjection
		} else {
			frame.PreviousView = frame.CurrentView
			frame.PreviousProjection = frame.CurrentProjection
		}
		view.previousView = frame.CurrentView
		view.previousProjection = frame.CurrentProjection
		view.historyValid = true
		view.historyReset = false
	} else {
		frame.HistoryReset = true
	}
	return frame
}

func (v RenderViewFrame) Name() string {
	if v.Options.Name == "" {
		return DefaultRenderViewName
	}
	return v.Options.Name
}

func (v RenderViewFrame) Target() *RenderTarget { return v.Options.Target }
func (v RenderViewFrame) Camera() any           { return v.Options.Camera }
func (v RenderViewFrame) LayerMask() RenderLayerMask {
	return normalizeRenderLayerMask(v.Options.LayerMask)
}
func (v RenderViewFrame) Clear() bool              { return v.Options.Clear }
func (v RenderViewFrame) Sort() int                { return v.Options.Sort }
func (v RenderViewFrame) ViewMode() RenderViewMode { return v.Options.ViewMode }
func (v RenderViewFrame) IsDestroyed() bool        { return v.Destroyed || v.View == nil }
func (v RenderViewFrame) IsEnabled() bool          { return v.Enabled && !v.IsDestroyed() }
func (v RenderViewFrame) Key() *RenderView         { return v.View }
func (v RenderViewFrame) ID() uint64               { return v.Order + 1 }

func (v *RenderView) Name() string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options.Name
}

func (v *RenderView) Options() RenderViewOptions {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options
}

func (v *RenderView) Target() *RenderTarget {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options.Target
}

func (v *RenderView) Camera() any {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options.Camera
}

func (v *RenderView) SetCamera(camera any) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if sameRenderViewCamera(v.options.Camera, camera) {
		return
	}
	v.options.Camera = camera
	v.historyReset = true
}

func sameRenderViewCamera(left, right any) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	lv := reflect.ValueOf(left)
	rv := reflect.ValueOf(right)
	return lv.IsValid() && rv.IsValid() && lv.Type() == rv.Type() &&
		lv.Kind() == reflect.Pointer && lv.Pointer() == rv.Pointer()
}

// ResetHistory invalidates temporal data for this view. Call this after a
// camera cut, teleporter, projection discontinuity, or render-target resize.
func (v *RenderView) ResetHistory() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.historyReset = true
}

func (v *RenderView) LayerMask() RenderLayerMask {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return normalizeRenderLayerMask(v.options.LayerMask)
}

func (v *RenderView) Clear() bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options.Clear
}

func (v *RenderView) Sort() int {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options.Sort
}

func (v *RenderView) ViewMode() RenderViewMode {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.options.ViewMode
}

func (v *RenderView) Enabled() bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.enabled && !v.destroyed
}

func (v *RenderView) SetEnabled(enabled bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.enabled = enabled
}

func (v *RenderView) SetViewMode(mode RenderViewMode) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.options.ViewMode = mode
}

func (v *RenderView) Destroyed() bool {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.destroyed
}

func (v *RenderView) MatchesDrawing(drawing *Drawing) bool {
	if drawing == nil || !v.Enabled() {
		return false
	}
	return drawing.MatchesLayer(v.LayerMask())
}

func (v *RenderView) MatchesGroup(group *DrawInstanceGroup) bool {
	if group == nil || !v.Enabled() {
		return false
	}
	return group.MatchesLayer(v.LayerMask())
}

func (v *RenderView) setOptions(options RenderViewOptions) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	options.LayerMask = normalizeRenderLayerMask(options.LayerMask)
	v.options = options
	v.enabled = true
	v.destroyed = false
	v.historyReset = true
}

func (v *RenderView) markDestroyed() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.destroyed = true
}

type RenderViewManager struct {
	views          map[string]*RenderView
	pendingDestroy map[*RenderView]struct{}
	nextOrder      uint64
	mutex          sync.RWMutex
}

func NewRenderViewManager(defaultOptions ...RenderViewOptions) RenderViewManager {
	options := RenderViewOptions{
		Name:      DefaultRenderViewName,
		LayerMask: RenderLayerAll,
		Clear:     true,
	}
	if len(defaultOptions) > 0 {
		options = defaultOptions[0]
		if options.Name == "" {
			options.Name = DefaultRenderViewName
		}
	}
	manager := RenderViewManager{
		views:          make(map[string]*RenderView),
		pendingDestroy: make(map[*RenderView]struct{}),
	}
	view := newRenderView(options, manager.nextOrder)
	manager.nextOrder++
	manager.views[view.Name()] = view
	return manager
}

func (m *RenderViewManager) Create(options RenderViewOptions) (*RenderView, error) {
	defer tracing.NewRegion("RenderViewManager.Create").End()
	if err := validateRenderViewOptions(options); err != nil {
		return nil, err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ensureLocked()
	if _, ok := m.views[options.Name]; ok {
		return nil, fmt.Errorf("render view %q already exists", options.Name)
	}
	view := newRenderView(options, m.nextOrder)
	m.nextOrder++
	m.views[options.Name] = view
	return view, nil
}

func (m *RenderViewManager) View(name string) (*RenderView, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	view, ok := m.views[name]
	return view, ok
}

func (m *RenderViewManager) Default() (*RenderView, bool) {
	return m.View(DefaultRenderViewName)
}

func (m *RenderViewManager) SetDefaultCamera(camera any) {
	if view, ok := m.Default(); ok {
		view.SetCamera(camera)
	}
}

func (m *RenderViewManager) ResetHistory(name string) error {
	view, ok := m.View(name)
	if !ok {
		return fmt.Errorf("render view %q not found", name)
	}
	view.ResetHistory()
	return nil
}

func (m *RenderViewManager) ReplaceDefault(options RenderViewOptions) (*RenderView, error) {
	defer tracing.NewRegion("RenderViewManager.ReplaceDefault").End()
	if options.Name == "" {
		options.Name = DefaultRenderViewName
	}
	if options.Name != DefaultRenderViewName {
		return nil, fmt.Errorf("default render view must be named %q", DefaultRenderViewName)
	}
	if err := validateRenderViewOptions(options); err != nil {
		return nil, err
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ensureLocked()
	view, ok := m.views[DefaultRenderViewName]
	if !ok {
		view = newRenderView(options, m.nextOrder)
		m.nextOrder++
		m.views[DefaultRenderViewName] = view
	} else {
		view.setOptions(options)
	}
	return view, nil
}

func (m *RenderViewManager) Destroy(name string) error {
	defer tracing.NewRegion("RenderViewManager.Destroy").End()
	if name == DefaultRenderViewName {
		return fmt.Errorf("cannot destroy default render view %q", DefaultRenderViewName)
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ensureLocked()
	view, ok := m.views[name]
	if !ok {
		return fmt.Errorf("render view %q not found", name)
	}
	view.markDestroyed()
	delete(m.views, name)
	m.pendingDestroy[view] = struct{}{}
	return nil
}

func (m *RenderViewManager) ProcessPending(device *GPUDevice, drawings *Drawings) {
	defer tracing.NewRegion("RenderViewManager.ProcessPending").End()
	m.mutex.Lock()
	m.ensureLocked()
	pending := make([]*RenderView, 0, len(m.pendingDestroy))
	for view := range m.pendingDestroy {
		pending = append(pending, view)
		delete(m.pendingDestroy, view)
	}
	m.mutex.Unlock()
	for i := range pending {
		destroyRenderViewResources(pending[i], device, drawings)
	}
}

func (m *RenderViewManager) DestroyAll(device *GPUDevice, drawings *Drawings) {
	defer tracing.NewRegion("RenderViewManager.DestroyAll").End()
	m.mutex.Lock()
	m.ensureLocked()
	views := make([]*RenderView, 0, len(m.views)+len(m.pendingDestroy))
	seen := make(map[*RenderView]struct{}, len(m.views)+len(m.pendingDestroy))
	for _, view := range m.views {
		view.markDestroyed()
		views = append(views, view)
		seen[view] = struct{}{}
	}
	for view := range m.pendingDestroy {
		if _, ok := seen[view]; !ok {
			views = append(views, view)
		}
	}
	m.views = make(map[string]*RenderView)
	m.pendingDestroy = make(map[*RenderView]struct{})
	m.mutex.Unlock()
	for i := range views {
		destroyRenderViewResources(views[i], device, drawings)
	}
}

func (m *RenderViewManager) Views() []*RenderView {
	defer tracing.NewRegion("RenderViewManager.Views").End()
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	views := make([]*RenderView, 0, len(m.views))
	for _, view := range m.views {
		if view.Destroyed() {
			continue
		}
		views = append(views, view)
	}
	sort.SliceStable(views, func(i, j int) bool {
		left := views[i].Options()
		right := views[j].Options()
		if left.Sort != right.Sort {
			return left.Sort < right.Sort
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		return views[i].order < views[j].order
	})
	return views
}

func (m *RenderViewManager) FrameViews() []RenderViewFrame {
	defer tracing.NewRegion("RenderViewManager.FrameViews").End()
	views := m.Views()
	frames := make([]RenderViewFrame, 0, len(views))
	for i := range views {
		frames = append(frames, newRenderViewFrame(views[i]))
	}
	return frames
}

func (m *RenderViewManager) ensureLocked() {
	if m.views == nil {
		m.views = make(map[string]*RenderView)
	}
	if m.pendingDestroy == nil {
		m.pendingDestroy = make(map[*RenderView]struct{})
	}
}

func validateRenderViewOptions(options RenderViewOptions) error {
	if options.Name == "" {
		return errors.New("render view name is required")
	}
	if !options.ViewMode.Valid() {
		return fmt.Errorf("render view %q has invalid view mode %d", options.Name, options.ViewMode)
	}
	return nil
}
