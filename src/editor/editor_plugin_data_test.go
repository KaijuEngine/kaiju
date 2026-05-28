/******************************************************************************/
/* editor_plugin_data_test.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kaijuengine.com/editor/editor_plugin"
)

// fakePlugin is a minimal EditorPlugin used to populate
// editorPluginRegistry during tests. Its Launch method is never invoked
// from MissingCompiledPlugins; it exists purely so the registry-key
// lookup succeeds for "registered" plugin entries.
type fakePlugin struct{}

func (fakePlugin) Launch(editor_plugin.EditorInterface) error { return nil }

// writePluginFixture creates a plugin folder under dir containing a
// plugin.json (encoding cfg) and an optional go.mod (when modContents !=
// ""). It returns the absolute folder path so tests can build a
// PluginInfo whose Path matches what AvailablePlugins would produce in
// production.
func writePluginFixture(t *testing.T, dir, folder string, cfg editor_plugin.PluginConfig, modContents string) string {
	t.Helper()
	pluginDir := filepath.Join(dir, folder)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	// plugin.json — not strictly required for MissingCompiledPlugins
	// (the tests construct PluginInfo directly), but written so the
	// fixture is faithful to production layout.
	cfgPath := filepath.Join(pluginDir, "plugin.json")
	f, err := os.Create(cfgPath)
	if err != nil {
		t.Fatalf("create plugin.json: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(`{"PackageName":"` + cfg.PackageName + `","Enabled":true}`); err != nil {
		t.Fatalf("write plugin.json: %v", err)
	}
	if modContents != "" {
		modPath := filepath.Join(pluginDir, "go.mod")
		if err := os.WriteFile(modPath, []byte(modContents), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
	}
	return pluginDir
}

// withRegistryEntry installs key→fakePlugin in editorPluginRegistry for
// the duration of the test, removing it on cleanup. Necessary because
// editorPluginRegistry is a package-level map and tests must not
// permanently mutate it.
func withRegistryEntry(t *testing.T, key string) {
	t.Helper()
	if _, exists := editorPluginRegistry[key]; exists {
		t.Fatalf("registry key %q already populated; test fixture would shadow real plugin", key)
	}
	editorPluginRegistry[key] = fakePlugin{}
	t.Cleanup(func() { delete(editorPluginRegistry, key) })
}

// withAvailablePluginsFn swaps availablePluginsFn for the test duration.
func withAvailablePluginsFn(t *testing.T, fn func() []editor_plugin.PluginInfo) {
	t.Helper()
	original := availablePluginsFn
	availablePluginsFn = fn
	t.Cleanup(func() { availablePluginsFn = original })
}

// captureSlog installs a handler on slog.Default for the test duration
// and returns a buffer the test can inspect for warning/error output.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	original := slog.Default()
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(original) })
	return &buf
}

func TestMissingCompiledPlugins_NoMismatchReturnsEmpty(t *testing.T) {
	const modulePath = "example.com/registered-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "registered_plugin", editor_plugin.PluginConfig{
		PackageName: "registered_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Registered", PackageName: "registered_plugin", Enabled: true},
		}}
	})

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.MissingCompiledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d entries: %+v", len(got), got)
	}
}

func TestMissingCompiledPlugins_RegistryGapDetected(t *testing.T) {
	const modulePath = "example.com/uncompiled-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "uncompiled_plugin", editor_plugin.PluginConfig{
		PackageName: "uncompiled_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	// Note: NO withRegistryEntry call — the registry intentionally
	// lacks this plugin so MissingCompiledPlugins should flag it.
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Uncompiled", PackageName: "uncompiled_plugin", Enabled: true},
		}}
	})

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.MissingCompiledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 missing entry, got %d: %+v", len(got), got)
	}
	if got[0].Config.PackageName != "uncompiled_plugin" {
		t.Fatalf("expected uncompiled_plugin in result, got %q", got[0].Config.PackageName)
	}
}

func TestMissingCompiledPlugins_SessionDisableSkipped(t *testing.T) {
	const modulePath = "example.com/session-disabled-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "session_disabled_plugin", editor_plugin.PluginConfig{
		PackageName: "session_disabled_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	// Registry intentionally empty for this module — without the
	// sessionDisabledPlugins entry the validator would flag it.
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "SessionDisabled", PackageName: "session_disabled_plugin", Enabled: true},
		}}
	})

	ed := &Editor{
		sessionDisabledPlugins: map[string]struct{}{modulePath: {}},
	}
	got, err := ed.MissingCompiledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (session-disabled), got %d entries: %+v", len(got), got)
	}
}

func TestMissingCompiledPlugins_DisabledIgnored(t *testing.T) {
	const modulePath = "example.com/disabled-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "disabled_plugin", editor_plugin.PluginConfig{
		PackageName: "disabled_plugin",
		Enabled:     false,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Disabled", PackageName: "disabled_plugin", Enabled: false},
		}}
	})

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.MissingCompiledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (plugin disabled), got %d entries: %+v", len(got), got)
	}
}

func TestMissingCompiledPlugins_GoModParseErrorTreatsMissing(t *testing.T) {
	dir := t.TempDir()
	// Intentionally malformed go.mod — modfile.Parse must reject this.
	pluginPath := writePluginFixture(t, dir, "broken_plugin", editor_plugin.PluginConfig{
		PackageName: "broken_plugin",
		Enabled:     true,
	}, "this is not a valid go.mod file\n")

	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Broken", PackageName: "broken_plugin", Enabled: true},
		}}
	})

	buf := captureSlog(t)
	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.MissingCompiledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 missing entry (broken go.mod treated as missing), got %d", len(got))
	}
	if got[0].Config.PackageName != "broken_plugin" {
		t.Fatalf("expected broken_plugin in result, got %q", got[0].Config.PackageName)
	}
	logOutput := buf.String()
	if !strings.Contains(logOutput, "failed to determine plugin module path") {
		t.Fatalf("expected warning about module path; slog output was:\n%s", logOutput)
	}
}

func TestRegisterPlugin_KeyHeuristicWarn(t *testing.T) {
	// First: a non-module-path-looking key should fire the warning.
	buf := captureSlog(t)

	const badKey = "badkey-no-slash"
	if _, already := editorPluginRegistry[badKey]; already {
		t.Fatalf("registry already contains key %q; test fixture would shadow", badKey)
	}
	RegisterPlugin(badKey, fakePlugin{})
	t.Cleanup(func() { delete(editorPluginRegistry, badKey) })

	out := buf.String()
	if !strings.Contains(out, "does not look like a module path") {
		t.Fatalf("expected slog warning about module path heuristic; got:\n%s", out)
	}
	if !strings.Contains(out, badKey) {
		t.Fatalf("expected slog warning to include the offending key %q; got:\n%s", badKey, out)
	}

	// Second: a real-looking module path must NOT emit the warning.
	buf.Reset()
	const goodKey = "github.com/foo/bar"
	if _, already := editorPluginRegistry[goodKey]; already {
		t.Fatalf("registry already contains key %q; test fixture would shadow", goodKey)
	}
	RegisterPlugin(goodKey, fakePlugin{})
	t.Cleanup(func() { delete(editorPluginRegistry, goodKey) })

	out = buf.String()
	if strings.Contains(out, "does not look like a module path") {
		t.Fatalf("expected NO module-path warning for %q; got:\n%s", goodKey, out)
	}
}

// withBinaryMtimeFn swaps binaryMtimeFn for the test duration so
// StalePlugins observes a deterministic "binary mtime" without depending
// on the actual test-binary's filesystem mtime.
func withBinaryMtimeFn(t *testing.T, fn func() (time.Time, error)) {
	t.Helper()
	original := binaryMtimeFn
	binaryMtimeFn = fn
	t.Cleanup(func() { binaryMtimeFn = original })
}

// writeFileWithMtime creates a file at path with the given contents and
// sets its access/modification timestamps to mtime. Used by the stale-
// detection tests so they can put a fake .go file "in the past" or "in
// the future" relative to the fake binary mtime without sleeping.
func writeFileWithMtime(t *testing.T, path string, contents string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes %q: %v", path, err)
	}
}

func TestStalePlugins_NoChangesReturnsEmpty(t *testing.T) {
	const modulePath = "example.com/fresh-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "fresh_plugin", editor_plugin.PluginConfig{
		PackageName: "fresh_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	// Source mtime in the past; binary mtime "now" — nothing stale.
	past := time.Now().Add(-2 * time.Hour)
	now := time.Now()
	// Touch the go.mod and add a .go file with the past mtime.
	if err := os.Chtimes(filepath.Join(pluginPath, "go.mod"), past, past); err != nil {
		t.Fatalf("chtimes go.mod: %v", err)
	}
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package fresh_plugin\n", past)

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Fresh", PackageName: "fresh_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return now, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (no source newer than binary), got %d entries: %+v", len(got), got)
	}
}

func TestStalePlugins_NewerSourceFlagged(t *testing.T) {
	const modulePath = "example.com/edited-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "edited_plugin", editor_plugin.PluginConfig{
		PackageName: "edited_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	// Binary "built two hours ago", source touched "now". Expect stale.
	binaryMtime := time.Now().Add(-2 * time.Hour)
	future := time.Now()
	// Push go.mod back so only the .go file is newer than the binary.
	if err := os.Chtimes(filepath.Join(pluginPath, "go.mod"), binaryMtime.Add(-time.Hour), binaryMtime.Add(-time.Hour)); err != nil {
		t.Fatalf("chtimes go.mod: %v", err)
	}
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package edited_plugin\n", future)

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Edited", PackageName: "edited_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return binaryMtime, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 stale entry, got %d: %+v", len(got), got)
	}
	if got[0].Config.PackageName != "edited_plugin" {
		t.Fatalf("expected edited_plugin in result, got %q", got[0].Config.PackageName)
	}
}

func TestStalePlugins_PluginNotInRegistryIgnored(t *testing.T) {
	const modulePath = "example.com/orphan-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "orphan_plugin", editor_plugin.PluginConfig{
		PackageName: "orphan_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	// Source newer than binary, but the plugin is NOT in the registry —
	// phase 09's MissingCompiledPlugins owns this case; StalePlugins must
	// skip it so the modal does not double-list.
	binaryMtime := time.Now().Add(-2 * time.Hour)
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package orphan_plugin\n", time.Now())

	// NO withRegistryEntry call.
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Orphan", PackageName: "orphan_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return binaryMtime, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (not in registry), got %d entries: %+v", len(got), got)
	}
}

func TestStalePlugins_VendorDirSkipped(t *testing.T) {
	const modulePath = "example.com/vendor-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "vendor_plugin", editor_plugin.PluginConfig{
		PackageName: "vendor_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	binaryMtime := time.Now().Add(-2 * time.Hour)
	// All real source files older than binary.
	past := binaryMtime.Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(pluginPath, "go.mod"), past, past); err != nil {
		t.Fatalf("chtimes go.mod: %v", err)
	}
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package vendor_plugin\n", past)
	// vendor/ contains a newer .go file — it MUST be ignored.
	writeFileWithMtime(t, filepath.Join(pluginPath, "vendor", "dep", "foo.go"), "package dep\n", time.Now())

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "Vendor", PackageName: "vendor_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return binaryMtime, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (vendor skipped), got %d entries: %+v", len(got), got)
	}
}

func TestStalePlugins_DotDirSkipped(t *testing.T) {
	const modulePath = "example.com/dotdir-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "dotdir_plugin", editor_plugin.PluginConfig{
		PackageName: "dotdir_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	binaryMtime := time.Now().Add(-2 * time.Hour)
	past := binaryMtime.Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(pluginPath, "go.mod"), past, past); err != nil {
		t.Fatalf("chtimes go.mod: %v", err)
	}
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package dotdir_plugin\n", past)
	// .cache/ contains a newer .go file — must be ignored (explicit name
	// check) and the dot-prefix rule independently covers other names.
	writeFileWithMtime(t, filepath.Join(pluginPath, ".cache", "foo.go"), "package x\n", time.Now())
	writeFileWithMtime(t, filepath.Join(pluginPath, ".idea", "bar.go"), "package y\n", time.Now())

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "DotDir", PackageName: "dotdir_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return binaryMtime, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (dot-dirs skipped), got %d entries: %+v", len(got), got)
	}
}

func TestStalePlugins_NonGoExtensionsIgnored(t *testing.T) {
	const modulePath = "example.com/html-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "html_plugin", editor_plugin.PluginConfig{
		PackageName: "html_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	binaryMtime := time.Now().Add(-2 * time.Hour)
	past := binaryMtime.Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(pluginPath, "go.mod"), past, past); err != nil {
		t.Fatalf("chtimes go.mod: %v", err)
	}
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package html_plugin\n", past)
	// Newer .html / .css / .png — must be ignored (only .go/go.mod/go.sum
	// count for mtime comparison; embedded assets propagate via their
	// embedding .go file's mtime).
	writeFileWithMtime(t, filepath.Join(pluginPath, "ui", "page.html"), "<html></html>\n", time.Now())
	writeFileWithMtime(t, filepath.Join(pluginPath, "ui", "style.css"), "body {}\n", time.Now())
	writeFileWithMtime(t, filepath.Join(pluginPath, "ui", "icon.png"), "fake-png-bytes\n", time.Now())

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "HTML", PackageName: "html_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return binaryMtime, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result (non-Go extensions ignored), got %d entries: %+v", len(got), got)
	}
}

func TestStalePlugins_GoModFlags(t *testing.T) {
	const modulePath = "example.com/gomod-plugin"
	dir := t.TempDir()
	pluginPath := writePluginFixture(t, dir, "gomod_plugin", editor_plugin.PluginConfig{
		PackageName: "gomod_plugin",
		Enabled:     true,
	}, "module "+modulePath+"\n\ngo 1.25.0\n")

	binaryMtime := time.Now().Add(-2 * time.Hour)
	past := binaryMtime.Add(-time.Hour)
	// .go file in the past, but go.mod touched "now" — go.mod alone is
	// enough to flag the plugin as stale.
	writeFileWithMtime(t, filepath.Join(pluginPath, "plugin.go"), "package gomod_plugin\n", past)
	if err := os.Chtimes(filepath.Join(pluginPath, "go.mod"), time.Now(), time.Now()); err != nil {
		t.Fatalf("chtimes go.mod: %v", err)
	}

	withRegistryEntry(t, modulePath)
	withAvailablePluginsFn(t, func() []editor_plugin.PluginInfo {
		return []editor_plugin.PluginInfo{{
			Path:   pluginPath,
			Config: editor_plugin.PluginConfig{Name: "GoMod", PackageName: "gomod_plugin", Enabled: true},
		}}
	})
	withBinaryMtimeFn(t, func() (time.Time, error) { return binaryMtime, nil })

	ed := &Editor{sessionDisabledPlugins: map[string]struct{}{}}
	got, err := ed.StalePlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 stale entry (go.mod newer), got %d: %+v", len(got), got)
	}
	if got[0].Config.PackageName != "gomod_plugin" {
		t.Fatalf("expected gomod_plugin in result, got %q", got[0].Config.PackageName)
	}
}
