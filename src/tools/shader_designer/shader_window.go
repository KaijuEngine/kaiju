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
	"kaiju/engine/systems/logging"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/rendering"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func setupShaderDoc(win *ShaderDesigner) {
	win.reloadShaderDoc()
	win.shaderDoc.Deactivate()
}

func collectFileOptions() map[string][]string {
	vert := []string{}
	frag := []string{}
	geom := []string{}
	tesc := []string{}
	tese := []string{}
	if dir, err := os.ReadDir(shaderSrcFolder); err == nil {
		for i := range dir {
			f := dir[i]
			if f.IsDir() {
				continue
			}
			switch filepath.Ext(f.Name()) {
			case ".vert":
				vert = append(vert, filepath.ToSlash(filepath.Join(shaderSrcFolder, f.Name())))
			case ".frag":
				frag = append(frag, filepath.ToSlash(filepath.Join(shaderSrcFolder, f.Name())))
			case ".geom":
				geom = append(geom, filepath.ToSlash(filepath.Join(shaderSrcFolder, f.Name())))
			case ".tesc":
				tesc = append(tesc, filepath.ToSlash(filepath.Join(shaderSrcFolder, f.Name())))
			case ".tese":
				tese = append(tese, filepath.ToSlash(filepath.Join(shaderSrcFolder, f.Name())))
			}
		}
	}
	return map[string][]string{
		"Vertex":                 sort.StringSlice(vert),
		"Fragment":               sort.StringSlice(frag),
		"Geometry":               sort.StringSlice(geom),
		"TessellationControl":    sort.StringSlice(tesc),
		"TessellationEvaluation": sort.StringSlice(tese),
	}
}

func (win *ShaderDesigner) reloadShaderDoc() {
	sy := float32(0)
	if win.shaderDoc != nil {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.shaderDoc.Destroy()
	}
	data := reflectUIStructure(&win.shader, "", collectFileOptions())
	data.Name = "Shader Editor"
	win.shaderDoc, _ = markup.DocumentFromHTMLAsset(&win.man, dataInputHTML,
		data, map[string]func(*document.Element){
			"showTooltip":  showShaderTooltip,
			"valueChanged": win.shaderValueChanged,
			"returnHome":   win.returnHome,
			"saveData":     win.shaderSave,
		})
	if sy != 0 {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		win.man.Host.RunAfterFrames(2, func() {
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
	setObjectValueFromUI(&win.shader, e)
}

func OpenShader(path string, logStream *logging.LogStream) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to load the shader file", "file", path, "error", err)
		return
	}
	var sh rendering.ShaderData
	if err := json.Unmarshal(data, &sh); err != nil {
		slog.Error("failed to unmarshal the shader data", "error", err)
		return
	}
	newInternal(StateRenderPass, logStream, func(sd *ShaderDesigner) {
		sd.shader = sh
		sd.ShowShaderWindow()
	})
}

func compileShaderFile(s *rendering.ShaderData, src, flags string) error {
	if src == "" {
		return nil
	}
	flags = strings.TrimSpace(flags)
	out := s.CompileVariantName(src, flags)
	args := []string{src, "-o", out}
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
	const saveRoot = shaderFolder
	if err := os.MkdirAll(saveRoot, os.ModePerm); err != nil {
		slog.Error("failed to create the shader folder",
			"folder", saveRoot, "error", err)
	}
	var err error
	if win.shader, err = importShaderLayout(win.shader); err != nil {
		slog.Error("failed to read the shader layout", "error", err)
		return
	}
	path := filepath.Join(saveRoot, win.shader.Name+fileExtensionShader)
	s := &win.shader
	addFlags := ""
	if s.EnableDebug {
		addFlags = " -g"
	}
	if err := compileShaderFile(s, s.Vertex, s.VertexFlags+addFlags); err != nil {
		slog.Error("failed to compile the vertex shader", "error", err)
		return
	}
	if err := compileShaderFile(s, s.Fragment, s.FragmentFlags+addFlags); err != nil {
		slog.Error("failed to compile the fragment shader", "error", err)
		return
	}
	if err := compileShaderFile(s, s.Geometry, s.GeometryFlags+addFlags); err != nil {
		slog.Error("failed to compile the geometry shader", "error", err)
		return
	}
	if err := compileShaderFile(s, s.TessellationControl, s.TessellationControlFlags+addFlags); err != nil {
		slog.Error("failed to compile the tessellation control shader", "error", err)
		return
	}
	if err := compileShaderFile(s, s.TessellationEvaluation, s.TessellationEvaluationFlags+addFlags); err != nil {
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
	if err := os.WriteFile(path, res, os.ModePerm); err != nil {
		slog.Error("failed to write the shader data to file", "error", err)
		return
	}
	slog.Info("shader successfully saved", "file", path)
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}
