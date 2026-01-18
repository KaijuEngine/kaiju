/******************************************************************************/
/* vfx_workspace.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package vfx_workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_overlay/color_picker"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_overlay/content_selector"
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering/vfx"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
)

type VfxWorkspace struct {
	common_workspace.CommonWorkspace
	ed                       VfxWorkspaceEditorInterface
	stageView                *editor_stage_view.StageView
	systemName               *document.Element
	emitterData              *document.Element
	emitterDataList          *document.Element
	emitterDataTemplate      *document.Element
	emitterList              *document.Element
	emitterListEntryTemplate *document.Element
	emitter                  *vfx.Emitter
	particleSystem           *vfx.ParticleSystem
	entity                   engine.Entity
	systemId                 string
}

func (w *VfxWorkspace) Initialize(host *engine.Host, ed VfxWorkspaceEditorInterface) {
	defer tracing.NewRegion("VfxWorkspace.Initialize").End()
	w.ed = ed
	w.stageView = ed.StageView()
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/vfx_workspace.go.html", nil, map[string]func(*document.Element){
			"clickAddEmitter":      w.clickAddEmitter,
			"clickSaveEmitter":     w.clickSaveEmitter,
			"clickSelectEmitter":   w.clickSelectEmitter,
			"clickDeleteEmitter":   w.clickDeleteEmitter,
			"showColorPicker":      w.showColorPicker,
			"changeEmitterData":    w.changeEmitterData,
			"clickSelectContentId": w.clickSelectContentId,
		})
	w.systemName, _ = w.Doc.GetElementById("systemName")
	w.emitterData, _ = w.Doc.GetElementById("emitterData")
	w.emitterDataList, _ = w.Doc.GetElementById("emitterDataList")
	w.emitterDataTemplate, _ = w.Doc.GetElementById("emitterDataTemplate")
	w.emitterList, _ = w.Doc.GetElementById("emitterList")
	w.emitterListEntryTemplate, _ = w.Doc.GetElementById("emitterListEntryTemplate")
}

func (w *VfxWorkspace) Open() {
	defer tracing.NewRegion("VfxWorkspace.Open").End()
	w.CommonOpen()
	w.stageView.Open()
	w.emitterData.UI.Hide()
	w.emitterDataTemplate.UI.Hide()
	w.emitterListEntryTemplate.UI.Hide()
	w.entity.Transform.SetPosition(w.Host.PrimaryCamera().LookAt())
	w.entity.Transform.SetRotation(matrix.Vec3Zero())
	w.entity.Transform.SetScale(matrix.Vec3One())
	if w.particleSystem == nil {
		w.particleSystem = &vfx.ParticleSystem{}
		w.particleSystem.Initialize(w.Host, &w.entity, vfx.ParticleSystemSpec{})
	}
	w.particleSystem.Activate()
}

func (w *VfxWorkspace) Close() {
	defer tracing.NewRegion("VfxWorkspace.Close").End()
	w.CommonClose()
	w.stageView.Close()
	w.particleSystem.Deactivate()
}

func (w *VfxWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}

func (w *VfxWorkspace) OpenParticleSystem(id string) {
	defer tracing.NewRegion("VfxWorkspace.OpenParticleSystem").End()
	cc, err := w.ed.Project().CacheDatabase().Read(id)
	if err != nil {
		slog.Error("could not find the particle system in the cache database", "id", id, "error", err)
		return
	}
	w.systemId = id
	spec, err := vfx.LoadSpec(w.Host, id)
	if err != nil {
		slog.Error("failed to locate/decode the particle system", "id", id, "error", err)
		return
	}
	w.systemName.UI.ToInput().SetText(cc.Config.Name)
	w.clear()
	for i := range spec {
		w.addEmitter(spec[i])
	}
}

func (w *VfxWorkspace) clear() {
	// > 0 here because we don't want to remove the template
	for i := len(w.emitterList.Children) - 1; i > 0; i-- {
		w.Doc.RemoveElementWithoutApplyStyles(w.emitterList.Children[i])
	}
	w.emitter = nil
	w.particleSystem.Clear()
}

func (w *VfxWorkspace) Update(deltaTime float64) {
	defer tracing.NewRegion("VfxWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.stageView.Update(deltaTime, w.ed.Project())
}

func (w *VfxWorkspace) clickAddEmitter(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.clickTest").End()
	w.entity.Transform.SetPosition(w.Host.PrimaryCamera().LookAt())
	w.addEmitter(vfx.EmitterConfig{
		Texture:          "smoke.png",
		SpawnRate:        0.05,
		ParticleLifeSpan: 2,
		Color:            matrix.ColorWhite(),
		DirectionMin:     matrix.NewVec3(-0.3, 1, -0.3),
		DirectionMax:     matrix.NewVec3(0.3, 1, 0.3),
		VelocityMinMax:   matrix.Vec2One().Scale(1),
		OpacityMinMax:    matrix.NewVec2(0.3, 1.0),
		FadeOutOverLife:  true,
		PathFuncScale:    1,
		PathFuncSpeed:    1,
	})
}

func (w *VfxWorkspace) addEmitter(cfg vfx.EmitterConfig) {
	emit := w.particleSystem.AddEmitter(cfg)
	cpy := w.Doc.DuplicateElementWithoutApplyStyles(w.emitterListEntryTemplate)
	w.Doc.SetElementId(cpy, "")
	cpy.Children[0].InnerLabel().SetText(fmt.Sprintf("Emitter %d", len(w.particleSystem.Emitters)))
	w.selectEmitter(emit)
}

func (w *VfxWorkspace) clickSaveEmitter(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.clickSaveEmitter").End()
	name := w.systemName.UI.ToInput().Text()
	spec := vfx.ParticleSystemSpec{}
	for i := range w.particleSystem.Emitters {
		spec = append(spec, w.particleSystem.Emitters[i].Config)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		slog.Error("failed to serialize the particle system", "error", err)
		return
	}
	pfs := w.ed.Project().FileSystem()
	cache := w.ed.Project().CacheDatabase()
	if w.systemId != "" {
		cc, err := cache.Read(w.systemId)
		if err != nil {
			slog.Error("failed to find the config cache for particle system", "id", w.systemId, "error", err)
			return
		}
		if cc, err = cache.Rename(w.systemId, name, pfs); err != nil && !errors.Is(err, content_database.CacheContentNameEqual) {
			slog.Error("failed to rename the particle system", "id", w.systemId, "error", err)
			return
		}
		path := cc.ContentPath()
		s, err := pfs.Stat(path)
		if err != nil {
			slog.Error("failed to write the particle system", "id", w.systemId, "error", err)
			return
		}
		if err = pfs.WriteFile(path, data, s.Mode()); err != nil {
			slog.Error("failed to write the particle system", "id", w.systemId, "error", err)
			return
		}
		w.ed.Events().OnContentChangesSaved.Execute(w.systemId)
	} else {
		ids := content_database.ImportRaw(name, data, content_database.ParticleSystem{}, pfs, cache)
		w.systemId = ids[0]
		w.ed.Events().OnContentAdded.Execute(ids)
	}
	slog.Info("particle system successfully saved")
}

func (w *VfxWorkspace) clickSelectEmitter(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.clickSelectEmitter").End()
	idx := w.emitterList.IndexOfChild(e) - 1
	if idx >= 0 && idx < len(w.particleSystem.Emitters) {
		w.selectEmitter(&w.particleSystem.Emitters[idx])
	}
}

func (w *VfxWorkspace) clickDeleteEmitter(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.clickDeleteEmitter").End()
	w.ed.BlurInterface()
	confirm_prompt.Show(w.Host, confirm_prompt.Config{
		Title:       "Delete emitter?",
		Description: "Are you sure you'd like to delete this emitter?",
		ConfirmText: "Delete",
		CancelText:  "Cancel",
		OnConfirm: func() {
			w.ed.FocusInterface()
			entry := e.Parent.Value()
			idx := w.emitterList.IndexOfChild(entry) - 1
			w.Doc.RemoveElement(entry)
			w.deleteEmitter(idx)
		},
		OnCancel: w.ed.FocusInterface,
	})
}

func (w *VfxWorkspace) deleteEmitter(idx int) {
	defer tracing.NewRegion("VfxWorkspace.deleteEmitter").End()
	if idx < 0 || idx >= len(w.particleSystem.Emitters) {
		return
	}
	curr := -1
	for i := range w.particleSystem.Emitters {
		if w.emitter == &w.particleSystem.Emitters[i] {
			curr = i
			break
		}
	}
	if curr == idx {
		w.clearSelected()
	}
	w.particleSystem.RemoveEmitter(idx)
}

func (w *VfxWorkspace) clearSelected() {
	defer tracing.NewRegion("VfxWorkspace.clearSelected").End()
	for _, e := range w.emitterList.Children {
		w.Doc.SetElementClassesWithoutApply(e, "edPanelBgHoverable")
	}
	w.emitter = nil
	w.emitterData.UI.Hide()
}

func (w *VfxWorkspace) selectEmitter(emit *vfx.Emitter) {
	defer tracing.NewRegion("VfxWorkspace.selectEmitter").End()
	w.clearSelected()
	w.emitter = emit
	name := "Particle Emitter Data"
	for i := range w.particleSystem.Emitters {
		if emit == &w.particleSystem.Emitters[i] {
			idx := i + 1 // +1 because of the template
			w.Doc.SetElementClassesWithoutApply(w.emitterList.Children[idx],
				"edPanelBgHoverable", "selected")
			name = fmt.Sprintf("Particle Emitter %d", idx)
			break
		}
	}
	w.loadEmitterConfig(name)
}

func (w *VfxWorkspace) loadEmitterConfig(name string) {
	defer tracing.NewRegion("VfxWorkspace.loadEmitterConfig").End()
	w.emitterData.UI.Show()
	w.emitterDataTemplate.UI.Hide()
	g := entity_data_binding.ToDataBinding(name, &w.emitter.Config)
	w.createDataBindingEntry(&g)
}

func (w *VfxWorkspace) createDataBindingEntry(g *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("VfxWorkspace.createDataBindingEntry").End()
	for _, e := range w.emitterDataList.Children[1:] {
		w.Doc.RemoveElement(e)
	}
	bindIdx := len(w.emitterDataTemplate.Parent.Value().Children) - 1
	cpy := w.Doc.DuplicateElement(w.emitterDataTemplate)
	nameSpan := cpy.Children[0]
	fieldDiv := cpy.Children[1]
	nameSpan.InnerLabel().SetText(g.Name)
	fields := []*document.Element{fieldDiv}
	if len(g.Fields) == 0 {
		fieldDiv.UI.Hide()
	} else if len(g.Fields) > 1 {
		fields = append(fields, w.Doc.DuplicateElementRepeat(fieldDiv, len(g.Fields)-1)...)
	}
	t := reflect.ValueOf(g.BoundData).Elem().Type()
	for i := range g.Fields {
		for _, c := range fields[i].Children {
			c.UI.Hide()
		}
		options := []string{}
		if f, ok := t.FieldByName(g.Fields[i].Name); ok {
			if f.Tag.Get("visible") == "false" {
				continue
			} else if op := f.Tag.Get("options"); op != "" {
				options = vfx.EditorReflectionOptions(op)
			}
		}
		fields[i].SetAttribute("data-fieldidx", strconv.Itoa(i))
		fields[i].SetAttribute("data-bindidx", strconv.Itoa(bindIdx))
		nameSpan := fields[i].Children[0]
		nameSpan.UI.Show()
		textInput := fields[i].Children[1]
		checkInput := fields[i].Children[2]
		vec2Input := fields[i].Children[3]
		vec3Input := fields[i].Children[4]
		vec4Input := fields[i].Children[5]
		colorInput := fields[i].Children[6]
		selectInput := fields[i].Children[7]
		var contentIdInput *document.Element
		if len(fields[i].Children) > 8 {
			contentIdInput = fields[i].Children[8]
		}
		nameSpan.InnerLabel().SetText(g.Fields[i].Name)
		if len(options) > 0 {
			selectInput.UI.Show()
			sel := selectInput.Children[0].UI.ToSelect()
			sel.ClearOptions()
			opts := []ui.SelectOption{}
			for _, s := range options {
				opts = append(opts, ui.SelectOption{Name: s, Value: s})
			}
			slices.SortStableFunc(opts, func(a, b ui.SelectOption) int {
				return klib.StringValueCompare(a.Value, b.Value)
			})
			for _, opt := range opts {
				sel.AddOption(opt.Name, opt.Value)
			}
			sel.PickOptionByLabelWithoutEvent(g.FieldNumberAsString(i))
		} else if g.Fields[i].IsContentId() {
			contentIdInput.UI.Show()
			child := contentIdInput.Children[0]
			child.SetAttribute("data-type", g.Fields[i].Type)
			str := g.FieldString(i)
			if str == "" {
				str = fmt.Sprintf("empty (%s)", g.Fields[i].Type)
			}
			child.InnerLabel().SetText(str)
		} else if g.Fields[i].IsInput() {
			textInput.UI.Show()
			u := textInput.Children[0].UI.ToInput()
			u.SetPlaceholder(g.Fields[i].Name + "...")
			if g.Fields[i].IsNumber() {
				u.SetTextWithoutEvent(g.FieldNumberAsString(i))
			} else {
				u.SetTextWithoutEvent(g.FieldString(i))
			}
			w.Doc.RemoveElement(checkInput)
		} else if g.Fields[i].IsCheckbox() {
			checkInput.UI.Show()
			checkInput.Children[0].UI.ToCheckbox().SetCheckedWithoutEvent(g.FieldBool(i))
			w.Doc.RemoveElement(textInput)
		} else if g.Fields[i].IsVec2() {
			vec2Input.UI.Show()
			for j := range 2 {
				c := vec2Input.Children[j].UI.ToInput()
				c.SetTextWithoutEvent(g.FieldVectorComponentAsString(i, j))
				vec2Input.Children[j].SetAttribute("data-inneridx", strconv.Itoa(j))
			}
		} else if g.Fields[i].IsVec3() {
			vec3Input.UI.Show()
			for j := range 3 {
				c := vec3Input.Children[j].UI.ToInput()
				c.SetTextWithoutEvent(g.FieldVectorComponentAsString(i, j))
				vec3Input.Children[j].SetAttribute("data-inneridx", strconv.Itoa(j))
			}
		} else if g.Fields[i].IsVec4() {
			vec4Input.UI.Show()
			for j := range 4 {
				c := vec4Input.Children[j].UI.ToInput()
				c.SetTextWithoutEvent(g.FieldVectorComponentAsString(i, j))
				vec4Input.Children[j].SetAttribute("data-inneridx", strconv.Itoa(j))
			}
		} else if g.Fields[i].IsColor() {
			colorInput.UI.Show()
			color := matrix.Color{}
			for j := range 4 {
				color[j] = g.FieldVectorComponent(i, j)
			}
			colorInput.Children[0].UI.ToPanel().SetColor(color)
		}
	}
	w.Doc.SetupInputTabIndexs()
}

func (w *VfxWorkspace) showColorPicker(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.showColorPicker").End()
	w.ed.BlurInterface()
	color_picker.Show(w.Host, color_picker.Config{
		Color: e.UI.ToPanel().Color(),
		OnAccept: func(color matrix.Color) {
			w.ed.FocusInterface()
			e.UI.ToPanel().SetColor(color)
			w.changeEmitterData(e)
		},
		OnCancel: w.ed.FocusInterface,
	})
}

func (w *VfxWorkspace) changeEmitterData(e *document.Element) {
	defer tracing.NewRegion("VfxWorkspace.changeEmitterData").End()
	root := e.Parent.Value().Parent.Value()
	idx, err := strconv.Atoi(root.Attribute("data-fieldidx"))
	if err != nil {
		return
	}
	v := reflect.ValueOf(&w.emitter.Config).Elem().Field(idx)
	ii := e.Attribute("data-inneridx")
	if ii != "" {
		if iidx, err := strconv.Atoi(ii); err != nil {
			return
		} else {
			v = v.Index(iidx)
		}
	}
	inputText := ""
	switch e.UI.Type() {
	case ui.ElementTypeInput:
		inputText = e.UI.ToInput().Text()
	case ui.ElementTypeSelect:
		inputText = e.UI.ToSelect().Value()
	case ui.ElementTypePanel:
		if e.HasClass("edContentPickInput") {
			inputText = e.InnerLabel().Text()
		}
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(toInt(inputText))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(toUint(inputText))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(toFloat(inputText))
	case reflect.String:
		v.SetString(inputText)
	case reflect.Bool:
		v.SetBool(e.UI.ToCheckbox().IsChecked())
	case reflect.Array, reflect.Slice:
		if e.HasClass("edColorPickInput") {
			color := e.UI.ToPanel().Color()
			for j := 0; j < len(color); j++ {
				v.Index(j).SetFloat(float64(color[j]))
			}
		}
	}
	w.Host.RunOnMainThread(func() {
		w.emitter.ReloadConfig(w.Host)
	})
}

func (w *VfxWorkspace) clickSelectContentId(e *document.Element) {
	defer tracing.NewRegion("WorkspaceDetailsUI.selectContentId").End()
	w.ed.BlurInterface()
	content_selector.Show(w.Host, e.Attribute("data-type"), w.ed.Project().CacheDatabase(),
		func(id string) {
			w.ed.FocusInterface()
			e.InnerLabel().SetText(id)
			w.changeEmitterData(e)
			w.Host.RunOnMainThread(func() {
				w.emitter.ForceReloadConfig(w.Host)
			})
		}, w.ed.FocusInterface)
}

func toInt(str string) int64 {
	defer tracing.NewRegion("toInt").End()
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toUint(str string) uint64 {
	defer tracing.NewRegion("toUint").End()
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseUint(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toFloat(str string) float64 {
	defer tracing.NewRegion("toFloat").End()
	if str == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}
	return 0
}
