/******************************************************************************/
/* editor_plugin_validation.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"fmt"
	"log/slog"
	"strings"

	"kaijuengine.com/editor/editor_overlay/confirm_prompt"
	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/platform/profiler/tracing"
)

// validateCompiledPlugins is the startup-validation gate sandwiched
// between BlurInterface() and newProjectOverlay() inside the
// RunAfterFrames(2, ...) callback at editor_game_interface.go. It runs
// two checks against the user's plugin.json + on-disk plugin sources:
//
//   - MissingCompiledPlugins: enabled plugins in plugin.json whose Go
//     module path is NOT a key in the compiled-in editorPluginRegistry
//     (i.e. the user enabled a plugin since the last editor build, so
//     this binary cannot load it).
//   - StalePlugins: enabled plugins that ARE compiled in but whose source
//     tree contains a .go/go.mod/go.sum file with an mtime newer than the
//     editor binary's mtime (i.e. the user edited the plugin source since
//     the last build, so this binary is running the previous compile).
//
// The two checks are joined into a single confirm_prompt modal:
//
//   - Empty missing + empty stale or any per-step error: onResolved() is
//     invoked immediately so the editor proceeds to the project picker
//     (fail-open — a broken validation step must NEVER block startup).
//   - Non-empty: a confirm_prompt modal is opened with two choices,
//     "Recompile now" and "Continue". The modal title and description
//     adapt to which list(s) are populated; see CONTEXT decision 5 in
//     phases/10-plugin-stale-detection. The modal owns the resolution
//     flow — validateCompiledPlugins does NOT call onResolved on the
//     modal path; the OnConfirm/OnCancel closures handle it.
//
// OnConfirm calls ed.RecompileWithPlugins(<all currently enabled>, onComplete)
// where onComplete is a slog.Error-logging closure (NOT nil — passing nil
// crashes the async build goroutine; see editor_plugin_data.go's
// onComplete dispatch). The existing recompile flow closes the host and
// re-launches the editor via generators/plugin_installer, so onResolved
// is never reached on the success path. Recompile-start failure logs an
// error and invokes onResolved so the user is not stranded on a frozen
// modal.
//
// OnCancel records each MISSING plugin's module path in
// ed.sessionDisabledPlugins so the validator does not re-fire within the
// same process; STALE plugins are NOT tracked there because they will
// load (with their currently-compiled, stale code) and the natural
// startup flow proceeds. The plugin.json on disk is NOT modified —
// restart brings the modal back unless the user explicitly disables the
// plugin from the settings workspace, or recompiles to absorb the source
// changes.
func (ed *Editor) validateCompiledPlugins(onResolved func()) {
	defer tracing.NewRegion("Editor.validateCompiledPlugins").End()
	missing, mErr := ed.MissingCompiledPlugins()
	if mErr != nil {
		slog.Error("editor: validateCompiledPlugins missing-check failed; proceeding to project picker",
			"error", mErr)
		onResolved()
		return
	}
	stale, sErr := ed.StalePlugins()
	if sErr != nil {
		slog.Error("editor: validateCompiledPlugins stale-check failed; proceeding to project picker",
			"error", sErr)
		onResolved()
		return
	}
	if len(missing) == 0 && len(stale) == 0 {
		onResolved()
		return
	}
	title, desc := buildValidationModalCopy(missing, stale)
	if _, err := confirm_prompt.Show(ed.host, confirm_prompt.Config{
		Title:       title,
		Description: desc,
		ConfirmText: "Recompile now",
		CancelText:  "Continue",
		OnConfirm: func() {
			all := availablePluginsFn()
			enabled := make([]editor_plugin.PluginInfo, 0, len(all))
			for _, p := range all {
				if p.Config.Enabled {
					enabled = append(enabled, p)
				}
			}
			// Pass a logger closure (not nil) so the async build/wait
			// goroutine can surface failures. The restart path closes
			// the host on success; on failure the user is already past
			// the startup window and the project picker is gone, so the
			// log is the only signal they get.
			onComplete := func(buildErr error) {
				if buildErr != nil {
					slog.Error("editor: async build for startup-modal recompile failed",
						"error", buildErr)
				}
			}
			if rerr := ed.RecompileWithPlugins(enabled, onComplete); rerr != nil {
				slog.Error("editor: recompile from startup-modal failed to start; proceeding to project picker",
					"error", rerr)
				onResolved()
				return
			}
			// On success, RecompileWithPlugins schedules an async restart
			// (host.Close + plugin_installer goroutine). Do NOT call
			// onResolved here — the project picker for THIS process must
			// not appear; the new process boots and re-runs validation.
		},
		OnCancel: func() {
			// Session-disable only the MISSING plugins so the modal does
			// not re-fire for them within this process. STALE plugins are
			// intentionally NOT tracked: they will load with their
			// currently-compiled (stale) code and the user has explicitly
			// chosen "Continue" knowing that.
			for _, m := range missing {
				mp, mErr := modulePathFromInfo(m)
				if mErr != nil {
					slog.Warn("editor: cannot session-disable plugin (no module path); modal may re-appear on next launch",
						"path", m.Path, "package", m.Config.PackageName, "error", mErr)
					continue
				}
				ed.sessionDisabledPlugins[mp] = struct{}{}
			}
			onResolved()
		},
	}); err != nil {
		slog.Error("editor: failed to show startup-validation modal; proceeding to project picker",
			"error", err)
		onResolved()
	}
}

// buildValidationModalCopy produces the modal title and description text
// from the two validation lists. Three cases:
//
//   - Both missing AND stale populated: combined-modal copy explains both
//     categories side-by-side.
//   - Only missing (phase 09's existing case): preserves the phase-09
//     wording verbatim, just paired with the new "Continue" cancel
//     button.
//   - Only stale: stale-only wording explaining the source-newer-than-
//     binary state.
//
// The copy uses pluralize() so single-plugin descriptions read naturally
// ("foo is …") rather than awkwardly ("foo are …").
func buildValidationModalCopy(missing, stale []editor_plugin.PluginInfo) (title, description string) {
	missingNames := pluginDisplayNames(missing)
	staleNames := pluginDisplayNames(stale)
	switch {
	case len(missing) > 0 && len(stale) > 0:
		title = "Plugins need attention"
		description = fmt.Sprintf(
			"%s %s enabled but not compiled into this editor binary (will not load this session). %s %s source changes that aren't compiled in (will load with current build). Recompile to pick everything up, or continue with the current binary.",
			strings.Join(missingNames, ", "),
			pluralize(len(missing), "is", "are"),
			strings.Join(staleNames, ", "),
			pluralize(len(stale), "has", "have"),
		)
	case len(missing) > 0:
		title = "Plugins not compiled in"
		description = fmt.Sprintf(
			"%s %s enabled in your settings but not compiled into this editor binary. Recompile to include them, or disable them for this session.",
			strings.Join(missingNames, ", "),
			pluralize(len(missing), "is", "are"),
		)
	case len(stale) > 0:
		title = "Plugins have unbuilt source changes"
		description = fmt.Sprintf(
			"%s %s source changes that aren't compiled into this editor binary. Recompile to pick them up, or continue with the current binary.",
			strings.Join(staleNames, ", "),
			pluralize(len(stale), "has", "have"),
		)
	}
	return title, description
}

// pluginDisplayNames extracts Config.Name from each PluginInfo in order
// so the modal description can list plugins in the same order the
// validation surfaced them.
func pluginDisplayNames(list []editor_plugin.PluginInfo) []string {
	names := make([]string, 0, len(list))
	for _, p := range list {
		names = append(names, p.Config.Name)
	}
	return names
}

// pluralize picks the singular or plural form based on n. Used by the
// validation modal so "foo is enabled" / "foo, bar are enabled" both
// read correctly.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
