---
title: Kaiju Engine | Writing UI (HTML/CSS)
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