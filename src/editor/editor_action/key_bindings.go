/******************************************************************************/
/* key_bindings.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_action

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"kaijuengine.com/platform/hid"
)

type BindingConflict struct {
	Binding ActionBinding
	Label   string
}

func EffectiveBindings(defaults, user []ActionBinding) []ActionBinding {
	defaultGroups := map[string][]ActionBinding{}
	defaultOrder := make([]string, 0)
	for _, binding := range defaults {
		key := bindingIdentity(binding)
		if _, ok := defaultGroups[key]; !ok {
			defaultOrder = append(defaultOrder, key)
		}
		defaultGroups[key] = append(defaultGroups[key], binding)
	}
	userGroups := map[string][]ActionBinding{}
	userOrder := make([]string, 0)
	for _, binding := range user {
		key := bindingIdentity(binding)
		if _, ok := userGroups[key]; !ok {
			userOrder = append(userOrder, key)
		}
		userGroups[key] = append(userGroups[key], binding)
	}
	out := make([]ActionBinding, 0, len(defaults)+len(user))
	seen := map[string]bool{}
	for _, key := range defaultOrder {
		seen[key] = true
		if overrides, ok := userGroups[key]; ok {
			out = append(out, enabledBindings(overrides)...)
			continue
		}
		out = append(out, enabledBindings(defaultGroups[key])...)
	}
	for _, key := range userOrder {
		if seen[key] {
			continue
		}
		out = append(out, enabledBindings(userGroups[key])...)
	}
	return out
}

func BindingsForAction(defaults, user []ActionBinding, action ActionID, workspace string) []ActionBinding {
	effective := EffectiveBindings(defaults, user)
	out := make([]ActionBinding, 0)
	for _, binding := range effective {
		if binding.Action == action && binding.Workspace == workspace {
			out = append(out, binding)
		}
	}
	return out
}

func BindingConflicts(active []ActionBinding, candidate ActionBinding, labels map[ActionID]string) []BindingConflict {
	if !candidate.Enabled || !HasChord(candidate.Chord) {
		return nil
	}
	out := make([]BindingConflict, 0)
	for _, existing := range active {
		if !existing.Enabled || !HasChord(existing.Chord) {
			continue
		}
		if existing.Action == candidate.Action && existing.Workspace == candidate.Workspace {
			continue
		}
		if !ChordsEqual(existing.Chord, candidate.Chord) {
			continue
		}
		if !bindingScopesConflict(existing.Workspace, candidate.Workspace) {
			continue
		}
		label := string(existing.Action)
		if labels != nil && labels[existing.Action] != "" {
			label = labels[existing.Action]
		}
		out = append(out, BindingConflict{Binding: existing, Label: label})
	}
	return out
}

func RemoveBindingConflicts(defaults, user []ActionBinding, conflicts []BindingConflict) []ActionBinding {
	if len(conflicts) == 0 {
		out := make([]ActionBinding, len(user))
		copy(out, user)
		return out
	}
	conflictActions := map[string]bool{}
	for _, conflict := range conflicts {
		conflictActions[bindingIdentity(conflict.Binding)] = true
	}
	out := make([]ActionBinding, 0, len(user)+len(conflicts))
	for _, binding := range user {
		if conflictActions[bindingIdentity(binding)] {
			continue
		}
		out = append(out, binding)
	}
	for _, binding := range defaults {
		if !conflictActions[bindingIdentity(binding)] {
			continue
		}
		out = append(out, ActionBinding{
			Action:    binding.Action,
			Workspace: binding.Workspace,
			Enabled:   false,
		})
		delete(conflictActions, bindingIdentity(binding))
	}
	return out
}

func ReplaceActionBindings(user []ActionBinding, action ActionID, workspace string, replacements []ActionBinding) []ActionBinding {
	out := make([]ActionBinding, 0, len(user)+len(replacements))
	for _, binding := range user {
		if binding.Action == action && binding.Workspace == workspace {
			continue
		}
		out = append(out, binding)
	}
	out = append(out, replacements...)
	return out
}

func HasChord(chord KeyChord) bool {
	return len(chord.Keys) > 0
}

func ValidChord(chord KeyChord) bool {
	if len(chord.Keys) == 0 {
		return false
	}
	for _, key := range chord.Keys {
		if key <= int(hid.KeyBoardKeyInvalid) || key >= int(hid.KeyboardKeyMaximum) {
			return false
		}
	}
	return true
}

func ChordsEqual(a, b KeyChord) bool {
	if a.Ctrl != b.Ctrl || a.Meta != b.Meta || a.CtrlOrMeta != b.CtrlOrMeta ||
		a.Shift != b.Shift || a.Alt != b.Alt {
		return false
	}
	if len(a.Keys) != len(b.Keys) {
		return false
	}
	ak := sortedKeys(a.Keys)
	bk := sortedKeys(b.Keys)
	for i := range ak {
		if ak[i] != bk[i] {
			return false
		}
	}
	return true
}

func FormatKeyChord(chord KeyChord) string {
	if !HasChord(chord) {
		return ""
	}
	parts := make([]string, 0, len(chord.Keys)+4)
	if chord.CtrlOrMeta {
		parts = append(parts, "Ctrl/Cmd")
	} else {
		if chord.Ctrl {
			parts = append(parts, "Ctrl")
		}
		if chord.Meta {
			parts = append(parts, "Cmd")
		}
	}
	if chord.Shift {
		parts = append(parts, "Shift")
	}
	if chord.Alt {
		parts = append(parts, "Alt")
	}
	for _, key := range chord.Keys {
		parts = append(parts, KeyName(key))
	}
	return strings.Join(parts, "+")
}

func KeyName(key int) string {
	switch hid.KeyboardKey(key) {
	case hid.KeyboardKeyLeft:
		return "Left"
	case hid.KeyboardKeyUp:
		return "Up"
	case hid.KeyboardKeyRight:
		return "Right"
	case hid.KeyboardKeyDown:
		return "Down"
	case hid.KeyboardKeyEscape:
		return "Escape"
	case hid.KeyboardKeyTab:
		return "Tab"
	case hid.KeyboardKeySpace:
		return "Space"
	case hid.KeyboardKeyBackspace:
		return "Backspace"
	case hid.KeyboardKeyBackQuote:
		return "`"
	case hid.KeyboardKeyDelete:
		return "Delete"
	case hid.KeyboardKeyReturn:
		return "Return"
	case hid.KeyboardKeyEnter:
		return "Enter"
	case hid.KeyboardKeyComma:
		return ","
	case hid.KeyboardKeyPeriod:
		return "."
	case hid.KeyboardKeyBackSlash:
		return "\\"
	case hid.KeyboardKeyForwardSlash:
		return "/"
	case hid.KeyboardKeyOpenBracket:
		return "["
	case hid.KeyboardKeyCloseBracket:
		return "]"
	case hid.KeyboardKeySemicolon:
		return ";"
	case hid.KeyboardKeyQuote:
		return "'"
	case hid.KeyboardKeyEqual:
		return "="
	case hid.KeyboardKeyMinus:
		return "-"
	case hid.KeyboardNumKeyDivide:
		return "Num /"
	case hid.KeyboardNumKeyMultiply:
		return "Num *"
	case hid.KeyboardNumKeyAdd:
		return "Num +"
	case hid.KeyboardNumKeySubtract:
		return "Num -"
	case hid.KeyboardNumKeyPeriod:
		return "Num ."
	case hid.KeyboardKeyCapsLock:
		return "Caps Lock"
	case hid.KeyboardKeyScrollLock:
		return "Scroll Lock"
	case hid.KeyboardKeyNumLock:
		return "Num Lock"
	case hid.KeyboardKeyPrintScreen:
		return "Print Screen"
	case hid.KeyboardKeyPause:
		return "Pause"
	case hid.KeyboardKeyInsert:
		return "Insert"
	case hid.KeyboardKeyHome:
		return "Home"
	case hid.KeyboardKeyPageUp:
		return "Page Up"
	case hid.KeyboardKeyPageDown:
		return "Page Down"
	case hid.KeyboardKeyEnd:
		return "End"
	}
	if key >= int(hid.KeyboardKeyA) && key <= int(hid.KeyboardKeyZ) {
		return string(rune('A' + key - int(hid.KeyboardKeyA)))
	}
	if key >= int(hid.KeyboardKey0) && key <= int(hid.KeyboardKey9) {
		return strconv.Itoa(key - int(hid.KeyboardKey0))
	}
	if key >= int(hid.KeyboardNumKey0) && key <= int(hid.KeyboardNumKey9) {
		return "Num " + strconv.Itoa(key-int(hid.KeyboardNumKey0))
	}
	if key >= int(hid.KeyboardKeyF1) && key <= int(hid.KeyboardKeyF12) {
		return fmt.Sprintf("F%d", key-int(hid.KeyboardKeyF1)+1)
	}
	return fmt.Sprintf("Key %d", key)
}

func IsModifierKey(key int) bool {
	switch hid.KeyboardKey(key) {
	case hid.KeyboardKeyLeftAlt, hid.KeyboardKeyRightAlt,
		hid.KeyboardKeyLeftCtrl, hid.KeyboardKeyRightCtrl,
		hid.KeyboardKeyLeftShift, hid.KeyboardKeyRightShift,
		hid.KeyboardKeyLeftMeta, hid.KeyboardKeyRightMeta:
		return true
	default:
		return false
	}
}

func enabledBindings(bindings []ActionBinding) []ActionBinding {
	out := make([]ActionBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Enabled && HasChord(binding.Chord) {
			out = append(out, binding)
		}
	}
	return out
}

func bindingIdentity(binding ActionBinding) string {
	return string(binding.Action) + "\x00" + binding.Workspace
}

func bindingScopesConflict(a, b string) bool {
	return a == "" || b == "" || a == b
}

func sortedKeys(keys []int) []int {
	out := make([]int, len(keys))
	copy(out, keys)
	sort.Ints(out)
	return out
}
