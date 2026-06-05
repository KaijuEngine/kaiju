package render_graph_workspace

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
	"kaijuengine.com/rendering"
)

func TestRenderGraphGeneratedContentUpsertPreservesID(t *testing.T) {
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer pfs.Close()
	cache := content_database.New()
	events := &editor_events.EditorEvents{}
	ed := &renderGraphOutputTestEditor{
		pfs:    &pfs,
		cache:  &cache,
		events: events,
	}
	workspace := &RenderGraphWorkspace{ed: ed}

	id, err := workspace.upsertRawContent("", "Generated Shader", []byte(`{"Name":"one"}`), content_database.Shader{})
	if err != nil {
		t.Fatalf("first upsert error = %v", err)
	}
	if id == "" {
		t.Fatal("first upsert returned empty id")
	}
	sameID, err := workspace.upsertRawContent(id, "Generated Shader Renamed", []byte(`{"Name":"two"}`), content_database.Shader{})
	if err != nil {
		t.Fatalf("second upsert error = %v", err)
	}
	if sameID != id {
		t.Fatalf("second upsert id = %q, want %q", sameID, id)
	}
	cc, err := cache.Read(id)
	if err != nil {
		t.Fatalf("cache read error = %v", err)
	}
	if cc.Config.Name != "Generated Shader Renamed" {
		t.Fatalf("config name = %q, want Generated Shader Renamed", cc.Config.Name)
	}
	data, err := pfs.ReadFile(cc.ContentPath())
	if err != nil {
		t.Fatalf("read content error = %v", err)
	}
	if !strings.Contains(string(data), `"two"`) {
		t.Fatalf("content = %q, want updated payload", string(data))
	}
}

func TestRenderGraphGeneratedShaderUsesPBRDrawInstanceData(t *testing.T) {
	pfs, err := project_file_system.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer pfs.Close()
	stockShader, err := json.Marshal(rendering.ShaderData{
		Name:         "pbr",
		LayoutGroups: []rendering.ShaderLayoutGroup{{Type: "Vertex"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = pfs.MkdirAll(project_file_system.StockFolder, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err = pfs.WriteFile(project_file_system.StockFolder+"/pbr.shader", stockShader, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	shader, err := buildRenderGraphShaderData(&pfs, "render_graph_abc", "database/src/render/shader/render_graph_abc.frag",
		"fragment.spv", rendering.ShaderLayoutGroup{Type: "Fragment"})
	if err != nil {
		t.Fatalf("buildRenderGraphShaderData error = %v", err)
	}
	if shader.Name != "render_graph_abc" {
		t.Fatalf("shader name = %q, want render_graph_abc", shader.Name)
	}
	if got := shader.DrawInstanceDataName(); got != "pbr" {
		t.Fatalf("shader draw instance data = %q, want pbr", got)
	}
	compiled := shader.Compile()
	if compiled.Name != shader.Name {
		t.Fatalf("compiled shader name = %q, want %q", compiled.Name, shader.Name)
	}
	if got := compiled.DrawInstanceDataName(); got != "pbr" {
		t.Fatalf("compiled shader draw instance data = %q, want pbr", got)
	}
}

type renderGraphOutputTestEditor struct {
	pfs    *project_file_system.FileSystem
	cache  *content_database.Cache
	events *editor_events.EditorEvents
}

var _ editor_workspace.WorkspaceEditorInterface = (*renderGraphOutputTestEditor)(nil)

func (e *renderGraphOutputTestEditor) Host() *engine.Host { return nil }
func (e *renderGraphOutputTestEditor) Cache() *content_database.Cache {
	return e.cache
}
func (e *renderGraphOutputTestEditor) ContentPreviewer() *content_previews.ContentPreviewer {
	return nil
}
func (e *renderGraphOutputTestEditor) Actions() *editor_action.Service { return nil }
func (e *renderGraphOutputTestEditor) Settings() *editor_settings.Settings {
	return nil
}
func (e *renderGraphOutputTestEditor) Events() *editor_events.EditorEvents {
	return e.events
}
func (e *renderGraphOutputTestEditor) History() *memento.History { return nil }
func (e *renderGraphOutputTestEditor) Project() *project.Project { return nil }
func (e *renderGraphOutputTestEditor) ProjectFileSystem() *project_file_system.FileSystem {
	return e.pfs
}
func (e *renderGraphOutputTestEditor) StageView() *editor_stage_view.StageView { return nil }
func (e *renderGraphOutputTestEditor) BlurInterface()                          {}
func (e *renderGraphOutputTestEditor) FocusInterface()                         {}
func (e *renderGraphOutputTestEditor) IsInputFocused() bool                    { return false }
func (e *renderGraphOutputTestEditor) SelectWorkspace(string) error            { return nil }
func (e *renderGraphOutputTestEditor) Workspace(string) (editor_workspace.Workspace, bool) {
	return nil, false
}
func (e *renderGraphOutputTestEditor) Workspaces() []editor_workspace.Workspace {
	return nil
}
func (e *renderGraphOutputTestEditor) UpdateSettings()       {}
func (e *renderGraphOutputTestEditor) ShowReferences(string) {}
