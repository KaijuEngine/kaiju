/******************************************************************************/
/* editor_action_keybindings.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/platform/hid"
)

func (ed *Editor) processActionPaletteShortcut(kb *hid.Keyboard) bool {
	if ed.IsInputFocused() {
		ed.actionPaletteKey.pending = false
		ed.actionPaletteKey.moved = false
		return false
	}
	if kb.KeyDown(hid.KeyboardKeySpace) && !kb.HasModifier() {
		ed.actionPaletteKey.pending = true
		ed.actionPaletteKey.moved = false
		return false
	}
	if ed.actionPaletteKey.pending {
		mouse := &ed.host.Window.Mouse
		if mouse.Moved() || mouse.Held(hid.MouseButtonLeft) ||
			mouse.Held(hid.MouseButtonMiddle) || mouse.Held(hid.MouseButtonRight) {
			ed.actionPaletteKey.moved = true
		}
		if kb.KeyUp(hid.KeyboardKeySpace) {
			open := !ed.actionPaletteKey.moved && !kb.HasModifier()
			ed.actionPaletteKey.pending = false
			ed.actionPaletteKey.moved = false
			if open {
				ed.Actions().Run(editor_action.Request{
					ID:     ActionEditorOpenPalette,
					Source: editor_action.SourceKeybind,
				})
				return true
			}
		}
	}
	return false
}

func (ed *Editor) processActionKeyBindings(kb *hid.Keyboard) bool {
	if ed.IsInputFocused() {
		return false
	}
	for _, binding := range ed.actionBindings() {
		if !binding.Enabled || binding.Action == ActionEditorOpenPalette {
			continue
		}
		if binding.Workspace != "" {
			if ed.currentWorkspace == nil || ed.currentWorkspace.ID() != binding.Workspace {
				continue
			}
		}
		if !actionBindingMatches(kb, binding.Chord) {
			continue
		}
		result := ed.Actions().Run(editor_action.Request{
			ID:     binding.Action,
			Params: binding.Params,
			Source: editor_action.SourceKeybind,
		})
		return result.OK
	}
	return false
}

func (ed *Editor) actionBindings() []editor_action.ActionBinding {
	return editor_action.EffectiveBindings(ed.Actions().DefaultBindings(), ed.settings.ActionBindings)
}

func actionBindingMatches(kb *hid.Keyboard, chord editor_action.KeyChord) bool {
	if len(chord.Keys) == 0 {
		return false
	}
	hasModifierRequirement := chord.Ctrl || chord.Meta || chord.CtrlOrMeta ||
		chord.Shift || chord.Alt
	if chord.Ctrl && !kb.HasCtrl() {
		return false
	}
	if chord.Meta && !kb.HasMeta() {
		return false
	}
	if chord.CtrlOrMeta && !kb.HasCtrlOrMeta() {
		return false
	}
	if chord.Shift && !kb.HasShift() {
		return false
	}
	if chord.Alt && !kb.HasAlt() {
		return false
	}
	if !hasModifierRequirement && kb.HasModifier() {
		return false
	}
	if !chord.Shift && kb.HasShift() {
		return false
	}
	if !chord.Alt && kb.HasAlt() {
		return false
	}
	if !chord.Ctrl && !chord.CtrlOrMeta && kb.HasCtrl() {
		return false
	}
	if !chord.Meta && !chord.CtrlOrMeta && kb.HasMeta() {
		return false
	}
	down := false
	for _, key := range chord.Keys {
		k := hid.KeyboardKey(key)
		if !kb.KeyHeld(k) {
			return false
		}
		down = down || kb.KeyDown(k)
	}
	return down
}
