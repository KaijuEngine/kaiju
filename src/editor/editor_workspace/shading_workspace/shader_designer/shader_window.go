/******************************************************************************/
/* shader_window.go                                                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package shader_designer

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KaijuEngine/kaiju/editor/editor_workspace/common_workspace"
	"github.com/KaijuEngine/kaiju/editor/project/project_file_system"
	"github.com/KaijuEngine/kaiju/engine/ui"
	"github.com/KaijuEngine/kaiju/engine/ui/markup"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/document"
	"github.com/KaijuEngine/kaiju/rendering"

	"github.com/KaijuEngine/uuid"
)

func collectFileOptions(pfs *project_file_system.FileSystem) map[string][]ui.SelectOption {
	vert := []ui.SelectOption{}
	frag := []ui.SelectOption{}
	geom := []ui.SelectOption{}
	tesc := []ui.SelectOption{}
	tese := []ui.SelectOption{}
	if dir, err := pfs.ReadDir(project_file_system.SrcShaderFolder); err == nil {
		for i := range dir {
			f := dir[i]
			if f.IsDir() {
				continue
			}
			op := ui.SelectOption{f.Name(), filepath.ToSlash(filepath.Join(project_file_system.SrcShaderFolder, f.Name()))}
			switch filepath.Ext(f.Name()) {
			case ".vert":
				vert = append(vert, op)
			case ".frag":
				frag = append(frag, op)
			case ".geom":
				geom = append(geom, op)
			case ".tesc":
				tesc = append(tesc, op)
			case ".tese":
				tese = append(tese, op)
			}
		}
	}
	sort.Slice(vert, func(i, j int) bool { return vert[i].Name < vert[j].Name })
	sort.Slice(frag, func(i, j int) bool { return frag[i].Name < frag[j].Name })
	sort.Slice(geom, func(i, j int) bool { return geom[i].Name < geom[j].Name })
	sort.Slice(tesc, func(i, j int) bool { return tesc[i].Name < tesc[j].Name })
	sort.Slice(tese, func(i, j int) bool { return tese[i].Name < tese[j].Name })
	return map[string][]ui.SelectOption{
		"Vertex":                 vert,
		"Fragment":               frag,
		"Geometry":               geom,
		"TessellationControl":    tesc,
		"TessellationEvaluation": tese,
	}
}

func (win *ShaderDesigner) reloadShaderDoc() {
	sy := float32(0)
	if win.shaderDoc != nil {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.shaderDoc.Destroy()
	}
	data := common_workspace.ReflectUIStructure(&win.shader.ShaderData, "", collectFileOptions(win.pfs))
	data.Name = "Shader Editor"
	win.shaderDoc, _ = markup.DocumentFromHTMLAsset(win.uiMan, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":  showShaderTooltip,
			"valueChanged": win.shaderValueChanged,
			"saveData":     win.shaderSave,
		})
	if sy != 0 {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		win.uiMan.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func showShaderTooltip(e *document.Element) {
	id := e.Attribute("data-tooltip")
	tip, ok := shaderTooltips[id]
	if !ok {
		return
	}
	tipElm := e.Root().FindElementById("ToolTip")
	if tipElm == nil || len(tipElm.Children) == 0 {
		return
	}
	lbl := tipElm.Children[0].UI
	if !lbl.IsType(ui.ElementTypeLabel) {
		return
	}
	lbl.ToLabel().SetText(tip)
}

func (win *ShaderDesigner) shaderValueChanged(e *document.Element) {
	common_workspace.SetObjectValueFromUI(&win.shader, e)
}

func compileShaderFile(pfs *project_file_system.FileSystem, s *rendering.ShaderData, src, flags string) error {
	if src == "" {
		return nil
	}
	path := pfs.FullPath(src)
	flags = strings.TrimSpace(flags)
	// TODO:  This CompileVariantName is outdated and needs to be updated
	out := s.CompileVariantName(path, flags)
	args := []string{path, "-o", out}
	if flags != "" {
		args = append(args, flags)
	}
	cmd := exec.Command("glslc", args...)
	if errStr, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errors.New(string(errStr))
		}
		return err
	}
	return nil
}

func (win *ShaderDesigner) shaderSave(e *document.Element) {
	if win.shader.id == "" {
		win.shader.id = uuid.NewString()
	}
	var err error
	if win.shader.ShaderData, err = importShaderLayout(win.pfs, win.shader.ShaderData); err != nil {
		slog.Error("failed to read the shader layout", "error", err)
		return
	}
	s := &win.shader
	addFlags := ""
	if s.EnableDebug {
		addFlags = " -g"
	}
	if err := compileShaderFile(win.pfs, &s.ShaderData, s.Vertex, s.VertexFlags+addFlags); err != nil {
		slog.Error("failed to compile the vertex shader", "error", err)
		return
	}
	if err := compileShaderFile(win.pfs, &s.ShaderData, s.Fragment, s.FragmentFlags+addFlags); err != nil {
		slog.Error("failed to compile the fragment shader", "error", err)
		return
	}
	if err := compileShaderFile(win.pfs, &s.ShaderData, s.Geometry, s.GeometryFlags+addFlags); err != nil {
		slog.Error("failed to compile the geometry shader", "error", err)
		return
	}
	if err := compileShaderFile(win.pfs, &s.ShaderData, s.TessellationControl, s.TessellationControlFlags+addFlags); err != nil {
		slog.Error("failed to compile the tessellation control shader", "error", err)
		return
	}
	if err := compileShaderFile(win.pfs, &s.ShaderData, s.TessellationEvaluation, s.TessellationEvaluationFlags+addFlags); err != nil {
		slog.Error("failed to compile the tessellation evaluation shader", "error", err)
		return
	}
	win.shader.Vertex = filepath.ToSlash(win.shader.Vertex)
	win.shader.Fragment = filepath.ToSlash(win.shader.Fragment)
	win.shader.Geometry = filepath.ToSlash(win.shader.Geometry)
	win.shader.TessellationControl = filepath.ToSlash(win.shader.TessellationControl)
	win.shader.TessellationEvaluation = filepath.ToSlash(win.shader.TessellationEvaluation)
	res, err := json.Marshal(win.shader)
	if err != nil {
		slog.Error("failed to marshal the shader data", "error", err)
		return
	}
	err = win.pfs.WriteFile(filepath.Join(project_file_system.ContentFolder,
		project_file_system.ContentShaderFolder, win.pipeline.id), res, os.ModePerm)
	if err != nil {
		slog.Error("failed to write the shader data to file", "error", err)
		return
	}
	slog.Info("shader successfully saved")
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}
