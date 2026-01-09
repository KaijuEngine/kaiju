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

package shader_designer

import (
	"encoding/json"
	"errors"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func collectFileOptions(pfs *project_file_system.FileSystem) map[string][]ui.SelectOption {
	vert := []ui.SelectOption{}
	frag := []ui.SelectOption{}
	geom := []ui.SelectOption{}
	tesc := []ui.SelectOption{}
	tese := []ui.SelectOption{}
	comp := []ui.SelectOption{}
	if dir, err := pfs.ReadDir(project_file_system.SrcShaderFolder); err == nil {
		for i := range dir {
			f := dir[i]
			if f.IsDir() {
				continue
			}
			op := ui.SelectOption{
				Name:  f.Name(),
				Value: filepath.ToSlash(filepath.Join(project_file_system.SrcShaderFolder, f.Name())),
			}
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
			case ".comp":
				comp = append(comp, op)
			}
		}
	}
	sort.Slice(vert, func(i, j int) bool { return vert[i].Name < vert[j].Name })
	sort.Slice(frag, func(i, j int) bool { return frag[i].Name < frag[j].Name })
	sort.Slice(geom, func(i, j int) bool { return geom[i].Name < geom[j].Name })
	sort.Slice(tesc, func(i, j int) bool { return tesc[i].Name < tesc[j].Name })
	sort.Slice(tese, func(i, j int) bool { return tese[i].Name < tese[j].Name })
	sort.Slice(comp, func(i, j int) bool { return comp[i].Name < comp[j].Name })
	return map[string][]ui.SelectOption{
		"Vertex":                 vert,
		"Fragment":               frag,
		"Geometry":               geom,
		"TessellationControl":    tesc,
		"TessellationEvaluation": tese,
		"Compute":                comp,
	}
}

func (win *ShaderDesigner) reloadShaderDoc() {
	defer tracing.NewRegion("ShaderDesigner.reloadShaderDoc").End()
	win.liveShader = false
	sy := float32(0)
	if win.shaderDoc != nil {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.shaderDoc.Destroy()
	}
	data := common_workspace.ReflectUIStructure(win.ed.Cache(),
		&win.shader.ShaderData, "", collectFileOptions(win.ed.ProjectFileSystem()))
	data.Name = "Shader Editor"
	win.shaderDoc, _ = markup.DocumentFromHTMLAsset(win.uiMan, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":     showShaderTooltip,
			"valueChanged":    win.shaderValueChanged,
			"saveData":        win.shaderSave,
			"clickLiveShader": win.clickLiveShader,
		})
	if sy != 0 {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		win.uiMan.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func showShaderTooltip(e *document.Element) {
	defer tracing.NewRegion("shader_designer.showShaderTooltip").End()
	id := e.Attribute("data-tooltip")
	tip, ok := shaderTooltips[id]
	if !ok {
		return
	}
	tipElm := e.Root().FindElementById("toolTip")
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
	defer tracing.NewRegion("ShaderDesigner.shaderValueChanged").End()
	common_workspace.SetObjectValueFromUI(&win.shader, e)
}

func compileShaderFile(id string, pfs *project_file_system.FileSystem, cache *content_database.Cache, src, flags string) (string, error) {
	defer tracing.NewRegion("ShaderDesigner.compileShaderFile").End()
	if src == "" {
		return "", errors.New("blank src")
	}
	path := pfs.FullPath(src)
	flags = strings.TrimSpace(flags)
	out := filepath.Join(os.TempDir(), filepath.Base(path)+".spv")
	if id != "" {
		out = pfs.FullPath(string(project_file_system.SpvPath(id)))
	} else {
		defer os.Remove(out)
	}
	args := []string{path, "-o", out}
	if flags != "" {
		args = append(args, flags)
	}
	cmd := exec.Command("glslc", args...)
	if errStr, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", errors.New(string(errStr))
		}
		return "", err
	}
	if id == "" {
		res, err := content_database.Import(out, pfs, cache, "")
		return res[0].Id, err
	}
	return id, nil
}

func (win *ShaderDesigner) shaderSave(e *document.Element) {
	defer tracing.NewRegion("ShaderDesigner.shaderSave").End()
	s := &win.shader
	addFlags := ""
	if s.EnableDebug {
		addFlags = " -g"
	}
	list := []struct {
		name  string
		stage string
		spv   *string
		flags string
	}{
		{"vertex", s.Vertex, &s.VertexSpv, s.VertexFlags},
		{"fragment", s.Fragment, &s.FragmentSpv, s.FragmentFlags},
		{"geometry", s.Geometry, &s.GeometrySpv, s.GeometryFlags},
		{"tessellation control", s.TessellationControl, &s.TessellationControlSpv, s.TessellationControlFlags},
		{"tessellation evaluation", s.TessellationEvaluation, &s.TessellationEvaluationSpv, s.TessellationEvaluationFlags},
		{"compute", s.Compute, &s.ComputeSpv, s.ComputeFlags},
	}
	if win.shader.Compute != "" {
		for i := range list {
			if list[i].name == "compute" {
				continue
			}
			if list[i].stage != "" {
				slog.Error("compute shaders can not have a " + list[i].name + " shader file associated with it")
				return
			}
		}
	}
	var err error
	pfs := win.ed.ProjectFileSystem()
	cache := win.ed.Cache()
	if win.shader.ShaderData, err = importShaderLayout(pfs, win.shader.ShaderData); err != nil {
		slog.Error("failed to read the shader layout", "error", err)
		return
	}
	for i := range list {
		if list[i].stage != "" {
			if id, err := compileShaderFile(*list[i].spv, pfs, cache, list[i].stage, list[i].flags+addFlags); err != nil {
				slog.Error("failed to compile the "+list[i].name+" shader", "error", err)
				return
			} else {
				*list[i].spv = id
			}
		} else if *list[i].spv != "" {
			content_database.Delete(*list[i].spv, pfs, cache)
			*list[i].spv = ""
		}
	}
	win.shader.Vertex = filepath.ToSlash(win.shader.Vertex)
	win.shader.Fragment = filepath.ToSlash(win.shader.Fragment)
	win.shader.Geometry = filepath.ToSlash(win.shader.Geometry)
	win.shader.TessellationControl = filepath.ToSlash(win.shader.TessellationControl)
	win.shader.TessellationEvaluation = filepath.ToSlash(win.shader.TessellationEvaluation)
	win.shader.Compute = filepath.ToSlash(win.shader.Compute)
	res, err := json.Marshal(win.shader)
	if err != nil {
		slog.Error("failed to marshal the shader data", "error", err)
		return
	}
	if win.shader.id != "" {
		err = pfs.WriteFile(string(project_file_system.ShaderPath(win.shader.id)), res, os.ModePerm)
	} else {
		ids := content_database.ImportRaw(win.shader.Name, res, content_database.Shader{}, pfs, cache)
		if len(ids) > 0 {
			win.shader.id = ids[0]
			win.ed.Events().OnContentAdded.Execute(ids)
		} else {
			err = errors.New("failed to import the raw shader file data to the database")
		}
	}
	if err != nil {
		slog.Error("failed to write the shader data to file", "error", err)
		return
	}
	win.host.ShaderCache().ReloadShader(win.shader.Compile())
	slog.Info("shader successfully saved")
	if e != nil && len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}

func (win *ShaderDesigner) clickLiveShader(elm *document.Element) {
	defer tracing.NewRegion("ShaderDesigner.clickLiveShader").End()
	if win.liveShader {
		elm.InnerLabel().SetText("Start live")
		win.liveShader = false
		return
	}
	elm.InnerLabel().SetText("Stop live")
	win.liveShader = true
	// goroutine
	go func() {
		paths := map[string]time.Time{}
		keys := [...]string{
			win.shader.Vertex,
			win.shader.Fragment,
			win.shader.Geometry,
			win.shader.TessellationControl,
			win.shader.TessellationEvaluation,
		}
		for i := range keys {
			if keys[i] != "" {
				s, err := win.ed.ProjectFileSystem().Stat(keys[i])
				if err == nil && !s.IsDir() {
					paths[keys[i]] = s.ModTime()
				}
			}
		}
		for win.liveShader {
			time.Sleep(time.Second * 1)
			for k, v := range paths {
				s, err := win.ed.ProjectFileSystem().Stat(k)
				if err != nil || s.IsDir() {
					delete(paths, k)
					continue
				}
				if s.ModTime().After(v) {
					paths[k] = s.ModTime()
					win.shaderSave(nil)
					break
				}
			}
		}
	}()
}
