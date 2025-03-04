package shader_designer

import (
	"encoding/json"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
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
				vert = append(vert, filepath.Join(shaderSrcFolder, f.Name()))
			case ".frag":
				frag = append(frag, filepath.Join(shaderSrcFolder, f.Name()))
			case ".geom":
				geom = append(geom, filepath.Join(shaderSrcFolder, f.Name()))
			case ".tesc":
				tesc = append(tesc, filepath.Join(shaderSrcFolder, f.Name()))
			case ".tese":
				tese = append(tese, filepath.Join(shaderSrcFolder, f.Name()))
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

func OpenShader(path string) {
	setup(func(win *ShaderDesigner) {
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to load the shader file", "file", path, "error", err)
			return
		}
		if err := json.Unmarshal(data, &win.shader); err != nil {
			slog.Error("failed to unmarshal the shader data", "error", err)
			return
		}
		win.ShowShaderWindow()
	})
}

func compileShaderFile(s *rendering.ShaderData, src, flags string) error {
	if src == "" {
		return nil
	}
	out := s.CompileVariantName(src, flags)
	args := []string{src, "-o", out}
	if flags != "" {
		args = append(args, strings.TrimSpace(flags))
	}
	compileCmd := exec.Command("glslc", args...)
	return compileCmd.Run()
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
	path := filepath.Join(saveRoot, win.shader.Name+editor_config.FileExtensionShader)
	if _, err := os.Stat(path); err == nil {
		ok := <-alert.New("Overwrite?", "You are about to overwrite a shader with the same name, would you like to continue?", "Yes", "No", win.man.Host)
		if !ok {
			return
		}
	}
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
