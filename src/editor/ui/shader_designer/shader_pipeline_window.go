package shader_designer

import (
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
	"reflect"
	"strconv"
	"strings"
)

const (
	shaderPipeline = "editor/ui/shader_designer/shader_pipeline_window.html"
)

type ShaderPipelineColorBlendAttachments struct {
	BlendEnable         bool
	SrcColorBlendFactor string
	DstColorBlendFactor string
	ColorBlendOp        string
	SrcAlphaBlendFactor string
	DstAlphaBlendFactor string
	AlphaBlendOp        string
	ColorWriteMaskR     bool
	ColorWriteMaskG     bool
	ColorWriteMaskB     bool
	ColorWriteMaskA     bool
}

type ShaderPipelineData struct {
	Path                    string
	Topology                string
	PrimitiveRestart        bool
	DepthClampEnable        bool
	RasterizerDiscardEnable bool
	PolygonMode             string
	CullMode                string
	FrontFace               string
	DepthBiasEnable         bool
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
	RasterizationSamples    string
	SampleShadingEnable     bool
	MinSampleShading        float32
	AlphaToCoverageEnable   bool
	AlphaToOneEnable        bool
	ColorBlendAttachments   []ShaderPipelineColorBlendAttachments
	LogicOpEnable           bool
	LogicOp                 string
	BlendConstants0         float32
	BlendConstants1         float32
	BlendConstants2         float32
	BlendConstants3         float32
	DepthTestEnable         bool
	DepthWriteEnable        bool
	DepthCompareOp          string
	DepthBoundsTestEnable   bool
	StencilTestEnable       bool
	FrontFailOp             string
	FrontPassOp             string
	FrontDepthFailOp        string
	FrontCompareOp          string
	FrontCompareMask        uint32
	FrontWriteMask          uint32
	FrontReference          uint32
	BackFailOp              string
	BackPassOp              string
	BackDepthFailOp         string
	BackCompareOp           string
	BackCompareMask         uint32
	BackWriteMask           uint32
	BackReference           uint32
	MinDepthBounds          float32
	MaxDepthBounds          float32
	PatchControlPoints      string
}

func setupShaderPipelineDefaults() ShaderPipelineData {
	return ShaderPipelineData{
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
	win.pipelineDoc, _ = markup.DocumentFromHTMLAsset(&win.man, shaderPipeline,
		win.pipeline, map[string]func(*document.Element){
			"showTooltip":   showPipelineTooltip,
			"valueChanged":  win.pipeline.valueChanged,
			"pathChanged":   win.pipeline.pathChanged,
			"addAttachment": win.addAttachment,
			"savePipeline":  win.savePipeline,
		})
}

func showPipelineTooltip(e *document.Element) {
	if len(e.Children) < 2 {
		return
	}
	tip, ok := pipelineTooltips[e.Children[1].Attribute("id")]
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

func (p *ShaderPipelineData) pathChanged(e *document.Element) {
	p.Path = e.UI.ToInput().Text()
}

func (win *ShaderDesigner) addAttachment(e *document.Element) {
	win.pipeline.ColorBlendAttachments = append(
		win.pipeline.ColorBlendAttachments, ShaderPipelineColorBlendAttachments{
			BlendEnable:         true,
			SrcColorBlendFactor: "SrcAlpha",
			DstColorBlendFactor: "OneMinusSrcAlpha",
			ColorBlendOp:        "Add",
			SrcAlphaBlendFactor: "One",
			DstAlphaBlendFactor: "Zero",
			AlphaBlendOp:        "Add",
			ColorWriteMaskR:     true,
			ColorWriteMaskG:     true,
			ColorWriteMaskB:     true,
			ColorWriteMaskA:     true,
		})
	content := win.pipelineDoc.GetElementsByClass("topFields")[0]
	sy := content.UIPanel.ScrollY()
	win.reloadPipelineDoc()
	content = win.pipelineDoc.GetElementsByClass("topFields")[0]
	win.man.Host.RunAfterFrames(2, func() {
		content.UIPanel.SetScrollY(sy)
	})
}

func (p *ShaderPipelineData) valueChanged(e *document.Element) {
	id := e.Attribute("id")
	idx := -1
	sep := strings.Index(id, "_")
	if sep >= 0 {
		id = id[:sep]
		if i, err := strconv.Atoi(id[sep+1:]); err != nil {
			idx = i
		}
	}
	var v reflect.Value
	if idx >= 0 {
		v = reflect.ValueOf(p.ColorBlendAttachments[idx])
	} else {
		v = reflect.ValueOf(p)
	}
	field := v.Elem().FieldByName(id)
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

func (win *ShaderDesigner) savePipeline(*document.Element) {
	// TODO:  Save the pipeline
	panic("not yet implemented")
}
