package shader_designer

import (
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
	DepthBiasConstantFactor float32
	DepthBiasClamp          float32
	DepthBiasSlopeFactor    float32
	LineWidth               float32
	RasterizationSamples    string
	SampleShadingEnable     bool
	MinSampleShading        float32
	AlphaToCoverageEnable   bool
	AlphaToOneEnable        bool
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

func setupPipelineDoc(win *ShaderDesigner, man *ui.Manager) {
	win.pipeline = setupShaderPipelineDefaults()
	win.pipelineDoc, _ = markup.DocumentFromHTMLAsset(man, shaderPipeline,
		win.pipeline, map[string]func(*document.Element){
			"showTooltip": showPipelineTooltip,
		})
	//win.pipelineDoc.Deactivate()
}
