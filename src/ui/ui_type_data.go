package ui

import (
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/events"
)

/******************************************************************************/
// Label
/******************************************************************************/
type ColorRange struct {
	start int
	end   int
	hue   matrix.Color
	bgHue matrix.Color
}

type LabelData struct {
	text              string
	fontFace          string
	colorRanges       []ColorRange
	fontSize          float32
	lineHeight        float32
	overrideMaxWidth  float32
	fgColor           matrix.Color
	bgColor           matrix.Color
	justify           rendering.FontJustify
	baseline          rendering.FontBaseline
	style             rendering.FontStyle
	diffScore         int
	runeDrawings      []rendering.Drawing
	lastRenderWidth   float32
	unEnforcedFGColor matrix.Color
	unEnforcedBGColor matrix.Color
	isForcedFGColor   bool
	isForcedBGColor   bool
	wordWrap          bool
	renderRequired    bool
}

/******************************************************************************/
// Panel
/******************************************************************************/
type PanelScrollDirection = int32

const (
	PanelScrollDirectionNone       = PanelScrollDirection(0x00)
	PanelScrollDirectionVertical   = PanelScrollDirection(0x01)
	PanelScrollDirectionHorizontal = PanelScrollDirection(0x02)
	PanelScrollDirectionBoth       = PanelScrollDirection(0x03)
)

type BorderStyle = int32

const (
	BorderStyleNone = BorderStyle(iota)
	BorderStyleHidden
	BorderStyleDotted
	BorderStyleDashed
	BorderStyleSolid
	BorderStyleDouble
	BorderStyleGroove
	BorderStyleRidge
	BorderStyleInset
	BorderStyleOutset
)

type ContentFit = int32

const (
	ContentFitNone = ContentFit(iota)
	ContentFitWidth
	ContentFitHeight
	ContentFitBoth
)

type Overflow = int

const (
	OverflowScroll = Overflow(iota)
	OverflowVisible
	OverflowHidden
)

type PanelShaderType = int

const (
	PanelShaderTypeNine = PanelShaderType(iota)
	PanelShaderTypeImage
)

type PanelChildScrollEvent struct {
	down   events.Id
	scroll events.Id
}

type PanelRequestScroll struct {
	to        float32
	requested bool
}

type PanelData struct {
	scroll             matrix.Vec2
	offset             matrix.Vec2
	maxScroll          matrix.Vec2
	scrollSpeed        float32
	scrollDirection    PanelScrollDirection
	scrollEvent        events.Id
	borderStyle        [4]BorderStyle
	drawings           []rendering.Drawing
	fitContent         ContentFit
	requestScrollX     PanelRequestScroll
	requestScrollY     PanelRequestScroll
	overflow           Overflow
	enforcedColorStack []matrix.Color
	color              matrix.Color
	activateEvtId      events.Id
	deactivateEvtId    events.Id
	shaderType         PanelShaderType
	isScrolling        bool
	dragging           bool
	frozen             bool
	allowDragScroll    bool
}

/******************************************************************************/
// Button
/******************************************************************************/
type ButtonData struct {
	PanelData
	label *Label
}

/******************************************************************************/
// Checkbox
/******************************************************************************/
type CheckboxData struct {
	PanelData
	checked bool
}

/******************************************************************************/
// Input
/******************************************************************************/
type InputType = int32

const (
	InputTypeDefault = InputType(iota)
	InputTypeText
	InputTypeNumber
	InputTypePhone
	InputTypeDatetime
)

type InputData struct {
	PanelData
	label          *Label
	placeholder    *Label
	highlight      *Panel
	cursor         *Panel
	title          string
	description    string
	onUpDown       events.Event
	cursorOffset   int
	dragStartClick float32
	cursorBlink    float32
	labelShift     float32
	selectStart    int
	selectEnd      int
	dragStart      int
	inputType      InputType
	isActive       bool
	nextFocusInput *UI
}

/******************************************************************************/
// Select
/******************************************************************************/
type SelectData struct {
	PanelData
	label    *Label
	list     *Panel
	options  []string
	selected int
	isOpen   bool
}

/******************************************************************************/
// Slider
/******************************************************************************/
type SliderData struct {
	PanelData
	bar       *Panel
	value     float32
	dragging  bool
	draggable bool
}

/******************************************************************************/
// Sprite
/******************************************************************************/
type SpriteData struct {
	PanelData
	frameDelay   float32
	fps          float32
	currentFrame int
	paused       bool
}

/******************************************************************************/
// Type functions
/******************************************************************************/
func initTypeUI(elmType ElementType, ui *UI, construct interface{}) {
	switch elmType {
	case ElementTypeButton:
		initButton(ui, construct)
	case ElementTypeCheckbox:
		initCheckbox(ui, construct)
	case ElementTypeInput:
		initInput(ui, construct)
	case ElementTypeLabel:
		initLabel(ui, construct)
	case ElementTypePanel:
		initPanel(ui, construct)
	case ElementTypeSelect:
		initSelect(ui, construct)
	case ElementTypeSlider:
		initSlider(ui, construct)
	case ElementTypeSprite:
		initSprite(ui, construct)
	default:
		panic("missing element type initialization")
	}
}

func updateTypeUI(elmType ElementType, ui *UI, deltaTime float64) {
	switch elmType {
	case ElementTypeButton:
		panelUpdate(ui, deltaTime)
	case ElementTypeCheckbox:
		panelUpdate(ui, deltaTime)
	case ElementTypeInput:
		inputUpdate(ui, deltaTime)
	case ElementTypeLabel:
		uiUpdate(ui, deltaTime)
	case ElementTypePanel:
		panelUpdate(ui, deltaTime)
	case ElementTypeSelect:
		uiUpdate(ui, deltaTime)
	case ElementTypeSlider:
		sliderUpdate(ui, deltaTime)
	case ElementTypeSprite:
		spriteUpdate(ui, deltaTime)
	default:
		panic("missing element type initialization")
	}
}
