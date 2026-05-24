/******************************************************************************/
/* automation.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_scripting

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/matrix"
	"kaijuengine.com/plugins"
)

type logSink interface {
	WriteScriptLog(message string)
}

type EditorAutomation struct {
	ed      editor_workspace.WorkspaceEditorInterface
	project *ProjectAutomation
	stage   *StageAutomation
	sink    logSink
}

type ProjectAutomation struct {
	ed   editor_workspace.WorkspaceEditorInterface
	sink logSink
}

type StageAutomation struct {
	ed   editor_workspace.WorkspaceEditorInterface
	sink logSink
}

type EntityAutomation struct {
	stage  *StageAutomation
	entity *editor_stage_manager.StageEntity
}

func AutomationTypes() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[EditorAutomation](),
		reflect.TypeFor[ProjectAutomation](),
		reflect.TypeFor[StageAutomation](),
		reflect.TypeFor[EntityAutomation](),
	}
}

func NewEditorAutomation(ed editor_workspace.WorkspaceEditorInterface, sink logSink) *EditorAutomation {
	a := &EditorAutomation{ed: ed, sink: sink}
	a.project = &ProjectAutomation{ed: ed, sink: sink}
	a.stage = &StageAutomation{ed: ed, sink: sink}
	return a
}

func (a *EditorAutomation) Log(message string) {
	if a.sink != nil {
		a.sink.WriteScriptLog(message)
	}
}

func (a *EditorAutomation) SelectWorkspace(id string) string {
	if err := a.ed.SelectWorkspace(id); err != nil {
		return err.Error()
	}
	return ""
}

func (a *EditorAutomation) Project() *ProjectAutomation { return a.project }
func (a *EditorAutomation) Stage() *StageAutomation     { return a.stage }

func (p *ProjectAutomation) ReadText(path string) string {
	data, err := p.ed.ProjectFileSystem().ReadFile(cleanProjectPath(path))
	if err != nil {
		p.writeLog(fmt.Sprintf("read failed: %s", err.Error()))
		return ""
	}
	return string(data)
}

func (p *ProjectAutomation) WriteText(path, text string) string {
	path = cleanProjectPath(path)
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := p.ed.ProjectFileSystem().MkdirAll(dir, 0755); err != nil {
			return err.Error()
		}
	}
	if err := p.ed.ProjectFileSystem().WriteFile(path, []byte(text), 0644); err != nil {
		return err.Error()
	}
	return ""
}

func (p *ProjectAutomation) ListContentIds(typeName string) []string {
	var content []content_database.CachedContent
	if strings.TrimSpace(typeName) == "" {
		content = p.ed.Cache().List()
	} else {
		content = p.ed.Cache().ListByType(typeName)
	}
	ids := make([]string, 0, len(content))
	for i := range content {
		ids = append(ids, content[i].Id())
	}
	return ids
}

func (p *ProjectAutomation) ScriptsFolder() string { return project_file_system.EditorScriptsFolder }

func (p *ProjectAutomation) writeLog(message string) {
	if p.sink != nil {
		p.sink.WriteScriptLog(message)
	}
}

func (s *StageAutomation) Open(id string) string {
	s.ed.Events().OnRequestOpenStage.Execute(id)
	if err := s.ed.SelectWorkspace("stage"); err != nil {
		return err.Error()
	}
	return ""
}

func (s *StageAutomation) Save() string {
	man := s.manager()
	if man.IsNew() {
		return "current stage has no content id; create/save it in the editor before script saving"
	}
	if err := man.SaveStage(s.ed.Cache(), s.ed.ProjectFileSystem()); err != nil {
		return err.Error()
	}
	return ""
}

func (s *StageAutomation) CurrentStageId() string { return s.manager().StageId() }

func (s *StageAutomation) Selection() []*EntityAutomation {
	selected := s.manager().Selection()
	out := make([]*EntityAutomation, 0, len(selected))
	for i := range selected {
		out = append(out, &EntityAutomation{stage: s, entity: selected[i]})
	}
	return out
}

func (s *StageAutomation) Entities() []*EntityAutomation {
	entities := s.manager().List()
	out := make([]*EntityAutomation, 0, len(entities))
	for i := range entities {
		out = append(out, &EntityAutomation{stage: s, entity: entities[i]})
	}
	return out
}

func (s *StageAutomation) ClearSelection() { s.manager().ClearSelection() }

func (s *StageAutomation) SelectById(id string) {
	s.manager().SelectAppendEntityById(id)
}

func (s *StageAutomation) CreateEntity(name string) *EntityAutomation {
	e := s.manager().AddEntity(name, matrix.Vec3Zero())
	return &EntityAutomation{stage: s, entity: e}
}

func (s *StageAutomation) DuplicateSelected() []*EntityAutomation {
	dups := s.manager().DuplicateSelectionInPlace(s.ed.Project())
	out := make([]*EntityAutomation, 0, len(dups))
	for i := range dups {
		out = append(out, &EntityAutomation{stage: s, entity: dups[i]})
	}
	return out
}

func (s *StageAutomation) DestroySelected() { s.manager().DestroySelected() }

func (s *StageAutomation) manager() *editor_stage_manager.StageManager {
	return s.ed.StageView().Manager()
}

func (e *EntityAutomation) Id() string {
	if e == nil || e.entity == nil {
		return ""
	}
	return e.entity.StageData.Description.Id
}

func (e *EntityAutomation) Name() string {
	if e == nil || e.entity == nil {
		return ""
	}
	return e.entity.Name()
}

func (e *EntityAutomation) SetName(name string) {
	if e != nil && e.entity != nil {
		e.entity.SetName(name)
	}
}

func (e *EntityAutomation) Position() matrix.Vec3 {
	if e == nil || e.entity == nil {
		return matrix.Vec3Zero()
	}
	return e.entity.Transform.Position()
}

func (e *EntityAutomation) SetPosition(pos matrix.Vec3) {
	if e != nil && e.entity != nil {
		e.entity.Transform.SetPosition(pos)
	}
}

func (e *EntityAutomation) Rotation() matrix.Vec3 {
	if e == nil || e.entity == nil {
		return matrix.Vec3Zero()
	}
	return e.entity.Transform.Rotation()
}

func (e *EntityAutomation) SetRotation(rot matrix.Vec3) {
	if e != nil && e.entity != nil {
		e.entity.Transform.SetRotation(rot)
	}
}

func (e *EntityAutomation) Scale() matrix.Vec3 {
	if e == nil || e.entity == nil {
		return matrix.Vec3One()
	}
	return e.entity.Transform.Scale()
}

func (e *EntityAutomation) SetScale(scale matrix.Vec3) {
	if e != nil && e.entity != nil {
		e.entity.Transform.SetScale(scale)
	}
}

func RunEditorScript(ed editor_workspace.WorkspaceEditorInterface, scriptPath string, sink logSink) error {
	automation := NewEditorAutomation(ed, sink)
	vm, err := plugins.LaunchScript(ed.Host().AssetDatabase(), scriptPath,
		AutomationTypes(), map[string]reflect.Value{
			"editor": reflect.ValueOf(automation),
		})
	if vm != nil {
		defer vm.Close()
	}
	if err != nil {
		return err
	}
	return vm.InvokeGlobalFunctionWithArgs("main", reflect.ValueOf(automation))
}

func cleanProjectPath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(path)), "/")
}
