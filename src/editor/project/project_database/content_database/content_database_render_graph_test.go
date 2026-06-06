package content_database

import (
	"path/filepath"
	"testing"

	"kaijuengine.com/editor/project/project_file_system"
)

func TestRenderGraphCategoryRegistration(t *testing.T) {
	cat, ok := CategoryFromTypeName("RenderGraph")
	if !ok {
		t.Fatal("RenderGraph category was not registered")
	}
	if got := cat.Path(); got != project_file_system.ContentRenderGraphFolder {
		t.Fatalf("Path() = %q, want %q", got, project_file_system.ContentRenderGraphFolder)
	}
	if got := cat.ExtNames(); len(got) != 1 || got[0] != ".rendergraph" {
		t.Fatalf("ExtNames() = %v, want [.rendergraph]", got)
	}
}

func TestRenderGraphCategorySelectedByExtension(t *testing.T) {
	cat, ok := selectCategoryForFile(filepath.FromSlash("materials/test.rendergraph"))
	if !ok {
		t.Fatal("selectCategoryForFile() did not find .rendergraph")
	}
	if got := cat.TypeName(); got != "RenderGraph" {
		t.Fatalf("TypeName() = %q, want RenderGraph", got)
	}
}
