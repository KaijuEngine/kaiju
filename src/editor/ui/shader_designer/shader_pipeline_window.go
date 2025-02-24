package shader_designer

import (
	"encoding/json"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const (
	shaderPipelineHTML = "editor/ui/shader_designer/shader_pipeline_window.html"
)

type shaderPipelineHTMLData struct {
	rendering.ShaderPipelineData
}

type flagState struct {
	List    []string
	Current []string
	Array   string
	Field   string
	Index   int
}

func (s flagState) Has(val string) bool {
	return slices.Contains(s.Current, val)
}

func (d shaderPipelineHTMLData) ColorWriteMaskFlagState(index int, a rendering.ShaderPipelineColorBlendAttachments) flagState {
	return flagState{
		List:    klib.MapKeys(rendering.StringVkColorComponentFlagBits),
		Current: a.ColorWriteMask,
		Array:   "ColorBlendAttachments",
		Field:   "ColorWriteMask",
		Index:   index,
	}
}

func setupShaderPipelineDefaults() rendering.ShaderPipelineData {
	return rendering.ShaderPipelineData{
		Topology:                "Triangles",
		PrimitiveRestart:        false,
		DepthClampEnable:        false,
		RasterizerDiscardEnable: false,
		PolygonMode:             "Fill",
		CullMode:                "Back",
		FrontFace:               "Clockwise",
		DepthBiasEnable:         false,
		DepthBiasConstantFactor: 0,
		DepthBiasClamp:          0,
		DepthBiasSlopeFactor:    0,
		LineWidth:               1,
		RasterizationSamples:    "1Bit",
		SampleShadingEnable:     true,
		MinSampleShading:        0.2,
		AlphaToCoverageEnable:   false,
		AlphaToOneEnable:        false,
		LogicOpEnable:           false,
		LogicOp:                 "Copy",
		BlendConstants0:         0,
		BlendConstants1:         0,
		BlendConstants2:         0,
		BlendConstants3:         0,
		DepthTestEnable:         true,
		DepthWriteEnable:        false,
		DepthCompareOp:          "Less",
		DepthBoundsTestEnable:   false,
		StencilTestEnable:       false,
		FrontFailOp:             "Keep",
		FrontPassOp:             "Keep",
		FrontDepthFailOp:        "Keep",
		FrontCompareOp:          "Never",
		FrontCompareMask:        0,
		FrontWriteMask:          0,
		FrontReference:          0,
		BackFailOp:              "Keep",
		BackPassOp:              "Keep",
		BackDepthFailOp:         "Keep",
		BackCompareOp:           "Never",
		BackCompareMask:         0,
		BackWriteMask:           0,
		BackReference:           0,
		MinDepthBounds:          0,
		MaxDepthBounds:          0,
		PatchControlPoints:      "Triangles",
	}
}

func setupPipelineDoc(win *ShaderDesigner, man *ui.Manager) {
	win.pipeline = setupShaderPipelineDefaults()
	win.reloadPipelineDoc()
	//win.pipelineDoc.Deactivate()
}

func (win *ShaderDesigner) reloadPipelineDoc() {
	if win.pipelineDoc != nil {
		win.pipelineDoc.Destroy()
	}
	data := shaderPipelineHTMLData{win.pipeline}
	win.pipelineDoc, _ = markup.DocumentFromHTMLAsset(&win.man, shaderPipelineHTML,
		data, map[string]func(*document.Element){
			"showTooltip":      showPipelineTooltip,
			"valueChanged":     win.pipelineValueChanged,
			"nameChanged":      win.pipelineNameChanged,
			"addAttachment":    win.pipelineAddAttachment,
			"deleteAttachment": win.pipelineDeleteAttachment,
			"savePipeline":     win.pipelineSave,
		})
}

func showPipelineTooltip(e *document.Element) {
	if len(e.Children) < 2 {
		return
	}
	id := e.Children[1].Attribute("data-field")
	if id == "" {
		id = e.Attribute("name")
	}
	if sep := strings.Index(id, "_"); sep >= 0 {
		id = id[:sep]
	}
	tip, ok := pipelineTooltips[id]
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

func (win *ShaderDesigner) pipelineNameChanged(e *document.Element) {
	win.pipeline.Name = e.UI.ToInput().Text()
}

func (win *ShaderDesigner) pipelineAddAttachment(e *document.Element) {
	win.pipeline.ColorBlendAttachments = append(
		win.pipeline.ColorBlendAttachments, rendering.ShaderPipelineColorBlendAttachments{
			BlendEnable:         true,
			SrcColorBlendFactor: "SrcAlpha",
			DstColorBlendFactor: "OneMinusSrcAlpha",
			ColorBlendOp:        "Add",
			SrcAlphaBlendFactor: "One",
			DstAlphaBlendFactor: "Zero",
			AlphaBlendOp:        "Add",
			ColorWriteMask:      []string{"R", "G", "B", "A"},
		})
	content := win.pipelineDoc.GetElementsByClass("topFields")[0]
	sy := content.UIPanel.ScrollY()
	win.reloadPipelineDoc()
	content = win.pipelineDoc.GetElementsByClass("topFields")[0]
	win.man.Host.RunAfterFrames(2, func() {
		content.UIPanel.SetScrollY(sy)
	})
}

func (win *ShaderDesigner) pipelineDeleteAttachment(e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this attachment entry? The action currently can't be undone.", "Yes", "No", win.man.Host)
	if ok {
		idxString := e.Attribute("data-index")
		idx, _ := strconv.Atoi(idxString)
		win.pipeline.ColorBlendAttachments = slices.Delete(
			win.pipeline.ColorBlendAttachments, idx, idx+1)
		content := win.pipelineDoc.GetElementsByClass("topFields")[0]
		sy := content.UIPanel.ScrollY()
		win.reloadPipelineDoc()
		win.man.Host.RunAfterFrames(2, func() {
			content.UIPanel.SetScrollY(sy)
		})
	}
}

func (win *ShaderDesigner) pipelineValueChanged(e *document.Element) {
	fieldName := e.Attribute("data-field")
	arrayName := e.Attribute("data-array")
	var v reflect.Value
	v = reflect.ValueOf(&win.pipeline).Elem()
	if arrayName != "" {
		// TODO:  Make this safer by checking bounds and index
		idxString := e.Attribute("data-index")
		idx, _ := strconv.Atoi(idxString)
		v = v.FieldByName(arrayName)
		v = v.Addr()
		v = v.Elem().Index(idx)
	}
	field := v.FieldByName(fieldName)
	if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
		// TODO:  Ensure switch e.UI.Type() == ui.ElementTypeCheckbox
		add := e.UI.ToCheckbox().IsChecked()
		str := e.Attribute("name")
		var slice []string
		if !field.IsNil() {
			slice = field.Interface().([]string)
		} else {
			slice = []string{}
		}
		if add {
			for _, s := range slice {
				if s == str {
					return // Already exists, no change
				}
			}
			slice = append(slice, str)
		} else {
			for i, s := range slice {
				if s == str {
					slice = slices.Delete(slice, i, i+1)
					break
				}
			}
		}
		field.Set(reflect.ValueOf(slice))
	} else {
		var val reflect.Value
		switch e.UI.Type() {
		case ui.ElementTypeInput:
			res := klib.StringToTypeValue(field.Type().String(), e.UI.ToInput().Text())
			val = reflect.ValueOf(res)
		case ui.ElementTypeSelect:
			val = reflect.ValueOf(e.UI.ToSelect().Value())
		case ui.ElementTypeCheckbox:
			val = reflect.ValueOf(e.UI.ToCheckbox().IsChecked())
		}
		field.Set(val)
	}
}

func OpenPipeline(path string) {
	setup(func(win *ShaderDesigner) {
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to load the shader pipeline file", "file", path, "error", err)
			return
		}
		if err := json.Unmarshal(data, &win.pipeline); err != nil {
			slog.Error("failed to unmarshal the shader pipeline data", "error", err)
			return
		}
		win.reloadPipelineDoc()
	})
}

func (win *ShaderDesigner) pipelineSave(e *document.Element) {
	const saveRoot = "content/shaders/pipelines"
	if err := os.MkdirAll(saveRoot, os.ModePerm); err != nil {
		slog.Error("failed to create the shader pipeline folder",
			"folder", saveRoot, "error", err)
	}
	path := filepath.Join(saveRoot, win.pipeline.Name+editor_config.FileExtensionShaderPipeline)
	if _, err := os.Stat(path); err == nil {
		ok := <-alert.New("Overwrite?", "You are about to overwrite a shader pipeline with the same name, would you like to continue?", "Yes", "No", win.man.Host)
		if !ok {
			return
		}
	}
	res, err := json.Marshal(win.pipeline)
	if err != nil {
		slog.Error("failed to marshal the pipeline data", "error", err)
		return
	}
	if err := os.WriteFile(path, res, os.ModePerm); err != nil {
		slog.Error("failed to write the pipeline data to file", "error", err)
		return
	}
	slog.Info("shader pipeline successfully saved", "file", path)
	// TODO:  Show an in-window popup for prompting that things saved
	if len(e.Children) > 0 {
		u := e.Children[0].UI
		if u.IsType(ui.ElementTypeLabel) {
			u.ToLabel().SetText("File saved!")
		}
	}
}
