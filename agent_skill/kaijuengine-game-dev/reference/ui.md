# UI System (`src/engine/ui/`)

A web-inspired UI with HTML/CSS-like layout, full event handling, and a markup
system. The UI renders on a separate orthographic camera (`host.Cameras.UI`) from
the main 3D scene. You can build UI imperatively (stylizers directly) or
declaratively (markup).

**Contents:** NO JavaScript runtime · Architecture · Manager · Creating elements
(+ element types) · Events · Layout · Panel / Label / Input operations · Dirty
flags · HTML/CSS markup · Complete example · Update cycle & camera.

## NO JavaScript runtime

`.go.html` files are **Go templates** parsed by `engine/ui/markup/document`
(`html_parser.go` uses `html/template` + a custom funcMap). An `onclick="someFunc"`
attribute maps to a **single** Go function name in the funcs map passed to
`CommonWorkspace.InitializeWithUI` (or equivalent). JS-style chains like
`onclick="setActiveTool(this); clickToolRaise()"` are **invalid**.

CSS is custom-parsed (`markup/css` + `Stylizer` + `ElementLayoutStylizer`). Manage
classes with `document.SetElementClasses(elm, "materialIcon", "active")`,
`elm.HasClass()`, `elm.ClassList()`, or `SetClasses`. UI elements (`ui.Panel`,
`ui.Button`, `ui.Label`, …) are backed by `engine.Entity` + custom layout/dirty
flags + Vulkan rendering. Active tool states use CSS classes like `.active` with
editor accent colors (`#682A2D` / `#881E1E`) plus Go-side class management in click
handlers (see `settings_workspace` for the `SetElementClasses` pattern). This
applies to ALL UI work in the engine and editor.

## Architecture

```
UI Manager (ui.Manager)
    └── UI Elements (Panel as base)
            ├── Button   ├── Label   ├── Input    ├── Image
            ├── Checkbox ├── Slider  ├── Select    └── ProgressBar

Markup System (ui/markup/)
    ├── document/  # DOM-like model
    ├── css/       # CSS parsing/styling
    └── spec_generator/  # code generation
```

## Manager

Create and initialize before use:

```go
import "kaijuengine.com/engine/ui"

man := ui.Manager{}
man.Init(host)          // REQUIRED before creating elements
button := man.Add().ToButton()
```

Methods: `man.Add()` (new element → `*UI`), `man.Remove(elm)`, `man.Clear()`,
`man.Reserve(100)`, `man.Hovered()`.

## Creating elements

All via the manager + a `To*` assertion, then `Init`:

```go
panel := man.Add().ToPanel();       panel.Init(nil, ui.ElementTypePanel)
button := man.Add().ToButton();     button.Init(texture, "Click Me")
label := man.Add().ToLabel();       label.Init("Hello World")
input := man.Add().ToInput();       input.Init("placeholder")
image := man.Add().ToImage();       image.Init(texture)
slider := man.Add().ToSlider();     slider.Init()
checkbox := man.Add().ToCheckbox(); checkbox.Init()
progress := man.Add().ToProgressBar(); progress.Init()
```

### Element types

| Type | Description | Base |
|------|-------------|------|
| `Panel` | Container, base for most elements | UI |
| `Button` | Clickable button with hover states | Panel |
| `Label` | Text display | UI |
| `Input` | Text input field | Panel |
| `Image` | Texture display | UI |
| `Checkbox` | Toggle checkbox | Panel |
| `Slider` | Value slider | Panel |
| `Select` | Dropdown selector | Panel |
| `ProgressBar` | Progress indicator | Panel |

## Events

```go
button.Base().AddEvent(ui.EventTypeClick, func() { /* ... */ })
button.Base().AddEvent(ui.EventTypeEnter, func() { /* hover in */ })
input.Base().AddEvent(ui.EventTypeChange, func() { /* value changed */ })
button.Base().RemoveEvent(ui.EventTypeClick, eventId)
```

Event types: `EventTypeEnter`, `EventTypeExit`, `EventTypeClick`,
`EventTypeRightClick`, `EventTypeDoubleClick`, `EventTypeDown`, `EventTypeUp`,
`EventTypeRightDown`, `EventTypeRightUp`, `EventTypeMiss`, `EventTypeDragStart`,
`EventTypeDragEnd`, `EventTypeDrop`, `EventTypeScroll`, `EventTypeChange`,
`EventTypeFocus`, `EventTypeBlur`, `EventTypeSubmit`, `EventTypeKeyDown`,
`EventTypeKeyUp`.

## Layout

```go
layout.SetPositioning(PositioningStatic)   // also Absolute, Fixed, Relative, Sticky
layout.Scale(width, height)                // ScaleWidth / ScaleHeight for one axis
layout.SetMargin(left, top, right, bottom)
layout.SetPadding(left, top, right, bottom)
layout.SetBorder(left, top, right, bottom)
layout.SetOffset(x, y)                      // for absolute positioning
layout.SetZ(z)

panel.FitContent()       // FitContentWidth / FitContentHeight / DontFitContent
```

## Panel operations

```go
panel.AddChild(child); panel.RemoveChild(child)
panel.SetColor(matrix.ColorWhite())
panel.SetBackground(texture)
panel.SetBorderRadius(tl, tr, br, bl)

panel.SetScrollDirection(PanelScrollDirectionVertical) // Horizontal / Both
panel.ScrollX(); panel.ScrollY(); panel.SetScrollY(50); panel.ResetScroll()
```

## Label operations

```go
label.SetText("New Text"); label.Text()
label.SetFontSize(16.0)
label.SetFontFace(rendering.FontRegular)
label.SetFontWeight("bold")  // normal, bold, bolder, lighter
label.SetFontStyle("italic") // normal, italic
label.SetJustify(rendering.FontJustifyLeft)  // Center, Right
label.SetBaseline(rendering.FontBaselineTop) // Center, Bottom
label.SetColor(matrix.ColorWhite()); label.SetBGColor(matrix.ColorBlack())
label.SetWrap(true)
label.SetWidthAutoHeight(width)
```

## Input operations

```go
input.Value(); input.SetValue("text")
input.Placeholder(); input.SetPlaceholder("...")
input.Base().AddEvent(ui.EventTypeSubmit, func() { /* enter key */ })
input.Base().AddEvent(ui.EventTypeChange, func() { /* value change */ })
```

## Dirty flags

```go
ui.SetDirty(DirtyTypeLayout)
ui.SetDirty(DirtyTypeResize)
ui.SetDirty(DirtyTypeGenerated)
ui.SetDirty(DirtyTypeColorChange)
```

## HTML/CSS markup

```go
import (
    "kaijuengine.com/engine/ui/markup"
    "kaijuengine.com/engine/ui/markup/document"
)

doc, err := markup.DocumentFromHTMLAsset(&man, "ui/main.html", nil, nil)
doc := markup.DocumentFromHTMLString(&man, htmlStr, cssStr, nil, nil, nil)

element, _ := doc.GetElementById("myButton")
element.UI.ToButton().Base().AddEvent(ui.EventTypeClick, handler)
rootPanel := doc.Elements[0].UI.ToPanel()
```

## Complete example

```go
func (g *Game) Launch(host *engine.Host) {
    g.host = host
    g.uiMan.Init(host)

    g.myPanel = g.uiMan.Add().ToPanel()
    g.myPanel.Init(nil, ui.ElementTypePanel)
    g.myPanel.SetColor(matrix.ColorGray().ScaleAlpha(0.8))
    g.myPanel.SetMargin(10, 10, 10, 10)
    g.myPanel.SetPadding(10, 10, 10, 10)

    g.myLabel = g.uiMan.Add().ToLabel()
    g.myLabel.Init("Hello, Kaiju!")
    g.myLabel.SetFontSize(24.0)
    g.myLabel.SetColor(matrix.ColorWhite())
    g.myPanel.AddChild(g.myLabel.Base())

    g.myButton = g.uiMan.Add().ToButton()
    tex, _ := host.TextureCache().Texture("square", rendering.TextureFilterLinear)
    g.myButton.Init(tex, "Click Me")
    g.myButton.Base().Layout().SetMargin(0, 10, 0, 0)
    g.myPanel.AddChild(g.myButton.Base())

    g.myButton.Base().AddEvent(ui.EventTypeClick, func() { g.myLabel.SetText("Button Clicked!") })
    g.myButton.Base().AddEvent(ui.EventTypeEnter, func() { g.myButton.SetColor(matrix.ColorGreen()) })
    g.myButton.Base().AddEvent(ui.EventTypeExit,  func() { g.myButton.SetColor(matrix.ColorWhite()) })
}
```

## Update cycle & camera

- **UIUpdater** runs before main game Update; **UILateUpdater** runs after it. The
  manager registers with these automatically in `Init()`.
- The UI renders to a separate orthographic camera: `host.Cameras.UI`.
