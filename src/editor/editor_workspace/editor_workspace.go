/******************************************************************************/
/* editor_workspace.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_workspace

import (
	"kaijuengine.com/editor/editor_workspace/common_workspace"
)

// Workspace is the contract every workspace (built-in or plugin) implements.
//
// Identity:
//   - ID returns a stable unique key (e.g. "stage", "com.foo.my_plugin"). It
//     is used by the editor's workspace state machine, by settings persistence,
//     and as the lookup key for cross-workspace event subscribers and typed
//     service queries.
//   - DisplayName returns the label shown on the menu bar tab.
//   - IsRequired marks a workspace that the editor must keep enabled.
//     Stage and Settings return true so the user always has a valid landing
//     surface and a way to re-enable other workspaces. Required workspaces
//     can still be hidden from the tab bar.
//
// Lifecycle:
//   - Initialize is called once after registration, when the editor has
//     decided this workspace is enabled. The shared editor interface gives
//     the workspace access to the host, project, settings, events, the stage
//     view, sibling workspaces, and the workspace switching API.
//   - Shutdown is called when a workspace is disabled at runtime via the
//     Workspaces settings panel. Implementations should drop subscriptions,
//     release UI documents, and clear any other resources tied to the
//     editor session.
//   - Open / Close mark the workspace as the active one — the editor calls
//     them when switching tabs.
//   - Focus / Blur signal whether the editor as a whole has input focus.
//   - Update is called every frame while the workspace is current.
type Workspace interface {
	ID() string
	DisplayName() string
	IsRequired() bool

	Initialize(ed WorkspaceEditorInterface) error
	Shutdown()

	Open()
	Close()
	Focus()
	Blur()
	Hotkeys() []common_workspace.HotKey
	Update(deltaTime float64)
	IsFocusedOnInput() bool
}
