/******************************************************************************/
/* settings_workspace_shortcuts_test.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package settings_workspace

import "testing"

func TestShortcutTextMatchesFilter(t *testing.T) {
	haystack := "Global Save Stage editor.saveStage Ctrl/Cmd+S Saves the current stage"
	for _, query := range []string{
		"",
		"save",
		"global save",
		"ctrl/cmd+s",
		"stage current",
	} {
		if !shortcutTextMatchesFilter(haystack, query) {
			t.Fatalf("expected query %q to match", query)
		}
	}
}

func TestShortcutTextMatchesFilterRequiresAllTokens(t *testing.T) {
	haystack := "Global Save Stage editor.saveStage Ctrl/Cmd+S"
	for _, query := range []string{
		"save missing",
		"content ctrl/cmd+s",
	} {
		if shortcutTextMatchesFilter(haystack, query) {
			t.Fatalf("expected query %q to not match", query)
		}
	}
}
