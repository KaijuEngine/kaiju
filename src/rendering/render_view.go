/******************************************************************************/
/* render_view.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"kaijuengine.com/platform/profiler/tracing"
)

const DefaultRenderViewName = "default"

type RenderViewMode int

const (
	RenderViewModeDefault RenderViewMode = iota
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
	options RenderViewOptions
	order   uint64
	mutex   sync.RWMutex
}

func newRenderView(options RenderViewOptions, order uint64) *RenderView {
	options.LayerMask = normalizeRenderLayerMask(options.LayerMask)
	return &RenderView{
		options: options,
		order:   order,
	}
}

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
	v.options.Camera = camera
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

func (v *RenderView) MatchesDrawing(drawing *Drawing) bool {
	if drawing == nil {
		return false
	}
	return drawing.MatchesLayer(v.LayerMask())
}

func (v *RenderView) MatchesGroup(group *DrawInstanceGroup) bool {
	if group == nil {
		return false
	}
	return group.MatchesLayer(v.LayerMask())
}

func (v *RenderView) setOptions(options RenderViewOptions) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	options.LayerMask = normalizeRenderLayerMask(options.LayerMask)
	v.options = options
}

type RenderViewManager struct {
	views     map[string]*RenderView
	nextOrder uint64
	mutex     sync.RWMutex
}

func NewRenderViewManager(defaultOptions ...RenderViewOptions) RenderViewManager {
	options := RenderViewOptions{
		Name:      DefaultRenderViewName,
		LayerMask: RenderLayerWorld,
		Clear:     true,
	}
	if len(defaultOptions) > 0 {
		options = defaultOptions[0]
		if options.Name == "" {
			options.Name = DefaultRenderViewName
		}
	}
	manager := RenderViewManager{
		views: make(map[string]*RenderView),
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
	if _, ok := m.views[name]; !ok {
		return fmt.Errorf("render view %q not found", name)
	}
	delete(m.views, name)
	return nil
}

func (m *RenderViewManager) Views() []*RenderView {
	defer tracing.NewRegion("RenderViewManager.Views").End()
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	views := make([]*RenderView, 0, len(m.views))
	for _, view := range m.views {
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

func (m *RenderViewManager) ensureLocked() {
	if m.views == nil {
		m.views = make(map[string]*RenderView)
	}
}

func validateRenderViewOptions(options RenderViewOptions) error {
	if options.Name == "" {
		return errors.New("render view name is required")
	}
	return nil
}
