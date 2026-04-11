---
title: Writing UI (HTML/CSS) | Kaiju Engine
---

# Writing UI (HTML/CSS)
There are only 2 primitive UI elements in the engine. (1) a panel, and (2) a label. From these 2 primitives all UI is created in the engine. You can manually create these elements yourself, but the easiest way to create UI is through the use of HTML/CSS templates. This is the preferred method for creating UI in Kaiju Engine.

## HTML/CSS
To begin creating UI, first create a `.html` file inside of the `content/ui` folder or any subfolder of that folder. We'll start with an example named `binding.html` which we place into a `tests` folder (`content/ui/tests/binding.html`).

```html
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" type="text/css" href="ui/tests/binding_style.css">
	</head>
	<body>
		<div class="container">
			<div id="nameList">
				{{range .EntityNames}}
					<div>Entity: {{.}}</div>
				{{end}}
			</div>
		</div>
	</body>
</html>
```

Here you'll notice a few things, the path to the `.css` stylesheet to use for this document, and some Go template syntax for binding a slice of strings named `EntityNames`. So this template expects something similar to the following structure as binding data:

```go
type BindingData struct {
	EntityNames []string
}
```

Next, you'll want to create the `.css` file since it was referenced in the HTML head. Create a file named `binding_style.css` in the same folder as the `.html` file (`content/ui/tests/binding_style.css`).

```css
body {
	padding: 0;
	margin: 0;
}
.content {
	position: absolute;
	top: 0;
	height: 300px;
	background-color: #000;
	padding: 10px;
	border-bottom: 1px solid white;
	z-index: 100;
}
#nameList {
	padding: 0;
	height: calc(100% - 32px);
	overflow-y: scroll;
	color: white;
}
```

At this point your UI is complete, though probably not what you'd consider pretty.

## Go
To load up this UI in Go, you'll have access to the host, and you'll need to call `DocumentFromHTMLAsset` and provide the path to your HTML file.

```go
data := struct{
	EntityNames []string
}{
	[]string{"Entity1", "Entity2", "Entity3"},
}
doc, err := markup.DocumentFromHTMLAsset(host, "ui/tests/binding.html", data, nil)
```

This will load up the HTML document and any of the CSS it references and build out your UI. The returned `doc` will contain the document and all the elements/panels/labels. This UI is immediately loaded into the `host` so you don't need to worry about doing that yourself. *The last argument is a funcmap used for inline template functions*
## Dynamic UI Updates (Go Interoperability)
While HTML/CSS provides the initial structural foundation and styling (the canvas initial geometric bounding box), you will commonly need to update the UI dynamically during runtime from your game's Go logic (e.g., updating a timer, health bar, or coordinate tracker).

### 1. Element Hierarchy (Panel vs Label)
When declaring a simple markup with text in HTML:
`html
<div id="fps-val">FPS: --</div>
`
The parsing engine instantiates a ui.UIPanel base structural wrapper, and creates an attached hierarchical child ui.UILabel for the string text logic. 
To retrieve and modify properties dynamically in the Go loop:
*   **Backgrounds & Solid Borders**: Are applied to the parent pane (element.UIPanel.SetBGColor(...)).
*   **Text Strings & Typography Colors**: Are applied to the inner textual label (element.InnerLabel().SetText(...) or element.InnerLabel().SetColor(...)).

### 2. Dirty Flags (Reactivity is NOT Automatic)
Unlike standard web browsers, the Kaiju Engine does NOT automatically redraw the UI canvas frame when you modify an element's property via backend Go scripts. You **must explicitly** flag the element as "dirty" in your game or system's update loop to force the Vulkan render pipeline to redraw the updated geometry on the screen:
`go
if e, ok := doc.GetElementById("fps-val"); ok {
    if lbl := e.InnerLabel(); lbl != nil {
        lbl.SetText(fmt.Sprintf("FPS: %d", fps))
    }
    // Critical: Signal the renderer that this specific element layout has mutated!
    e.UI.SetDirty(ui.DirtyTypeLayout) 
}
`

### 3. Layout Control Pitfalls (The Invisible Canvas Bug)
**Never strip foundational HTML positioning anchors in an attempt to control the Screen Layout purely via Go Transform functions.**

The document parser deeply relies on CSS properties (position: absolute, 	op, ottom, ight, width, height) to anchor the UI primitives to the Screen Viewport properly. 

If you declare empty/naked <div> tags in the HTML and attempt to explicitly align and mathematically scale them exclusively via Transform.SetLocalPosition(...) or manual offset matrices natively in Go, the engine's internal UI-Layout System will fight the DOM logic. This typically cascades in elements infinitely collapsing to coordinates (0,0), or losing their mesh dimension entirely, making your UI visually disappear from the screen logic.
*   **Golden Rule**: HTML/CSS solely owns the global layout spatial anchors, responsivity bounds, and z-index ordering. Go owns dynamic text injection, loop data binding, and situational state-color updates.
