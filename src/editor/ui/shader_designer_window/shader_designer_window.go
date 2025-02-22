package shader_designer_window

import (
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
)

const (
	shaderPipeline = "editor/ui/shader_designer/shader_pipeline_window.html"
)

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
	DepthBiasConstantFactor string
	DepthBiasClamp          string
	DepthBiasSlopeFactor    string
	LineWidth               string
	RasterizationSamples    string
	SampleShadingEnable     bool
	MinSampleShading        string
	AlphaToCoverageEnable   bool
	AlphaToOneEnable        bool
	LogicOpEnable           bool
	LogicOp                 string
	BlendConstants0         string
	BlendConstants1         string
	BlendConstants2         string
	BlendConstants3         string
	DepthTestEnable         bool
	DepthWriteEnable        bool
	DepthCompareOp          string
	DepthBoundsTestEnable   bool
	StencilTestEnable       bool
	FrontFailOp             string
	FrontPassOp             string
	FrontDepthFailOp        string
	FrontCompareOp          string
	FrontCompareMask        string
	FrontWriteMask          string
	FrontReference          string
	BackFailOp              string
	BackPassOp              string
	BackDepthFailOp         string
	BackCompareOp           string
	BackCompareMask         string
	BackWriteMask           string
	BackReference           string
	MinDepthBounds          string
	MaxDepthBounds          string
	PatchControlPoints      string
}

type ShaderDesigner struct {
	pipeline    ShaderPipelineData
	pipelineDoc *document.Document
}

func uiInit(host *engine.Host) {
	uiMan := ui.Manager{}
	uiMan.Init(host)
	win := ShaderDesigner{}
	win.pipelineDoc, _ = markup.DocumentFromHTMLAsset(&uiMan, shaderPipeline,
		win.pipeline, map[string]func(*document.Element){
			"showTooltip": showPipelineTooltip,
		})
	//win.pipelineDoc.Deactivate()
}

func New() {
	container := host_container.New("Shader Designer", nil)
	go container.Run(640, 480, -1, -1)
	<-container.PrepLock
	container.RunFunction(func() { uiInit(container.Host) })
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
