/******************************************************************************/
/* settings_workspace_shortcuts.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package settings_workspace

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_overlay/confirm_prompt"
	"kaijuengine.com/editor/editor_overlay/file_browser"
	"kaijuengine.com/editor/editor_overlay/input_prompt"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

const shortcutProfileVersion = 1

type shortcutSectionData struct {
	ID   string
	Name string
	Rows []shortcutRowData
}

type shortcutRowData struct {
	Action      string
	Workspace   string
	Label       string
	Description string
	Bindings    []shortcutBindingData
	Empty       bool
}

type shortcutBindingData struct {
	Index int
	Chord string
}

type shortcutProfile struct {
	Version  int                           `json:"version"`
	Bindings []editor_action.ActionBinding `json:"bindings"`
}

type shortcutCaptureState struct {
	Action      editor_action.ActionID
	Workspace   string
	BindingIdx  int
	CallbackID  hid.KeyCallbackId
	Description string
	Button      *document.Element
	Original    string
}

func (w *SettingsWorkspace) buildShortcutSections() []shortcutSectionData {
	defaults := w.editor.Actions().DefaultBindings()
	defs := w.editor.Actions().Definitions()
	sections := map[string]*shortcutSectionData{
		"": {ID: "", Name: "Global"},
	}
	workspaceNames := map[string]string{}
	for _, row := range w.buildWorkspaceRows() {
		workspaceNames[row.ID] = row.DisplayName
		if _, ok := sections[row.ID]; !ok {
			sections[row.ID] = &shortcutSectionData{ID: row.ID, Name: row.DisplayName}
		}
	}
	for _, def := range defs {
		if !shortcutDefinitionVisible(def) {
			continue
		}
		workspace := def.RequiredWorkspace
		section := sections[workspace]
		if section == nil {
			name := workspaceNames[workspace]
			if name == "" {
				name = workspace
			}
			section = &shortcutSectionData{ID: workspace, Name: name}
			sections[workspace] = section
		}
		bindings := editor_action.BindingsForAction(defaults, w.editorSettings.ActionBindings, def.ID, workspace)
		row := shortcutRowData{
			Action:      string(def.ID),
			Workspace:   workspace,
			Label:       def.Label,
			Description: def.Description,
			Bindings:    make([]shortcutBindingData, 0, len(bindings)),
			Empty:       len(bindings) == 0,
		}
		for i, binding := range bindings {
			row.Bindings = append(row.Bindings, shortcutBindingData{
				Index: i,
				Chord: editor_action.FormatKeyChord(binding.Chord),
			})
		}
		section.Rows = append(section.Rows, row)
	}
	for _, section := range sections {
		sort.Slice(section.Rows, func(i, j int) bool {
			return section.Rows[i].Label < section.Rows[j].Label
		})
	}
	out := make([]shortcutSectionData, 0, len(sections))
	out = append(out, *sections[""])
	for _, cfg := range w.editorSettings.Workspaces {
		if section := sections[cfg.ID]; section != nil && len(section.Rows) > 0 {
			out = append(out, *section)
			delete(sections, cfg.ID)
		}
	}
	rest := make([]shortcutSectionData, 0, len(sections))
	for id, section := range sections {
		if id == "" || len(section.Rows) == 0 {
			continue
		}
		rest = append(rest, *section)
	}
	sort.Slice(rest, func(i, j int) bool { return rest[i].Name < rest[j].Name })
	out = append(out, rest...)
	return out
}

func shortcutDefinitionVisible(def editor_action.Definition) bool {
	return def.Visible && !def.Unbindable
}

func (w *SettingsWorkspace) captureShortcut(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.captureShortcut").End()
	action := editor_action.ActionID(e.Attribute("data-action"))
	if action == "" {
		return
	}
	workspace := e.Attribute("data-workspace")
	idx, _ := strconv.Atoi(e.Attribute("data-index"))
	w.stopShortcutCapture()
	original := ""
	if label := e.InnerLabel(); label != nil {
		original = label.Text()
		label.SetText("Press a key")
	}
	w.setShortcutCaptureButtonActive(e, true)
	w.editor.BlurInterface()
	state := shortcutCaptureState{
		Action:      action,
		Workspace:   workspace,
		BindingIdx:  idx,
		Description: actionLabel(w.editor.Actions().Definitions(), action),
		Button:      e,
		Original:    original,
	}
	state.CallbackID = w.Host.Window.Keyboard.AddKeyCallback(func(keyID int, keyState hid.KeyState) {
		w.onShortcutCaptureKey(keyID, keyState)
	})
	w.shortcutCapture = &state
}

func (w *SettingsWorkspace) onShortcutCaptureKey(keyID int, keyState hid.KeyState) {
	if w.shortcutCapture == nil {
		return
	}
	if keyState == hid.KeyStateUp {
		if editor_action.IsModifierKey(keyID) {
			w.updateShortcutCapturePreview(0)
		}
		return
	}
	if keyState != hid.KeyStateDown && keyState != hid.KeyStatePressedAndReleased {
		return
	}
	if hid.KeyboardKey(keyID) == hid.KeyboardKeyEscape {
		w.cancelShortcutCapture()
		return
	}
	if editor_action.IsModifierKey(keyID) {
		w.updateShortcutCapturePreview(0)
		return
	}
	kb := &w.Host.Window.Keyboard
	chord := editor_action.KeyChord{
		Keys:       []int{keyID},
		CtrlOrMeta: kb.HasCtrl() || kb.HasMeta(),
		Shift:      kb.HasShift(),
		Alt:        kb.HasAlt(),
	}
	w.updateShortcutCapturePreview(keyID)
	state := w.acceptShortcutCapture()
	w.applyCapturedShortcut(state, chord)
}

func (w *SettingsWorkspace) applyCapturedShortcut(state shortcutCaptureState, chord editor_action.KeyChord) {
	if !editor_action.ValidChord(chord) {
		w.reloadKeyboardSettings()
		return
	}
	replacements := w.replacementBindings(state, chord)
	candidate := editor_action.ActionBinding{
		Action:    state.Action,
		Workspace: state.Workspace,
		Enabled:   true,
		Chord:     chord,
	}
	defaults := w.editor.Actions().DefaultBindings()
	active := editor_action.EffectiveBindings(defaults, w.editorSettings.ActionBindings)
	conflicts := editor_action.BindingConflicts(active, candidate, w.actionLabels())
	if len(conflicts) == 0 {
		w.applyShortcutReplacement(state.Action, state.Workspace, replacements, nil)
		return
	}
	w.editor.BlurInterface()
	confirm_prompt.Show(w.Host, confirm_prompt.Config{
		Title:       "Replace shortcut",
		Description: w.shortcutConflictMessage(chord, conflicts),
		ConfirmText: "Replace",
		CancelText:  "Cancel",
		OnConfirm: func() {
			w.editor.FocusInterface()
			w.applyShortcutReplacement(state.Action, state.Workspace, replacements, conflicts)
		},
		OnCancel: func() {
			w.editor.FocusInterface()
			w.reloadKeyboardSettings()
		},
	})
}

func (w *SettingsWorkspace) replacementBindings(state shortcutCaptureState, chord editor_action.KeyChord) []editor_action.ActionBinding {
	current := editor_action.BindingsForAction(w.editor.Actions().DefaultBindings(),
		w.editorSettings.ActionBindings, state.Action, state.Workspace)
	replacements := make([]editor_action.ActionBinding, 0, len(current)+1)
	for i, binding := range current {
		binding.Action = state.Action
		binding.Workspace = state.Workspace
		binding.Enabled = true
		if i == state.BindingIdx {
			binding.Chord = chord
		}
		if !containsChord(replacements, binding.Chord) {
			replacements = append(replacements, binding)
		}
	}
	if state.BindingIdx < 0 || state.BindingIdx >= len(current) {
		binding := editor_action.ActionBinding{
			Action:    state.Action,
			Workspace: state.Workspace,
			Enabled:   true,
			Chord:     chord,
		}
		if !containsChord(replacements, chord) {
			replacements = append(replacements, binding)
		}
	}
	return replacements
}

func (w *SettingsWorkspace) clearShortcut(e *document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.clearShortcut").End()
	action := editor_action.ActionID(e.Attribute("data-action"))
	if action == "" {
		return
	}
	w.applyShortcutReplacement(action, e.Attribute("data-workspace"), nil, nil)
}

func (w *SettingsWorkspace) applyShortcutReplacement(action editor_action.ActionID, workspace string,
	replacements []editor_action.ActionBinding, conflicts []editor_action.BindingConflict) {
	defaults := w.editor.Actions().DefaultBindings()
	settings := editor_action.RemoveBindingConflicts(defaults, w.editorSettings.ActionBindings, conflicts)
	settings = editor_action.ReplaceActionBindings(settings, action, workspace, replacements)
	if len(replacements) == 0 {
		settings = append(settings, editor_action.ActionBinding{
			Action:    action,
			Workspace: workspace,
			Enabled:   false,
		})
	}
	w.editorSettings.ActionBindings = settings
	w.persistShortcutSettings()
	w.reloadKeyboardSettings()
}

func (w *SettingsWorkspace) resetShortcuts(*document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.resetShortcuts").End()
	w.editor.BlurInterface()
	confirm_prompt.Show(w.Host, confirm_prompt.Config{
		Title:       "Reset keyboard shortcuts",
		Description: "Reset all keyboard shortcuts to their defaults?",
		ConfirmText: "Reset",
		CancelText:  "Cancel",
		OnConfirm: func() {
			w.editor.FocusInterface()
			w.editorSettings.ActionBindings = nil
			w.persistShortcutSettings()
			w.reloadKeyboardSettings()
		},
		OnCancel: w.editor.FocusInterface,
	})
}

func (w *SettingsWorkspace) exportShortcuts(*document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.exportShortcuts").End()
	w.editor.BlurInterface()
	start, _ := filesystem.GameDirectory()
	file_browser.Show(w.Host, file_browser.Config{
		Title:        "Export keyboard shortcuts",
		StartingPath: start,
		OnlyFolders:  true,
		OnConfirm: func(paths []string) {
			input_prompt.Show(w.Host, input_prompt.Config{
				Title:       "Export keyboard shortcuts",
				Description: "Choose a file name for the shortcut profile.",
				Placeholder: "File name",
				Value:       "kaiju-shortcuts.json",
				ConfirmText: "Export",
				CancelText:  "Cancel",
				OnConfirm: func(name string) {
					w.editor.FocusInterface()
					w.writeShortcutProfile(paths[0], name)
				},
				OnCancel: w.editor.FocusInterface,
			})
		},
		OnCancel: w.editor.FocusInterface,
	})
}

func (w *SettingsWorkspace) importShortcuts(*document.Element) {
	defer tracing.NewRegion("SettingsWorkspace.importShortcuts").End()
	w.editor.BlurInterface()
	start, _ := filesystem.GameDirectory()
	file_browser.Show(w.Host, file_browser.Config{
		Title:        "Import keyboard shortcuts",
		StartingPath: start,
		ExtFilter:    []string{".json"},
		OnlyFiles:    true,
		OnConfirm: func(paths []string) {
			if len(paths) == 0 {
				w.editor.FocusInterface()
				return
			}
			if err := w.readShortcutProfile(paths[0]); err != nil {
				slog.Error("failed to import keyboard shortcuts", "error", err)
			}
			w.editor.FocusInterface()
		},
		OnCancel: w.editor.FocusInterface,
	})
}

func (w *SettingsWorkspace) writeShortcutProfile(folder, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "kaiju-shortcuts.json"
	}
	if filepath.Ext(name) == "" {
		name += ".json"
	}
	profile := shortcutProfile{
		Version:  shortcutProfileVersion,
		Bindings: w.shortcutProfileBindings(),
	}
	data, err := json.MarshalIndent(profile, "", "\t")
	if err != nil {
		slog.Error("failed to encode keyboard shortcuts", "error", err)
		return
	}
	if err = os.WriteFile(filepath.Join(folder, name), data, os.ModePerm); err != nil {
		slog.Error("failed to write keyboard shortcuts", "error", err)
	}
}

func (w *SettingsWorkspace) shortcutProfileBindings() []editor_action.ActionBinding {
	defaults := w.editor.Actions().DefaultBindings()
	defs := w.editor.Actions().Definitions()
	out := make([]editor_action.ActionBinding, 0, len(defs))
	for _, def := range defs {
		if !shortcutDefinitionVisible(def) {
			continue
		}
		bindings := editor_action.BindingsForAction(defaults, w.editorSettings.ActionBindings,
			def.ID, def.RequiredWorkspace)
		if len(bindings) == 0 {
			out = append(out, editor_action.ActionBinding{
				Action:    def.ID,
				Workspace: def.RequiredWorkspace,
				Enabled:   false,
			})
			continue
		}
		for _, binding := range bindings {
			binding.Action = def.ID
			binding.Workspace = def.RequiredWorkspace
			binding.Enabled = true
			out = append(out, binding)
		}
	}
	return out
}

func (w *SettingsWorkspace) readShortcutProfile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	profile := shortcutProfile{}
	if err = json.Unmarshal(data, &profile); err != nil {
		return err
	}
	if profile.Version != shortcutProfileVersion {
		return fmt.Errorf("unsupported shortcut profile version %d", profile.Version)
	}
	bindings, err := w.validateShortcutBindings(profile.Bindings)
	if err != nil {
		return err
	}
	w.editorSettings.ActionBindings = bindings
	w.persistShortcutSettings()
	w.reloadKeyboardSettings()
	return nil
}

func (w *SettingsWorkspace) validateShortcutBindings(bindings []editor_action.ActionBinding) ([]editor_action.ActionBinding, error) {
	defs := w.actionDefinitions()
	active := make([]editor_action.ActionBinding, 0, len(bindings))
	normalized := make([]editor_action.ActionBinding, 0, len(bindings))
	for _, binding := range bindings {
		def, ok := defs[binding.Action]
		if !ok || !shortcutDefinitionVisible(def) {
			return nil, fmt.Errorf("unknown bindable action %q", binding.Action)
		}
		if binding.Workspace == "" {
			binding.Workspace = def.RequiredWorkspace
		}
		if binding.Workspace != def.RequiredWorkspace {
			return nil, fmt.Errorf("action %q belongs to workspace %q", binding.Action, def.RequiredWorkspace)
		}
		normalized = append(normalized, binding)
		if !binding.Enabled {
			continue
		}
		if !editor_action.ValidChord(binding.Chord) {
			return nil, fmt.Errorf("action %q has an invalid chord", binding.Action)
		}
		for _, existing := range active {
			if !editor_action.ChordsEqual(existing.Chord, binding.Chord) {
				continue
			}
			if existing.Action == binding.Action && existing.Workspace == binding.Workspace {
				return nil, fmt.Errorf("action %q has duplicate chord %s", binding.Action,
					editor_action.FormatKeyChord(binding.Chord))
			}
			conflicts := editor_action.BindingConflicts([]editor_action.ActionBinding{existing},
				binding, w.actionLabels())
			if len(conflicts) > 0 {
				return nil, fmt.Errorf("shortcut %s conflicts between %q and %q",
					editor_action.FormatKeyChord(binding.Chord), existing.Action, binding.Action)
			}
		}
		active = append(active, binding)
	}
	return normalized, nil
}

func (w *SettingsWorkspace) persistShortcutSettings() {
	if err := w.editorSettings.Save(); err != nil {
		slog.Error("failed to persist keyboard shortcut settings", "error", err)
	}
}

func (w *SettingsWorkspace) reloadKeyboardSettings() {
	scrollY := w.keyboardSettingsScrollY()
	w.Host.RunNextFrame(func() {
		w.ReloadUI(uiFile, w.uiData(), w.funcMap())
		w.reloadedUI()
		w.CommonOpen()
		w.projectSettingsBox.UI.Hide()
		w.editorSettingsBox.UI.Hide()
		w.pluginSettingsBox.UI.Hide()
		w.workspaceSettingsBox.UI.Hide()
		w.keyboardSettingsBox.UI.Show()
		for _, e := range w.Doc.GetElementsByClass("edPanelBgHoverable") {
			if e.InnerLabel().Text() == "Keyboard Shortcuts" {
				w.Doc.SetElementClasses(e, "edPanelBgHoverable", "selected")
				break
			}
		}
		w.Host.RunAfterNextUIClean(func() {
			w.restoreKeyboardSettingsScrollY(scrollY)
		})
	})
}

func (w *SettingsWorkspace) keyboardSettingsScrollY() float32 {
	if w.keyboardSettingsBox == nil {
		return 0
	}
	parent := w.keyboardSettingsBox.Parent.Value()
	if parent == nil || parent.UIPanel == nil {
		return 0
	}
	return parent.UIPanel.ScrollY()
}

func (w *SettingsWorkspace) restoreKeyboardSettingsScrollY(scrollY float32) {
	if w.keyboardSettingsBox == nil {
		return
	}
	parent := w.keyboardSettingsBox.Parent.Value()
	if parent == nil || parent.UIPanel == nil {
		return
	}
	parent.UIPanel.SetScrollY(scrollY)
}

func (w *SettingsWorkspace) stopShortcutCapture() {
	w.stopShortcutCaptureInternal(true)
}

func (w *SettingsWorkspace) acceptShortcutCapture() shortcutCaptureState {
	state := *w.shortcutCapture
	w.stopShortcutCaptureInternal(false)
	return state
}

func (w *SettingsWorkspace) stopShortcutCaptureInternal(restoreButton bool) {
	if w.shortcutCapture == nil {
		return
	}
	state := w.shortcutCapture
	w.Host.Window.Keyboard.RemoveKeyCallback(state.CallbackID)
	if restoreButton {
		w.restoreShortcutCaptureButton(state)
	}
	w.shortcutCapture = nil
	w.editor.FocusInterface()
}

func (w *SettingsWorkspace) cancelShortcutCapture() {
	w.stopShortcutCapture()
	w.reloadKeyboardSettings()
}

func (w *SettingsWorkspace) updateShortcutCapturePreview(keyID int) {
	if w.shortcutCapture == nil || w.shortcutCapture.Button == nil {
		return
	}
	if label := w.shortcutCapture.Button.InnerLabel(); label != nil {
		label.SetText(shortcutCapturePreview(&w.Host.Window.Keyboard, keyID))
	}
}

func shortcutCapturePreview(kb *hid.Keyboard, keyID int) string {
	parts := make([]string, 0, 4)
	if kb.HasCtrl() || kb.HasMeta() {
		parts = append(parts, "Ctrl/Cmd")
	}
	if kb.HasShift() {
		parts = append(parts, "Shift")
	}
	if kb.HasAlt() {
		parts = append(parts, "Alt")
	}
	hasPrimaryKey := keyID > int(hid.KeyBoardKeyInvalid) && !editor_action.IsModifierKey(keyID)
	if hasPrimaryKey {
		parts = append(parts, editor_action.KeyName(keyID))
	}
	if len(parts) == 0 {
		return "Press a key"
	}
	preview := strings.Join(parts, "+")
	if !hasPrimaryKey {
		preview += "+"
	}
	return preview
}

func (w *SettingsWorkspace) restoreShortcutCaptureButton(state *shortcutCaptureState) {
	if state == nil || state.Button == nil {
		return
	}
	if label := state.Button.InnerLabel(); label != nil {
		label.SetText(state.Original)
	}
	w.setShortcutCaptureButtonActive(state.Button, false)
}

func (w *SettingsWorkspace) setShortcutCaptureButtonActive(button *document.Element, active bool) {
	if w.Doc == nil || button == nil {
		return
	}
	w.Doc.SetElementClasses(button, shortcutCaptureClasses(button, active)...)
}

func shortcutCaptureClasses(button *document.Element, active bool) []string {
	classes := make([]string, 0, len(button.ClassList())+1)
	hasActive := false
	for _, class := range button.ClassList() {
		class = strings.TrimSpace(class)
		if class == "" {
			continue
		}
		if class == "shortcutCaptureActive" {
			hasActive = true
			if !active {
				continue
			}
		}
		classes = append(classes, class)
	}
	if active && !hasActive {
		classes = append(classes, "shortcutCaptureActive")
	}
	return classes
}

func (w *SettingsWorkspace) shortcutConflictMessage(chord editor_action.KeyChord, conflicts []editor_action.BindingConflict) string {
	names := make([]string, 0, len(conflicts))
	for _, conflict := range conflicts {
		names = append(names, conflict.Label)
	}
	sort.Strings(names)
	return fmt.Sprintf("%s is already bound to %s. Replace the existing binding?",
		editor_action.FormatKeyChord(chord), strings.Join(names, ", "))
}

func (w *SettingsWorkspace) actionDefinitions() map[editor_action.ActionID]editor_action.Definition {
	defs := w.editor.Actions().Definitions()
	out := make(map[editor_action.ActionID]editor_action.Definition, len(defs))
	for _, def := range defs {
		out[def.ID] = def
	}
	return out
}

func (w *SettingsWorkspace) actionLabels() map[editor_action.ActionID]string {
	defs := w.editor.Actions().Definitions()
	out := make(map[editor_action.ActionID]string, len(defs))
	for _, def := range defs {
		out[def.ID] = def.Label
	}
	return out
}

func actionLabel(defs []editor_action.Definition, action editor_action.ActionID) string {
	for _, def := range defs {
		if def.ID == action {
			return def.Label
		}
	}
	return string(action)
}

func containsChord(bindings []editor_action.ActionBinding, chord editor_action.KeyChord) bool {
	for _, binding := range bindings {
		if editor_action.ChordsEqual(binding.Chord, chord) {
			return true
		}
	}
	return false
}
