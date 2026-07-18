/******************************************************************************/
/* frame_graph.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"errors"
	"fmt"
	"slices"
)

// FrameGraphResource identifies a logical resource used by one or more frame
// graph passes. The zero value is invalid.
type FrameGraphResource uint32

// FrameGraphPassID identifies a pass in a compiled frame graph. The zero value
// is invalid.
type FrameGraphPassID uint32

type FrameGraphResourceKind uint8

const (
	FrameGraphResourceImage FrameGraphResourceKind = iota
	FrameGraphResourceBuffer
)

type FrameGraphQueue uint8

const (
	FrameGraphQueueGraphics FrameGraphQueue = iota
	FrameGraphQueueCompute
)

type FrameGraphAccess uint8

const (
	FrameGraphAccessRead FrameGraphAccess = iota
	FrameGraphAccessWrite
	FrameGraphAccessReadWrite
)

func (a FrameGraphAccess) reads() bool {
	return a == FrameGraphAccessRead || a == FrameGraphAccessReadWrite
}

func (a FrameGraphAccess) writes() bool {
	return a == FrameGraphAccessWrite || a == FrameGraphAccessReadWrite
}

type FrameGraphResourceDescription struct {
	Name       string
	Kind       FrameGraphResourceKind
	Imported   bool
	Persistent bool
	PerView    bool
}

type FrameGraphResourceUse struct {
	Resource FrameGraphResource
	Access   FrameGraphAccess
}

// FrameGraphExecutionContext is populated by the renderer when a compiled
// pass is executed. Values is intentionally provider-neutral so a pass can
// carry backend state without leaking it into the graph scheduler.
type FrameGraphExecutionContext struct {
	Values     map[string]any
	BeforePass func(FrameGraphScheduledPass) any
	AfterPass  func(FrameGraphScheduledPass, any)
}

type FrameGraphExecute func(*FrameGraphExecutionContext) error

type FrameGraphPassDescription struct {
	Name    string
	Queue   FrameGraphQueue
	Uses    []FrameGraphResourceUse
	Execute FrameGraphExecute
}

type frameGraphResourceRecord struct {
	id          FrameGraphResource
	description FrameGraphResourceDescription
}

type frameGraphPassRecord struct {
	id          FrameGraphPassID
	index       int
	description FrameGraphPassDescription
}

// FrameGraph collects logical resources and GPU work for a single frame. It
// is deliberately separate from the editor's material Render Graph.
type FrameGraph struct {
	resources   []frameGraphResourceRecord
	passes      []frameGraphPassRecord
	resourceIDs map[string]FrameGraphResource
}

func NewFrameGraph() *FrameGraph {
	return &FrameGraph{resourceIDs: make(map[string]FrameGraphResource)}
}

func (g *FrameGraph) AddResource(description FrameGraphResourceDescription) (FrameGraphResource, error) {
	if g == nil {
		return 0, errors.New("frame graph is nil")
	}
	if description.Name == "" {
		return 0, errors.New("frame graph resource name is empty")
	}
	if _, exists := g.resourceIDs[description.Name]; exists {
		return 0, fmt.Errorf("frame graph resource %q already exists", description.Name)
	}
	id := FrameGraphResource(len(g.resources) + 1)
	g.resources = append(g.resources, frameGraphResourceRecord{id: id, description: description})
	g.resourceIDs[description.Name] = id
	return id, nil
}

func (g *FrameGraph) Resource(name string) (FrameGraphResource, bool) {
	if g == nil {
		return 0, false
	}
	id, ok := g.resourceIDs[name]
	return id, ok
}

func (g *FrameGraph) AddPass(description FrameGraphPassDescription) (FrameGraphPassID, error) {
	if g == nil {
		return 0, errors.New("frame graph is nil")
	}
	if description.Name == "" {
		return 0, errors.New("frame graph pass name is empty")
	}
	seen := make(map[FrameGraphResource]FrameGraphAccess, len(description.Uses))
	for i := range description.Uses {
		use := description.Uses[i]
		if use.Resource == 0 || int(use.Resource) > len(g.resources) {
			return 0, fmt.Errorf("frame graph pass %q references unknown resource %d", description.Name, use.Resource)
		}
		if previous, exists := seen[use.Resource]; exists {
			if previous != use.Access {
				return 0, fmt.Errorf("frame graph pass %q declares conflicting access for resource %d", description.Name, use.Resource)
			}
			return 0, fmt.Errorf("frame graph pass %q declares resource %d more than once", description.Name, use.Resource)
		}
		seen[use.Resource] = use.Access
	}
	description.Uses = slices.Clone(description.Uses)
	id := FrameGraphPassID(len(g.passes) + 1)
	g.passes = append(g.passes, frameGraphPassRecord{id: id, index: len(g.passes), description: description})
	return id, nil
}

type FrameGraphScheduledPass struct {
	ID           FrameGraphPassID
	Name         string
	Queue        FrameGraphQueue
	Uses         []FrameGraphResourceUse
	Dependencies []FrameGraphPassID
	Execute      FrameGraphExecute
}

type FrameGraphSchedule struct {
	Passes []FrameGraphScheduledPass
}

// Compile derives resource hazards and returns a deterministic topological
// schedule. Read-after-write, write-after-read, and write-after-write hazards
// are all preserved.
func (g *FrameGraph) Compile() (FrameGraphSchedule, error) {
	if g == nil {
		return FrameGraphSchedule{}, errors.New("frame graph is nil")
	}
	dependencies := make([]map[int]struct{}, len(g.passes))
	lastWriter := make(map[FrameGraphResource]int)
	readers := make(map[FrameGraphResource]map[int]struct{})
	for i := range g.passes {
		dependencies[i] = make(map[int]struct{})
		for _, use := range g.passes[i].description.Uses {
			if use.Access.reads() {
				if writer, ok := lastWriter[use.Resource]; ok {
					dependencies[i][writer] = struct{}{}
				}
				if readers[use.Resource] == nil {
					readers[use.Resource] = make(map[int]struct{})
				}
				readers[use.Resource][i] = struct{}{}
			}
			if use.Access.writes() {
				if writer, ok := lastWriter[use.Resource]; ok {
					dependencies[i][writer] = struct{}{}
				}
				for reader := range readers[use.Resource] {
					if reader != i {
						dependencies[i][reader] = struct{}{}
					}
				}
				readers[use.Resource] = make(map[int]struct{})
				lastWriter[use.Resource] = i
			}
		}
	}

	indegree := make([]int, len(g.passes))
	dependents := make([][]int, len(g.passes))
	for pass := range dependencies {
		indegree[pass] = len(dependencies[pass])
		for dependency := range dependencies[pass] {
			dependents[dependency] = append(dependents[dependency], pass)
		}
	}
	ready := make([]int, 0, len(g.passes))
	for i := range indegree {
		if indegree[i] == 0 {
			ready = append(ready, i)
		}
	}
	slices.Sort(ready)
	order := make([]int, 0, len(g.passes))
	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]
		order = append(order, current)
		slices.Sort(dependents[current])
		for _, dependent := range dependents[current] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
				slices.Sort(ready)
			}
		}
	}
	if len(order) != len(g.passes) {
		return FrameGraphSchedule{}, errors.New("frame graph contains a dependency cycle")
	}

	schedule := FrameGraphSchedule{Passes: make([]FrameGraphScheduledPass, 0, len(order))}
	for _, passIndex := range order {
		record := g.passes[passIndex]
		deps := make([]FrameGraphPassID, 0, len(dependencies[passIndex]))
		for dependency := range dependencies[passIndex] {
			deps = append(deps, g.passes[dependency].id)
		}
		slices.Sort(deps)
		schedule.Passes = append(schedule.Passes, FrameGraphScheduledPass{
			ID:           record.id,
			Name:         record.description.Name,
			Queue:        record.description.Queue,
			Uses:         slices.Clone(record.description.Uses),
			Dependencies: deps,
			Execute:      record.description.Execute,
		})
	}
	return schedule, nil
}

func (s FrameGraphSchedule) Execute(context *FrameGraphExecutionContext) error {
	if context == nil {
		context = &FrameGraphExecutionContext{}
	}
	if context.Values == nil {
		context.Values = make(map[string]any)
	}
	for i := range s.Passes {
		var timingToken any
		if context.BeforePass != nil {
			timingToken = context.BeforePass(s.Passes[i])
		}
		if s.Passes[i].Execute == nil {
			if context.AfterPass != nil {
				context.AfterPass(s.Passes[i], timingToken)
			}
			continue
		}
		err := s.Passes[i].Execute(context)
		if context.AfterPass != nil {
			context.AfterPass(s.Passes[i], timingToken)
		}
		if err != nil {
			return fmt.Errorf("frame graph pass %q failed: %w", s.Passes[i].Name, err)
		}
	}
	return nil
}
