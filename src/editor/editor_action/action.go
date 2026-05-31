/******************************************************************************/
/* action.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_action

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type ActionID string
type Source string
type UndoPolicy int

const (
	SourceUnknown Source = ""
	SourceMenu    Source = "menu"
	SourcePalette Source = "palette"
	SourceKeybind Source = "keybinding"
	SourceLua     Source = "lua"
	SourceREST    Source = "rest"
)

const (
	UndoPolicyNone UndoPolicy = iota
	UndoPolicyTransaction
	UndoPolicyManaged
)

var (
	ErrDuplicateAction = errors.New("editor_action: duplicate action")
	ErrInvalidAction   = errors.New("editor_action: invalid action")
	ErrActionNotFound  = errors.New("editor_action: action not found")
)

type KeyChord struct {
	Keys       []int `json:"keys,omitempty"`
	Ctrl       bool  `json:"ctrl,omitempty"`
	Meta       bool  `json:"meta,omitempty"`
	CtrlOrMeta bool  `json:"ctrlOrMeta,omitempty"`
	Shift      bool  `json:"shift,omitempty"`
	Alt        bool  `json:"alt,omitempty"`
}

type ActionBinding struct {
	Action    ActionID `json:"action"`
	Params    any      `json:"params,omitempty"`
	Workspace string   `json:"workspace,omitempty"`
	Enabled   bool     `json:"enabled"`
	Chord     KeyChord `json:"chord"`
}

type Parameter struct {
	Name        string   `json:"name"`
	Label       string   `json:"label,omitempty"`
	Type        string   `json:"type,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"`
	Description string   `json:"description,omitempty"`
}

type Variant struct {
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Params      any      `json:"params,omitempty"`
	Hidden      bool     `json:"hidden,omitempty"`
}

type Definition struct {
	ID                ActionID        `json:"id"`
	Label             string          `json:"label"`
	Description       string          `json:"description,omitempty"`
	Category          string          `json:"category,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	Aliases           []string        `json:"aliases,omitempty"`
	DefaultParams     any             `json:"defaultParams,omitempty"`
	NewParams         ParamsFactory   `json:"-"`
	Parameters        []Parameter     `json:"parameters,omitempty"`
	DefaultBindings   []ActionBinding `json:"defaultBindings,omitempty"`
	UndoPolicy        UndoPolicy      `json:"undoPolicy"`
	Visible           bool            `json:"visible"`
	Unbindable        bool            `json:"unbindable,omitempty"`
	RequiredWorkspace string          `json:"requiredWorkspace,omitempty"`
	Variants          []Variant       `json:"variants,omitempty"`
}

type Request struct {
	ID            ActionID `json:"id"`
	Params        any      `json:"params,omitempty"`
	Source        Source   `json:"source,omitempty"`
	CorrelationID string   `json:"correlationId,omitempty"`
}

type Result struct {
	OK                bool           `json:"ok"`
	Message           string         `json:"message,omitempty"`
	Error             string         `json:"error,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
	Warnings          []string       `json:"warnings,omitempty"`
	AffectedEntityIDs []string       `json:"affectedEntityIds,omitempty"`
	SelectedEntityIDs []string       `json:"selectedEntityIds,omitempty"`
}

type Context struct {
	CurrentWorkspace string
	InputFocused     bool
	Services         map[string]any
	Feedback         func(Result)
}

type Handler func(Context, Request) Result
type CanRunFunc func(Context, Request) Result
type ParamsFactory func() any

type Entry struct {
	Definition
	Params       any `json:"params,omitempty"`
	VariantIndex int `json:"variantIndex,omitempty"`
}

type registeredAction struct {
	def     Definition
	handler Handler
	canRun  CanRunFunc
}

type Registry struct {
	mu      sync.RWMutex
	actions map[ActionID]registeredAction
}

func NewRegistry() *Registry {
	return &Registry{actions: map[ActionID]registeredAction{}}
}

func (r *Registry) Register(def Definition, handler Handler, canRun CanRunFunc) error {
	def.ID = ActionID(strings.TrimSpace(string(def.ID)))
	def.Label = strings.TrimSpace(def.Label)
	if def.ID == "" || def.Label == "" {
		return fmt.Errorf("%w: action id and label are required", ErrInvalidAction)
	}
	if handler == nil {
		return fmt.Errorf("%w: handler is nil for %s", ErrInvalidAction, def.ID)
	}
	if canRun == nil {
		canRun = func(Context, Request) Result { return Success("") }
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.actions == nil {
		r.actions = map[ActionID]registeredAction{}
	}
	if _, exists := r.actions[def.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateAction, def.ID)
	}
	r.actions[def.ID] = registeredAction{
		def:     def,
		handler: handler,
		canRun:  canRun,
	}
	return nil
}

func (r *Registry) Definition(id ActionID) (Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	action, ok := r.actions[id]
	return action.def, ok
}

func (r *Registry) Registered(id ActionID) (registeredAction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	action, ok := r.actions[id]
	return action, ok
}

func (r *Registry) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Definition, 0, len(r.actions))
	for _, action := range r.actions {
		out = append(out, action.def)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category == out[j].Category {
			return out[i].Label < out[j].Label
		}
		return out[i].Category < out[j].Category
	})
	return out
}

func (r *Registry) Entries(ctx Context, query string) []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	query = strings.ToLower(strings.TrimSpace(query))
	out := make([]Entry, 0, len(r.actions))
	for _, action := range r.actions {
		def := action.def
		if !def.Visible {
			continue
		}
		if def.RequiredWorkspace != "" && ctx.CurrentWorkspace != "" &&
			def.RequiredWorkspace != ctx.CurrentWorkspace {
			continue
		}
		base := Entry{Definition: def, Params: def.DefaultParams}
		if matchesEntry(base, query) {
			out = append(out, base)
		}
		for i, variant := range def.Variants {
			if variant.Hidden {
				continue
			}
			entry := Entry{
				Definition:   def,
				Params:       variant.Params,
				VariantIndex: i + 1,
			}
			entry.Label = variant.Label
			if variant.Description != "" {
				entry.Description = variant.Description
			}
			entry.Tags = append(append([]string{}, def.Tags...), variant.Tags...)
			if entry.Params == nil {
				entry.Params = def.DefaultParams
			}
			if matchesEntry(entry, query) {
				out = append(out, entry)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category == out[j].Category {
			return out[i].Label < out[j].Label
		}
		return out[i].Category < out[j].Category
	})
	return out
}

func matchesEntry(entry Entry, query string) bool {
	if query == "" {
		return true
	}
	parts := []string{
		string(entry.ID),
		entry.Label,
		entry.Description,
		entry.Category,
	}
	parts = append(parts, entry.Tags...)
	parts = append(parts, entry.Aliases...)
	haystack := strings.ToLower(strings.Join(parts, " "))
	for _, token := range strings.Fields(query) {
		if !strings.Contains(haystack, token) {
			return false
		}
	}
	return true
}

type Service struct {
	registry          *Registry
	context           func() Context
	beginTransaction  func()
	commitTransaction func()
	cancelTransaction func()
	runOnMainThread   func(func())
	usageMu           sync.RWMutex
	usageSequence     int64
	usageOrder        map[ActionID]int64
}

func NewService() *Service {
	return &Service{registry: NewRegistry()}
}

func (s *Service) Registry() *Registry {
	if s.registry == nil {
		s.registry = NewRegistry()
	}
	return s.registry
}

func (s *Service) SetContextProvider(provider func() Context) {
	s.context = provider
}

func (s *Service) SetTransactionHooks(begin, commit, cancel func()) {
	s.beginTransaction = begin
	s.commitTransaction = commit
	s.cancelTransaction = cancel
}

func (s *Service) SetMainThreadScheduler(run func(func())) {
	s.runOnMainThread = run
}

func (s *Service) Register(def Definition, handler Handler, canRun CanRunFunc) error {
	return s.Registry().Register(def, handler, canRun)
}

func (s *Service) Definitions() []Definition {
	return s.Registry().Definitions()
}

func (s *Service) Search(query string) []Entry {
	ctx := s.currentContext()
	entries := s.Registry().Entries(ctx, query)
	out := entries[:0]
	for _, entry := range entries {
		action, ok := s.Registry().Registered(entry.ID)
		if !ok {
			continue
		}
		req, err := s.normalizeRequest(action, Request{ID: entry.ID, Params: entry.Params})
		if err != nil {
			continue
		}
		if result := action.canRun(ctx, req); result.OK {
			out = append(out, entry)
		}
	}
	s.sortByUsage(out)
	return out
}

func (s *Service) SearchOnMainThread(query string) []Entry {
	if s.runOnMainThread == nil {
		return s.Search(query)
	}
	result := make(chan []Entry, 1)
	s.runOnMainThread(func() {
		result <- s.Search(query)
	})
	return <-result
}

func (s *Service) DefaultBindings() []ActionBinding {
	defs := s.Definitions()
	out := make([]ActionBinding, 0)
	for _, def := range defs {
		if def.Unbindable {
			continue
		}
		for _, binding := range def.DefaultBindings {
			if binding.Action == "" {
				binding.Action = def.ID
			}
			if binding.Params == nil {
				binding.Params = def.DefaultParams
			}
			if binding.Workspace == "" {
				binding.Workspace = def.RequiredWorkspace
			}
			if !binding.Enabled {
				binding.Enabled = true
			}
			out = append(out, binding)
		}
	}
	return out
}

func (s *Service) CanRun(req Request) Result {
	action, ok := s.Registry().Registered(req.ID)
	if !ok {
		return Failure(fmt.Sprintf("action %q was not found", req.ID))
	}
	req, err := s.normalizeRequest(action, req)
	if err != nil {
		return Failure(err.Error())
	}
	return action.canRun(s.currentContext(), req)
}

func (s *Service) CanRunOnMainThread(req Request) Result {
	if s.runOnMainThread == nil {
		return s.CanRun(req)
	}
	result := make(chan Result, 1)
	s.runOnMainThread(func() {
		result <- s.CanRun(req)
	})
	return <-result
}

func (s *Service) Run(req Request) Result {
	action, ok := s.Registry().Registered(req.ID)
	if !ok {
		return Failure(fmt.Sprintf("action %q was not found", req.ID))
	}
	var err error
	req, err = s.normalizeRequest(action, req)
	if err != nil {
		return Failure(err.Error())
	}
	ctx := s.currentContext()
	if can := action.canRun(ctx, req); !can.OK {
		return can
	}
	if action.def.UndoPolicy == UndoPolicyTransaction && s.beginTransaction != nil {
		s.beginTransaction()
		result := action.handler(ctx, req)
		if result.OK {
			if s.commitTransaction != nil {
				s.commitTransaction()
			}
			s.recordUsage(req.ID)
		} else if s.cancelTransaction != nil {
			s.cancelTransaction()
		} else if s.commitTransaction != nil {
			s.commitTransaction()
		}
		s.feedback(ctx, result)
		return result
	}
	result := action.handler(ctx, req)
	if result.OK {
		s.recordUsage(req.ID)
	}
	s.feedback(ctx, result)
	return result
}

func (s *Service) RunOnMainThread(req Request) Result {
	if s.runOnMainThread == nil {
		return s.Run(req)
	}
	result := make(chan Result, 1)
	s.runOnMainThread(func() {
		result <- s.Run(req)
	})
	return <-result
}

func (s *Service) NormalizeRequest(req Request) (Request, error) {
	action, ok := s.Registry().Registered(req.ID)
	if !ok {
		return req, fmt.Errorf("action %q was not found", req.ID)
	}
	return s.normalizeRequest(action, req)
}

func (s *Service) normalizeRequest(action registeredAction, req Request) (Request, error) {
	params, err := normalizeParams(action.def, req.Params)
	if err != nil {
		return req, err
	}
	req.Params = params
	return req, nil
}

func normalizeParams(def Definition, params any) (any, error) {
	if params == nil {
		params = def.DefaultParams
	}
	if params == nil {
		return nil, nil
	}
	if def.NewParams == nil {
		return params, nil
	}
	target := def.NewParams()
	if target == nil {
		return params, nil
	}
	if sameParamType(params, target) {
		return params, nil
	}
	switch v := params.(type) {
	case json.RawMessage:
		if len(v) == 0 {
			return target, nil
		}
		if err := json.Unmarshal(v, target); err != nil {
			return nil, err
		}
		return target, nil
	case []byte:
		if len(v) == 0 {
			return target, nil
		}
		if err := json.Unmarshal(v, target); err != nil {
			return nil, err
		}
		return target, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return target, nil
		}
		if err := json.Unmarshal([]byte(v), target); err != nil {
			return nil, err
		}
		return target, nil
	default:
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(data, target); err != nil {
			return nil, err
		}
		return target, nil
	}
}

func sameParamType(params, target any) bool {
	pt, tt := reflect.TypeOf(params), reflect.TypeOf(target)
	if pt == nil || tt == nil {
		return false
	}
	if pt.AssignableTo(tt) {
		return true
	}
	if tt.Kind() == reflect.Pointer && pt.AssignableTo(tt.Elem()) {
		return true
	}
	return false
}

func (s *Service) currentContext() Context {
	if s.context == nil {
		return Context{}
	}
	return s.context()
}

func (s *Service) recordUsage(id ActionID) {
	s.usageMu.Lock()
	defer s.usageMu.Unlock()
	if s.usageOrder == nil {
		s.usageOrder = map[ActionID]int64{}
	}
	s.usageSequence++
	s.usageOrder[id] = s.usageSequence
}

func (s *Service) sortByUsage(entries []Entry) {
	s.usageMu.RLock()
	if len(s.usageOrder) == 0 {
		s.usageMu.RUnlock()
		return
	}
	order := make(map[ActionID]int64, len(s.usageOrder))
	for id, sequence := range s.usageOrder {
		order[id] = sequence
	}
	s.usageMu.RUnlock()
	sort.SliceStable(entries, func(i, j int) bool {
		iOrder := order[entries[i].ID]
		jOrder := order[entries[j].ID]
		if iOrder == jOrder {
			return false
		}
		if iOrder == 0 {
			return false
		}
		if jOrder == 0 {
			return true
		}
		return iOrder > jOrder
	})
}

func (s *Service) feedback(ctx Context, result Result) {
	if ctx.Feedback != nil {
		ctx.Feedback(result)
	}
}

func Success(message string) Result {
	return Result{OK: true, Message: message}
}

func Failure(message string) Result {
	return Result{OK: false, Message: message, Error: message}
}

func Params(value any) any {
	return value
}

func Param[T any](req Request) (T, bool) {
	var zero T
	if req.Params == nil {
		return zero, false
	}
	if value, ok := req.Params.(T); ok {
		return value, true
	}
	if ptr, ok := req.Params.(*T); ok && ptr != nil {
		return *ptr, true
	}
	data, err := json.Marshal(req.Params)
	if err != nil {
		return zero, false
	}
	var out T
	if err = json.Unmarshal(data, &out); err != nil {
		return zero, false
	}
	return out, true
}
