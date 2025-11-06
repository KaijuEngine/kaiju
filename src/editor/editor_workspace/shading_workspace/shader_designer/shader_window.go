package shader_designer

import (
	"encoding/json"
	"errors"
	"kaiju/editor/project/project_file_system"
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

	"github.com/KaijuEngine/uuid"
)

func collectFileOptions() map[string][]ui.SelectOption {
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
	out := map[string][]ui.SelectOption{
		"Vertex":                 make([]ui.SelectOption, len(vert)),
		"Fragment":               make([]ui.SelectOption, len(frag)),
		"Geometry":               make([]ui.SelectOption, len(geom)),
		"TessellationControl":    make([]ui.SelectOption, len(tesc)),
		"TessellationEvaluation": make([]ui.SelectOption, len(tese)),
	}
	vert = sort.StringSlice(vert)
	frag = sort.StringSlice(frag)
	geom = sort.StringSlice(geom)
	tesc = sort.StringSlice(tesc)
	tese = sort.StringSlice(tese)
	for i := range vert {
		out["Vertex"][i] = ui.SelectOption{vert[i], vert[i]}
	}
	for i := range frag {
		out["Fragment"][i] = ui.SelectOption{frag[i], frag[i]}
	}
	for i := range geom {
		out["Geometry"][i] = ui.SelectOption{geom[i], geom[i]}
	}
	for i := range tesc {
		out["TessellationControl"][i] = ui.SelectOption{tesc[i], tesc[i]}
	}
	for i := range tese {
		out["TessellationEvaluation"][i] = ui.SelectOption{tese[i], tese[i]}
	}
	return out
}

func (win *ShaderDesigner) reloadShaderDoc() {
	sy := float32(0)
	if win.shaderDoc != nil {
		content := win.shaderDoc.GetElementsByClass("topFields")[0]
		sy = content.UIPanel.ScrollY()
		win.shaderDoc.Destroy()
	}
	data := reflectUIStructure(&win.shader.ShaderData, "", collectFileOptions())
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
	setObjectValueFromUI(&win.shader, e)
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
	if win.shader.id == "" {
		win.shader.id = uuid.NewString()
	}
	var err error
	if win.shader.ShaderData, err = importShaderLayout(win.shader.ShaderData); err != nil {
		slog.Error("failed to read the shader layout", "error", err)
		return
	}
	s := &win.shader
	addFlags := ""
	if s.EnableDebug {
		addFlags = " -g"
	}
	if err := compileShaderFile(&s.ShaderData, s.Vertex, s.VertexFlags+addFlags); err != nil {
		slog.Error("failed to compile the vertex shader", "error", err)
		return
	}
	if err := compileShaderFile(&s.ShaderData, s.Fragment, s.FragmentFlags+addFlags); err != nil {
		slog.Error("failed to compile the fragment shader", "error", err)
		return
	}
	if err := compileShaderFile(&s.ShaderData, s.Geometry, s.GeometryFlags+addFlags); err != nil {
		slog.Error("failed to compile the geometry shader", "error", err)
		return
	}
	if err := compileShaderFile(&s.ShaderData, s.TessellationControl, s.TessellationControlFlags+addFlags); err != nil {
		slog.Error("failed to compile the tessellation control shader", "error", err)
		return
	}
	if err := compileShaderFile(&s.ShaderData, s.TessellationEvaluation, s.TessellationEvaluationFlags+addFlags); err != nil {
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
