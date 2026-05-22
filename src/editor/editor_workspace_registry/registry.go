/******************************************************************************/
/* registry.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

// Package editor_workspace_registry holds the global registry of workspaces.
// Workspace packages (built-in and plugin) call Register from their init() to
// make themselves discoverable. The editor reads the registry during startup
// to decide which workspaces to initialize, in what order, and which tabs to
// surface in the menu bar.
//
// The registry is split into its own package (rather than living in editor or
// editor_workspace) so that workspace packages can register themselves without
// creating an import cycle with the editor package.
package editor_workspace_registry

import (
	"log/slog"

	"kaijuengine.com/editor/editor_workspace"
)

var (
	registry          = map[string]editor_workspace.Workspace{}
	registrationOrder = []string{}
)

// Register adds a workspace to the registry. It is intended to be called from
// the package init() of every workspace (built-in or plugin). Duplicate IDs
// are logged and ignored.
func Register(w editor_workspace.Workspace) {
	id := w.ID()
	if id == "" {
		slog.Error("attempted to register a workspace with an empty ID")
		return
	}
	if _, exists := registry[id]; exists {
		slog.Error("a workspace with the given id is already registered", "id", id)
		return
	}
	registry[id] = w
	registrationOrder = append(registrationOrder, id)
}

// Get returns the workspace registered under the given id, or false.
func Get(id string) (editor_workspace.Workspace, bool) {
	w, ok := registry[id]
	return w, ok
}

// All returns every registered workspace in the order they were registered.
// The editor's reconciliation logic uses this order as the default for
// workspaces that have no entry in persisted settings yet.
func All() []editor_workspace.Workspace {
	out := make([]editor_workspace.Workspace, 0, len(registrationOrder))
	for _, id := range registrationOrder {
		out = append(out, registry[id])
	}
	return out
}

// IDs returns every registered workspace id in registration order.
func IDs() []string {
	out := make([]string, len(registrationOrder))
	copy(out, registrationOrder)
	return out
}
