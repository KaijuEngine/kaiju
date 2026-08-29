/******************************************************************************/
/* content_database_vector_test.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"path/filepath"
	"testing"

	"kaijuengine.com/editor/project/project_file_system"
)

func TestVectorCategoryRegistration(t *testing.T) {
	cat, ok := CategoryFromTypeName("Vector")
	if !ok {
		t.Fatal("Vector category was not registered")
	}
	if got := cat.Path(); got != project_file_system.ContentVectorFolder {
		t.Fatalf("Path() = %q, want %q", got, project_file_system.ContentVectorFolder)
	}
	if got := cat.ExtNames(); len(got) != 1 || got[0] != ".svg" {
		t.Fatalf("ExtNames() = %v, want [.svg]", got)
	}
}

func TestVectorCategorySelectedByExtension(t *testing.T) {
	cat, ok := selectCategoryForFile(filepath.FromSlash("graphics/icon.svg"))
	if !ok {
		t.Fatal("selectCategoryForFile() did not find .svg")
	}
	if got := cat.TypeName(); got != "Vector" {
		t.Fatalf("TypeName() = %q, want Vector", got)
	}
}

func TestVectorImportPreservesRawText(t *testing.T) {
	dir := t.TempDir()
	svgSrc := `<svg xmlns="http://www.w3.org/2000/svg" width="8" height="8"><rect width="8" height="8" fill="red"/></svg>`
	writeFile(t, dir, "icon.svg", []byte(svgSrc))
	proc, err := (Vector{}).Import(filepath.Join(dir, "icon.svg"), nil)
	if err != nil {
		t.Fatalf("Import() failed: %v", err)
	}
	if len(proc.Variants) != 1 {
		t.Fatalf("Import() returned %d variants, want 1", len(proc.Variants))
	}
	v := proc.Variants[0]
	if v.Name != "icon" {
		t.Fatalf("variant Name = %q, want icon", v.Name)
	}
	if string(v.Data) != svgSrc {
		t.Fatalf("variant Data = %q, want %q", string(v.Data), svgSrc)
	}
}
