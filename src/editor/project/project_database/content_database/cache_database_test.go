/******************************************************************************/
/* cache_database_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"os"
	"path/filepath"
	"testing"

	"kaijuengine.com/editor/project/project_file_system"
)

// writeFile creates a file at <root>/<rel> with the given bytes. Fatals
// the test on any error so the calling test body can be flat and linear.
func writeFile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	full := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(full), os.ModePerm); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

// newConfigFS builds a tmp-rooted FileSystem with the minimum project
// folder layout required by Cache.Build: just the database/config dir.
// Callers populate it with test fixture files via writeFile.
func newConfigFS(t *testing.T) (*project_file_system.FileSystem, string) {
	t.Helper()
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, project_file_system.ContentConfigFolder), os.ModePerm); err != nil {
		t.Fatalf("mkdir database/config: %v", err)
	}
	pfs, err := project_file_system.New(tmpDir)
	if err != nil {
		t.Fatalf("project_file_system.New: %v", err)
	}
	return &pfs, tmpDir
}

// validConfig is the smallest ContentConfig JSON the decoder will accept
// without surfacing a Name/Type mismatch — the cache stores whatever it
// decodes, so any valid object is fine for "is this file indexed?" checks.
const validConfig = `{"Name":"placeholder","Type":"texture"}`

// macOS Finder writes a 6148-byte Bud1 record into .DS_Store files. The
// fixture below is the first 16 bytes of that format — leading NUL is what
// trips json.NewDecoder.Decode in the unguarded path.
var dsStoreMagic = []byte{0, 0, 0, 1, 'B', 'u', 'd', '1', 0, 0, 0x10, 0, 0, 0, 8, 0}

func TestCacheBuild_SkipsHiddenFiles(t *testing.T) {
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	// Real config plus an assortment of hidden / OS-droppings files.
	writeFile(t, cfgDir, "real.json", []byte(validConfig))
	writeFile(t, cfgDir, ".DS_Store", dsStoreMagic)
	writeFile(t, cfgDir, ".DS_Store.json", dsStoreMagic) // the actual lockout file
	writeFile(t, cfgDir, "._scratch", []byte{0, 0, 0})
	writeFile(t, cfgDir, ".gitignore", []byte("!*.gitignore\n"))

	c := New()
	if err := c.Build(pfs); err != nil {
		t.Fatalf("Build should not fail on hidden files: %v", err)
	}
	if got := len(c.List()); got != 1 {
		t.Fatalf("expected only real.json indexed, got %d entries: %#v", got, c.List())
	}
}

func TestCacheBuild_SkipsNonJSON(t *testing.T) {
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	writeFile(t, cfgDir, "real.json", []byte(validConfig))
	writeFile(t, cfgDir, "notes.txt", []byte("not a config"))
	writeFile(t, cfgDir, "scratch.bak", []byte("backup"))

	c := New()
	if err := c.Build(pfs); err != nil {
		t.Fatalf("Build should not fail on non-.json files: %v", err)
	}
	if got := len(c.List()); got != 1 {
		t.Fatalf("expected only real.json indexed, got %d entries: %#v", got, c.List())
	}
}

func TestCacheBuild_TolerantOfBadJSON(t *testing.T) {
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	writeFile(t, cfgDir, "real.json", []byte(validConfig))
	// Truncated JSON: passes the .json extension + non-hidden filter,
	// but the decoder will return an unexpected-EOF error. Cache must
	// warn + continue rather than abort the whole project open.
	writeFile(t, cfgDir, "bad.json", []byte(`{"Name":`))

	c := New()
	if err := c.Build(pfs); err != nil {
		t.Fatalf("Build should tolerate a single corrupt config: %v", err)
	}
	if got := len(c.List()); got != 1 {
		t.Fatalf("expected real.json indexed despite bad.json corruption, got %d", got)
	}
}

func TestCacheBuild_DSStoreReproRecovery(t *testing.T) {
	// Reproduces the user's 2026-05-26 lockout state verbatim: a fresh
	// project with only a misnamed .DS_Store.json + a real config.
	// Without the fix the original Build would fail with the
	// "invalid character '\x00' looking for beginning of value" error
	// and the user would be locked out. With the fix Build completes
	// and indexes only the real config.
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	writeFile(t, cfgDir, ".DS_Store.json", dsStoreMagic)
	writeFile(t, cfgDir, "ok.json", []byte(validConfig))

	c := New()
	if err := c.Build(pfs); err != nil {
		t.Fatalf("regression: .DS_Store.json must not abort cache build: %v", err)
	}
	if got := len(c.List()); got != 1 {
		t.Fatalf("expected only ok.json indexed, got %d", got)
	}
}

func TestReadConfig_RefusesUpgradeOfHiddenFile(t *testing.T) {
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	// .DS_Store with no extension — the original upgrade path would
	// rename it to .DS_Store.json. Verify both that ReadConfig returns
	// an error AND that the file is not renamed on disk.
	writeFile(t, cfgDir, ".DS_Store", dsStoreMagic)
	relPath := filepath.Join(project_file_system.ContentConfigFolder, ".DS_Store")

	if _, err := ReadConfig(relPath, pfs); err == nil {
		t.Fatal("ReadConfig should refuse hidden non-.json files")
	}
	if _, err := os.Stat(filepath.Join(cfgDir, ".DS_Store")); err != nil {
		t.Fatalf(".DS_Store should still exist (was wrongly renamed?): %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfgDir, ".DS_Store.json")); err == nil {
		t.Fatal("ReadConfig must NOT rename hidden files to .json")
	}
}

func TestReadConfig_RefusesUpgradeOfBinaryFile(t *testing.T) {
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	// Non-hidden binary file. Mimics e.g. a user dropping an image
	// into database/config by accident. First byte is 0x89 (PNG magic),
	// not '{', so the upgrade-rename must refuse.
	writeFile(t, cfgDir, "stray", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	relPath := filepath.Join(project_file_system.ContentConfigFolder, "stray")

	if _, err := ReadConfig(relPath, pfs); err == nil {
		t.Fatal("ReadConfig should refuse non-JSON binary files")
	}
	if _, err := os.Stat(filepath.Join(cfgDir, "stray.json")); err == nil {
		t.Fatal("ReadConfig must NOT rename binary files to .json")
	}
}

func TestReadConfig_UpgradesValidJSONWithoutExtension(t *testing.T) {
	// Sanity check that the original upgrade behaviour still works for
	// real JSON content lacking the .json suffix (the legacy use case
	// the upgrade path was added for in the first place).
	pfs, root := newConfigFS(t)
	cfgDir := filepath.Join(root, project_file_system.ContentConfigFolder)

	writeFile(t, cfgDir, "legacy", []byte(validConfig))
	relPath := filepath.Join(project_file_system.ContentConfigFolder, "legacy")

	cfg, err := ReadConfig(relPath, pfs)
	if err != nil {
		t.Fatalf("ReadConfig of valid legacy config failed: %v", err)
	}
	if cfg.Name != "placeholder" {
		t.Fatalf("expected Name=placeholder, got %q", cfg.Name)
	}
	if _, err := os.Stat(filepath.Join(cfgDir, "legacy.json")); err != nil {
		t.Fatalf("legacy file should have been renamed to legacy.json: %v", err)
	}
}
