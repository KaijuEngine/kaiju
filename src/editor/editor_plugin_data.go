/******************************************************************************/
/* editor_plugin_data.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

var editorPluginRegistry = map[string]editor_plugin.EditorPlugin{}

// availablePluginsFn is an indirection over editor_plugin.AvailablePlugins
// so tests can inject a fake plugin set without touching the user's real
// plugin folder. Production code never reassigns it.
var availablePluginsFn = editor_plugin.AvailablePlugins

// binaryMtimeFn is an indirection over os.Executable + os.Stat used by
// StalePlugins so tests can supply a deterministic "binary mtime" without
// touching the real test-binary on disk. Production code never reassigns
// it. The default implementation resolves the running executable's path
// (resolving symlinks on darwin/linux so the mtime reflects the actual
// binary rather than a stable launcher symlink) and stats it for ModTime.
var binaryMtimeFn = defaultBinaryMtime

func defaultBinaryMtime() (time.Time, error) {
	exe, err := os.Executable()
	if err != nil {
		return time.Time{}, err
	}
	// EvalSymlinks lets darwin app-bundle setups and Linux distros that
	// ship a stable launcher symlink to a versioned binary observe the
	// real binary's mtime rather than the symlink's (symlink mtimes are
	// commonly the install timestamp, which never changes across rebuilds
	// and would mask every recompile). On platforms where the executable
	// path is already a regular file, EvalSymlinks is a no-op.
	if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil && resolved != "" {
		exe = resolved
	}
	info, err := os.Stat(exe)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// RegisterPlugin records a plugin implementation in the compiled-in plugin
// registry. It is called from each plugin's package init() function.
//
// Convention: pluginKey MUST equal the plugin's Go module path (as declared
// in the plugin's go.mod `module` line). The startup validator
// (validateCompiledPlugins → MissingCompiledPlugins) depends on this
// convention to match enabled entries in plugin.json against the
// compiled-in registry. A best-effort heuristic emits a slog.Warn if the
// supplied key does not contain a slash, since real Go module paths almost
// always do (e.g. "go.digitalxero.dev/kaiju-code", "github.com/foo/bar").
// Registration still proceeds — the warning is purely advisory so authors
// can fix their plugin without bricking the editor.
//
// Example call shape (placed inside the plugin's package init()):
//
//	func init() {
//	    editor.RegisterPlugin("go.digitalxero.dev/kaiju-code", &Plugin{})
//	}
//
// Why the convention matters: when a fresh editor binary is built from
// source, plugins that are Enabled=true in plugin.json but absent from the
// compiled-in registry are surfaced via a startup modal that offers to
// recompile the editor with them included. The validator joins the two
// sets by module path, so a mismatched key would produce a false positive
// on every launch.
func RegisterPlugin(key string, plugin editor_plugin.EditorPlugin) {
	if !strings.Contains(key, "/") {
		slog.Warn("editor_plugin: RegisterPlugin key does not look like a module path", "key", key)
	}
	if _, ok := editorPluginRegistry[key]; ok {
		slog.Error("a plugin with the given key is already registered", "key", key)
		return
	}
	editorPluginRegistry[key] = plugin
}

// MissingCompiledPlugins returns the enabled plugins from
// editor_plugin.AvailablePlugins() whose Go module path is NOT a key in
// the compiled-in editorPluginRegistry — i.e., plugins the user marked
// Enabled=true in plugin.json but that the current editor binary does
// not contain. The set is used by the startup-validation modal
// (validateCompiledPlugins) to offer a Recompile-or-Continue choice.
//
// Behaviour notes:
//   - Disabled plugins (Config.Enabled == false) are ignored entirely.
//   - Plugins whose module path is in ed.sessionDisabledPlugins are
//     skipped so the modal does not repeat within one process after the
//     user chose "Continue" (which session-disables the missing entry).
//   - Git-source plugins (Path starts with "git://") cannot be inspected
//     for a local go.mod; they are reported as missing if not in the
//     registry, with a slog.Warn so users see the cause.
//   - Per-plugin failures during go.mod parsing are slog.Warn-logged and
//     the plugin is treated as missing (better to surface a stray modal
//     than to silently mask a misconfiguration).
//
// The current implementation cannot return a non-nil error — the
// signature reserves room for future failure modes (e.g. registry
// inspection errors) without an API break.
func (ed *Editor) MissingCompiledPlugins() ([]editor_plugin.PluginInfo, error) {
	defer tracing.NewRegion("Editor.MissingCompiledPlugins").End()
	all := availablePluginsFn()
	missing := make([]editor_plugin.PluginInfo, 0, len(all))
	for _, info := range all {
		if !info.Config.Enabled {
			continue
		}
		modulePath, err := modulePathFromInfo(info)
		if err != nil {
			slog.Warn("editor: failed to determine plugin module path; treating as missing",
				"path", info.Path, "package", info.Config.PackageName, "error", err)
			missing = append(missing, info)
			continue
		}
		if ed.sessionDisabledPlugins != nil {
			if _, suppressed := ed.sessionDisabledPlugins[modulePath]; suppressed {
				continue
			}
		}
		if _, registered := editorPluginRegistry[modulePath]; registered {
			continue
		}
		missing = append(missing, info)
	}
	return missing, nil
}

// modulePathFromInfo is the canonical accessor for the Go module path of
// a plugin discovered via editor_plugin.AvailablePlugins(). It centralises
// the special-case handling for git-source plugins (Path prefixed with
// "git://") and delegates to parsePluginModule for source-tree plugins.
// Used by both MissingCompiledPlugins and validateCompiledPlugins (the
// OnCancel path needs the module path to populate sessionDisabledPlugins).
func modulePathFromInfo(info editor_plugin.PluginInfo) (string, error) {
	if strings.HasPrefix(info.Path, "git://") {
		moduleRef := strings.TrimPrefix(info.Path, "git://")
		return strings.Split(moduleRef, "@")[0], nil
	}
	return parsePluginModule(info.Path, info.Config.PackageName)
}

// pluginSourceMaxMtime walks root (a plugin's source-tree folder) and
// returns the latest mtime among files that influence the compiled binary:
// any file with extension ".go", and the bare names "go.mod" / "go.sum".
// Other files (HTML, CSS, images) are ignored — they are pulled into the
// binary via go:embed directives whose timestamps already propagate to
// the embedding .go file's mtime, so a real source change that affects the
// build is always represented by a touched .go/.mod/.sum file.
//
// The walk skips directories named "vendor", "node_modules", ".git",
// ".cache", and any dir whose name starts with "." (broad dot-dir
// exclusion catches editor caches like .vscode, .idea, .DS_Store-flavoured
// helpers, etc.). Skips happen via filepath.SkipDir so descent into those
// trees is avoided entirely — important for performance on large plugins.
//
// Per-file errors (os.Stat racing a transient FS event, a dangling
// symlink, a permission glitch) are slog.Warn-logged and the walk
// continues. Returning a non-nil error from the WalkDirFunc would abort
// the entire scan, which would bias the stale-detection toward false
// negatives in surprising ways; defensive continuation produces the most
// accurate result given best-effort guarantees.
//
// The returned (time.Time, error) is (max-mtime-seen, walk-level-error).
// A walk-level error (e.g. root does not exist) returns the zero time.
// When no qualifying files are found, the zero time is returned with a
// nil error; callers should treat zero as "no source files observed" and
// decide their own semantics (StalePlugins treats it as not-stale).
func pluginSourceMaxMtime(root string) (time.Time, error) {
	defer tracing.NewRegion("editor.pluginSourceMaxMtime").End()
	var maxMtime time.Time
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Per-entry walk error — log + continue. Returning the error
			// here would abort the entire walk; we want best-effort.
			slog.Warn("editor: pluginSourceMaxMtime walk error; continuing",
				"path", path, "root", root, "error", err)
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			// Skip the root itself even if it starts with "." — only
			// descend-prune children. The root check guards against a
			// plugin folder that legitimately starts with "." being
			// silently skipped.
			if path == root {
				return nil
			}
			if name == "vendor" || name == "node_modules" || name == ".git" || name == ".cache" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		ext := filepath.Ext(name)
		if ext != ".go" && name != "go.mod" && name != "go.sum" {
			return nil
		}
		info, statErr := os.Stat(path)
		if statErr != nil {
			slog.Warn("editor: pluginSourceMaxMtime failed to stat candidate; continuing",
				"path", path, "root", root, "error", statErr)
			return nil
		}
		mt := info.ModTime()
		if mt.After(maxMtime) {
			maxMtime = mt
		}
		return nil
	})
	return maxMtime, walkErr
}

// StalePlugins returns the enabled plugins from editor_plugin.AvailablePlugins()
// whose source folder contains a .go / go.mod / go.sum file newer than the
// running editor binary's mtime — i.e., plugins the user has edited since
// the last `go build` of the editor. The set is used by the
// startup-validation modal (validateCompiledPlugins) alongside
// MissingCompiledPlugins to offer a Recompile-or-Continue choice.
//
// Behaviour notes:
//   - Disabled plugins (Config.Enabled == false) are ignored.
//   - Plugins NOT in editorPluginRegistry are skipped — they are the
//     "missing" case owned by MissingCompiledPlugins; including them
//     here would double-list them in the modal.
//   - Git-source plugins (Path prefixed with "git://") are skipped:
//     there is no local source tree to walk, and a git plugin can only
//     become stale via a fresh `go get`, which is its own workflow.
//   - On os.Executable() / stat failure: slog.Warn + return (nil, nil).
//     Fail-quiet — better silence than false-positive modals.
//   - On per-plugin module-path resolution or walk failure: slog.Warn +
//     skip the plugin. Defensive: false negatives are preferable to
//     surfacing a stale flag with no actionable source data.
//
// The current implementation cannot return a non-nil error — the
// signature reserves room for future failure modes without an API break.
func (ed *Editor) StalePlugins() ([]editor_plugin.PluginInfo, error) {
	defer tracing.NewRegion("Editor.StalePlugins").End()
	binaryMtime, err := binaryMtimeFn()
	if err != nil {
		slog.Warn("editor: StalePlugins cannot resolve binary mtime; skipping check",
			"error", err)
		return nil, nil
	}
	all := availablePluginsFn()
	stale := make([]editor_plugin.PluginInfo, 0, len(all))
	for _, info := range all {
		if !info.Config.Enabled {
			continue
		}
		if strings.HasPrefix(info.Path, "git://") {
			// Git-source plugins have no local source tree to walk;
			// staleness for them is governed by the user's go-get
			// workflow, not by mtime comparison.
			continue
		}
		modulePath, mpErr := modulePathFromInfo(info)
		if mpErr != nil {
			slog.Warn("editor: StalePlugins cannot determine module path; skipping plugin",
				"path", info.Path, "package", info.Config.PackageName, "error", mpErr)
			continue
		}
		if _, registered := editorPluginRegistry[modulePath]; !registered {
			// Not compiled in — MissingCompiledPlugins owns this case.
			continue
		}
		srcMtime, walkErr := pluginSourceMaxMtime(info.Path)
		if walkErr != nil {
			slog.Warn("editor: StalePlugins walk failed; skipping plugin",
				"path", info.Path, "package", info.Config.PackageName, "error", walkErr)
			continue
		}
		// time.After uses wall-clock comparison (monotonic clock readings
		// are stripped by os.Stat / os.Executable since both reach the
		// kernel's mtime field). Zero srcMtime means no qualifying source
		// files were observed — treat as not-stale.
		if !srcMtime.IsZero() && srcMtime.After(binaryMtime) {
			stale = append(stale, info)
		}
	}
	return stale, nil
}

func (ed *Editor) RecompileWithPlugins(plugins []editor_plugin.PluginInfo, onComplete func(err error)) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir, err := filesystem.GameDirectory()
	if err != nil {
		return err
	}
	to := filepath.Join(dir, "editor_build")
	// Phase 07: resolve the absolute plugin-storage folder once. The engine
	// `go.mod`'s `replace <plugin-module> => <absolute-path>` rewrite (see
	// appendPluginRequiresToEngineGoMod below) needs this prefix to identify
	// previously-installed plugin-slot replaces during the prune step. A
	// failure here means we cannot reliably distinguish plugin replaces from
	// other replaces, so abort before mutating editor_build/.
	pluginsFolderPath, err := editor_plugin.PluginsFolder()
	if err != nil {
		slog.Error("failed to resolve plugins folder for recompile", "error", err)
		return fmt.Errorf("resolve plugins folder: %w", err)
	}
	os.MkdirAll(to, os.ModePerm)
	if err = copyEditorCodeForRecompile(to, project_file_system.EngineFS.EngineFileSystemInterface); err != nil {
		return err
	}
	// The recompile rewrites editor_plugin_registry.go from scratch on every
	// invocation. copyEditorCodeForRecompile above just pulled a snapshot of
	// editor_plugin_registry.go out of EngineFS — but that snapshot was
	// produced by the previous recompile's `go build` (which captured this
	// very file via the //go:embed * directive in main.ed.go), so it may
	// carry stale `_ "<previously-enabled-plugin>"` lines that the current
	// `plugins` slice no longer wants. Truncate-on-open discards the polluted
	// copy; the header below + the per-plugin loop populate the file purely
	// from the enabled-plugin slice. Without the truncate, disabled plugins
	// would survive across recompile cycles forever via the embed snapshot.
	registry, err := os.OpenFile(filepath.Join(to, "editor_plugin_registry.go"), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer registry.Close()
	if _, err := registry.WriteString("// Code generated by editor_plugin package; DO NOT EDIT.\n\npackage main\n\nimport (\n"); err != nil {
		return err
	}
	// enabledPluginSourcePaths caches the source-tree plugins enabled this
	// recompile, keyed by the plugin's declared module path (parsed from its
	// go.mod) and valued at the plugin's absolute source folder under
	// editor_plugin.PluginsFolder(). The map is consumed by
	// appendPluginRequiresToEngineGoMod after the per-plugin loop to rewrite
	// the engine go.mod's plugin-slot region: each entry becomes one
	// `require <module>` + one `replace <module> => <absolute-source-path>`.
	enabledPluginSourcePaths := map[string]string{}
	for i := range plugins {
		if !plugins[i].Config.Enabled {
			continue
		}
		if strings.HasPrefix(plugins[i].Path, "git://") {
			moduleRef := strings.TrimPrefix(plugins[i].Path, "git://")
			modulePath := strings.Split(moduleRef, "@")[0]

			getCmd := exec.Command("go", "get", moduleRef)
			getCmd.Dir = to
			if output, getErr := getCmd.CombinedOutput(); getErr != nil {
				slog.Error("failed to get git plugin module", "module", moduleRef, "error", getErr, "output", string(output))
				return getErr
			}

			if _, err := registry.WriteString(fmt.Sprintf("\t_ \"%s\"\n", modulePath)); err != nil {
				return err
			}
			continue
		}
		dstName := plugins[i].Config.PackageName
		// Phase 07: validate the plugin's go.mod BEFORE registering it so a
		// plugin with a disallowed `replace` directive hard-fails the install
		// before any editor_build/go.mod mutation runs. No source copy occurs
		// — the engine go.mod's `replace <plugin-module> => <plugins[i].Path>`
		// (written by appendPluginRequiresToEngineGoMod after this loop)
		// resolves the plugin in-place from its own source folder.
		modulePath, err := parsePluginModule(plugins[i].Path, dstName)
		if err != nil {
			return err
		}
		enabledPluginSourcePaths[modulePath] = plugins[i].Path

		if _, err := registry.WriteString(fmt.Sprintf("\t_ \"%s\"\n", modulePath)); err != nil {
			return err
		}
		if err = editor_plugin.UpdatePluginConfigState(plugins[i]); err != nil {
			slog.Warn("failed to update the enabled state of the plugin",
				"name", plugins[i].Config.Name, "package", plugins[i].Config.PackageName, "error", err)
		}
	}
	if _, err := registry.WriteString(")\n"); err != nil {
		return err
	}

	// Phase 07: rewrite the engine `go.mod` plugin-slot region with
	// prune-then-add semantics. The helper first parses
	// editor_build/go.mod, drops every require + replace pair whose replace
	// target is under editor_plugin.PluginsFolder() (cleaning out entries
	// from any prior recompile — including plugins disabled since), then
	// adds one `require <module>` + one `replace <module> => <absolute
	// source path>` per currently-enabled source-tree plugin. The helper
	// runs UNCONDITIONALLY on every recompile so disable-all scenarios
	// still produce a clean engine go.mod (the prune step removes stale
	// entries even when enabledPluginSourcePaths is empty). The zero
	// pseudo-version is the standard Go placeholder for a require whose
	// actual resolution comes from a local replace. This must run BEFORE
	// the `go mod tidy -e` block below so tidy reconciles the rewritten
	// entries against each plugin's transitive deps.
	if err := appendPluginRequiresToEngineGoMod(filepath.Join(to, "go.mod"), pluginsFolderPath, enabledPluginSourcePaths); err != nil {
		return err
	}

	// Plugins may introduce dependencies that are not yet present in the
	// regenerated editor's go.mod (since each plugin's go.mod was stripped
	// during the copy above). Run `go mod tidy` from the editor_build
	// directory so the subsequent `go build` can resolve those modules.
	// `-e` lets tidy proceed past unresolved imports in test files (the
	// embedded engine snapshot can contain test-only imports that resolve
	// to packages outside the module graph — those do not block the actual
	// editor build, which never compiles test files). Any genuine
	// plugin-dep resolution failure still surfaces at the subsequent
	// `go build` step below, and is logged here as a warning.
	tidyCmd := exec.Command("go", "mod", "tidy", "-e")
	tidyCmd.Dir = to
	var tidyOutput strings.Builder
	tidyCmd.Stdout = &tidyOutput
	tidyCmd.Stderr = &tidyOutput
	if err := tidyCmd.Run(); err != nil {
		slog.Error("failed to run go mod tidy on plugin-augmented editor build", "error", err, "output", tidyOutput.String())
		return fmt.Errorf("go mod tidy: %w", err)
	}
	if out := tidyOutput.String(); out != "" {
		slog.Warn("go mod tidy emitted diagnostics on plugin-augmented editor build", "output", out)
	}

	var cmd *exec.Cmd
	if build.Debug {
		cmd = exec.Command("go", "build", "-tags=debug,editor,filedrop", "-o", filepath.Base(exe), ".")
	} else {
		cmd = exec.Command("go", "build", "-tags=editor,filedrop", "-o", filepath.Base(exe), ".")
	}
	cmd.Dir = to

	var buildOutput strings.Builder
	cmd.Stdout = &buildOutput
	cmd.Stderr = &buildOutput

	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		// Nil-safe callback dispatch: callers passing nil (e.g., the
		// startup-validation modal's OnConfirm — see
		// editor_plugin_validation.go) should not crash the restart
		// goroutine. Engine-level fix landed alongside M1 phase 09.
		defer func() {
			if onComplete != nil {
				onComplete(err)
			}
		}()
		err = cmd.Wait()
		if err != nil {
			slog.Error("failed to compile the editor with the plugins", "error", err, "output", buildOutput.String())
			return
		}
		toExe := filepath.Join(to, filepath.Base(exe))
		boot := exec.Command("go", "run", "generators/plugin_installer/main.go", exe, toExe)
		boot.Dir = to
		if err = boot.Start(); err != nil {
			slog.Error("failed to start the restart boot process", "error", err)
			return
		}
		slog.Info("attempting to restart editor with new build")
		ed.host.Close()
	}()
	return nil
}

// Phase 06+07 module-graph rewrite helpers below use the following
// golang.org/x/mod/modfile API surface:
//   - modfile.Parse            — parse plugin and engine go.mod files in-memory
//   - modfile.DropRequire      — remove a `require <module>` directive from the engine go.mod (phase 07 prune step)
//   - modfile.DropReplace      — remove a `replace <module>` directive from the engine go.mod (phase 07 prune step)
//   - modfile.AddRequire       — append/upsert `require <plugin-module> v0.0.0-...` on the engine go.mod
//   - modfile.AddReplace       — upsert `replace <plugin-module> => <absolute plugin source path>` on the engine go.mod
//   - modfile.Format           — serialise the mutated *modfile.File back to []byte before os.WriteFile

// parsePluginModule reads and parses the plugin's source-tree go.mod at
// <pluginDir>/go.mod, validates that every `replace` directive only rewrites
// `kaijuengine.com` (the dev-only engine alias), and returns the plugin's
// declared module path. Any `replace` directive whose Old.Path is not
// "kaijuengine.com" causes a hard-fail: ALL offenders are collected into a
// single error so plugin authors can fix everything in one pass rather than
// chase failures one at a time. The check happens before the plugin's
// source is copied so disallowed-replace plugins never produce partial
// editor_build/ state.
func parsePluginModule(pluginDir, packageName string) (string, error) {
	path := filepath.Join(pluginDir, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("plugin %q: missing go.mod at %s: %w", packageName, path, err)
	}
	file, err := modfile.Parse(path, data, nil)
	if err != nil {
		return "", fmt.Errorf("plugin %q: parse go.mod: %w", packageName, err)
	}
	var offenders []string
	for _, r := range file.Replace {
		if r.Old.Path == "kaijuengine.com" {
			continue
		}
		old := r.Old.Path
		if r.Old.Version != "" {
			old = fmt.Sprintf("%s %s", r.Old.Path, r.Old.Version)
		}
		newRef := r.New.Path
		if r.New.Version != "" {
			newRef = fmt.Sprintf("%s %s", r.New.Path, r.New.Version)
		}
		offenders = append(offenders, fmt.Sprintf("%s => %s", old, newRef))
	}
	if len(offenders) > 0 {
		slog.Error("plugin has disallowed replace directive in go.mod; only 'kaijuengine.com' may be replaced",
			"plugin", packageName, "replaces", offenders)
		return "", fmt.Errorf("plugin %q: disallowed replace directives: %v", packageName, offenders)
	}
	if file.Module == nil || file.Module.Mod.Path == "" {
		return "", fmt.Errorf("plugin %q: go.mod is missing a module declaration at %s", packageName, path)
	}
	return file.Module.Mod.Path, nil
}

// appendPluginRequiresToEngineGoMod rewrites the plugin-slot region of the
// regenerated editor_build/go.mod with prune-then-add semantics. The helper:
//
//  1. Reads + parses engineGoModPath (the absolute path to
//     editor_build/go.mod).
//  2. Walks file.Replace; collects every replace whose New.Path is under
//     pluginsFolderPath (the absolute path returned by
//     editor_plugin.PluginsFolder()) into a drop slice. These are the
//     plugin-slot entries left over from any prior recompile.
//  3. For each module path in the drop slice, calls DropRequire +
//     DropReplace so a plugin disabled since the last recompile is fully
//     removed from the engine go.mod.
//  4. For each (modulePath, absoluteSourcePath) entry in
//     enabledPluginSourcePaths, calls AddRequire with the standard zero
//     pseudo-version and AddReplace pointing at the absolute plugin source
//     folder. AddRequire / AddReplace upsert in place so a plugin pruned in
//     step 3 and re-added here lands cleanly with no duplicates.
//  5. Formats + writes the result back to engineGoModPath.
//
// The helper runs UNCONDITIONALLY on every recompile: empty
// enabledPluginSourcePaths is still useful because step 3 must run to clean
// up any previously-installed plugin entries. The subsequent
// `go mod tidy -e` reconciles the engine's go.sum with each enabled
// plugin's transitive deps.
//
// Plugin-slot identification uses filepath.Rel(pluginsFolderPath,
// replace.New.Path): if Rel returns an error, an empty string, ".", or a
// path starting with "..", the replace is NOT under pluginsFolderPath and
// is left untouched. This is more robust than a HasPrefix check against
// sibling directories that share a leading substring.
func appendPluginRequiresToEngineGoMod(engineGoModPath, pluginsFolderPath string, enabledPluginSourcePaths map[string]string) error {
	data, err := os.ReadFile(engineGoModPath)
	if err != nil {
		slog.Error("failed to read regenerated engine go.mod", "path", engineGoModPath, "error", err)
		return fmt.Errorf("engine go.mod read: %w", err)
	}
	file, err := modfile.Parse(engineGoModPath, data, nil)
	if err != nil {
		slog.Error("failed to parse regenerated engine go.mod", "path", engineGoModPath, "error", err)
		return fmt.Errorf("engine go.mod parse: %w", err)
	}
	// Prune step: collect modules whose replace target lives under the
	// plugins folder. Build the drop list during the scan to avoid mutating
	// file.Replace while iterating it.
	var toDrop []string
	for _, r := range file.Replace {
		if r == nil {
			continue
		}
		if !isPathUnder(r.New.Path, pluginsFolderPath) {
			continue
		}
		toDrop = append(toDrop, r.Old.Path)
	}
	for _, modPath := range toDrop {
		if err := file.DropRequire(modPath); err != nil {
			slog.Error("failed to drop require from engine go.mod",
				"module", modPath, "path", engineGoModPath, "error", err)
			return fmt.Errorf("engine go.mod drop require %q: %w", modPath, err)
		}
		if err := file.DropReplace(modPath, ""); err != nil {
			slog.Error("failed to drop replace from engine go.mod",
				"module", modPath, "path", engineGoModPath, "error", err)
			return fmt.Errorf("engine go.mod drop replace %q: %w", modPath, err)
		}
	}
	// Add step: append (or upsert) one require + replace per currently
	// enabled source-tree plugin, pointing the replace at the plugin's
	// absolute source folder under editor_plugin.PluginsFolder().
	const pluginZeroVersion = "v0.0.0-00010101000000-000000000000"
	for modulePath, sourcePath := range enabledPluginSourcePaths {
		if err := file.AddRequire(modulePath, pluginZeroVersion); err != nil {
			slog.Error("failed to add require to engine go.mod",
				"module", modulePath, "path", engineGoModPath, "error", err)
			return fmt.Errorf("engine go.mod add require %q: %w", modulePath, err)
		}
		if err := file.AddReplace(modulePath, "", sourcePath, ""); err != nil {
			slog.Error("failed to add replace to engine go.mod",
				"module", modulePath, "target", sourcePath, "path", engineGoModPath, "error", err)
			return fmt.Errorf("engine go.mod add replace %q: %w", modulePath, err)
		}
	}
	out, err := file.Format()
	if err != nil {
		slog.Error("failed to format regenerated engine go.mod", "path", engineGoModPath, "error", err)
		return fmt.Errorf("engine go.mod format: %w", err)
	}
	if err := os.WriteFile(engineGoModPath, out, 0o644); err != nil {
		slog.Error("failed to write regenerated engine go.mod", "path", engineGoModPath, "error", err)
		return fmt.Errorf("engine go.mod write: %w", err)
	}
	return nil
}

// isPathUnder reports whether candidate lives at or under base, using
// filepath.Rel to handle trailing separators, case, and sibling directories
// with similar names robustly. A candidate that equals base, or whose
// relative path starts with "..", is NOT considered under base.
func isPathUnder(candidate, base string) bool {
	if candidate == "" || base == "" {
		return false
	}
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return false
	}
	if rel == "" || rel == "." {
		return false
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

func copyEditorCodeForRecompile(to string, efs project_file_system.EngineFileSystemInterface) error {
	const from = "."
	var err error
	var copyFolder func(path string) error
	os.RemoveAll(to)
	os.MkdirAll(to, os.ModePerm)
	copyFolder = func(path string) error {
		relPath, _ := filepath.Rel(from, path)
		folder := filepath.Join(to, relPath)
		if path != "." {
			if err := os.MkdirAll(folder, os.ModePerm); err != nil {
				return err
			}
		}
		var dir []fs.DirEntry
		if dir, err = efs.ReadDir(path); err != nil {
			return err
		}
		for i := range dir {
			name := dir[i].Name()
			entryPath := filepath.ToSlash(filepath.Join(path, name))
			if dir[i].IsDir() {
				if copyFolder(entryPath); err != nil {
					return err
				} else {
					continue
				}
			}
			if strings.HasPrefix(name, "__") {
				continue
			}
			f, err := efs.Open(entryPath)
			if err != nil {
				return err
			}
			defer f.Close()
			t, err := os.Create(filepath.Join(folder, dir[i].Name()))
			if err != nil {
				return err
			}
			defer t.Close()
			if _, err := io.Copy(t, f); err != nil {
				return err
			}
		}
		return nil
	}
	copyFolder(from)
	return err
}
