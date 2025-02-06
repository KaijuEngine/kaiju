---
title: Kaiju Engine | Preview UI (Live)
---

# Preview UI (Live)
Since the engine UI is built using HTML and CSS to build out the interface, we don't provide any graphical wysiwyg tools at the moment. To ease the process of designing your UI though, we've added the ability to have a live preview of your UI as you write your HTML and CSS code.

## Live preview
To activate the live preview, open up the console window (&#96;) and type `preview path/to/file.html`. So if we are going off of the example we gave in [Writing](writing.md), then we'd type `preview content/ui/tests/binding.html` into the console and hit enter. This will load up the preview in a separate window for us to monitor. The preview updates every time there is a change to the file (each time the HTML file is saved).

## Binding dummy data
You'll often need to preview some dummy data in your HTML UI, to do this simply create a `.json` file next to your HTML file. So in the case of `content/ui/tests/binding.html` we would have a `content/ui/tests/binding.html.json` file. This file should contain the data you want to bind to your HTML file for the preview. For example, if we wanted to bind a list of entity names to our HTML file, we would create a `content/ui/tests/binding.html.json` file with the following content:

```json
{
	"EntityNames": ["Entity1", "Entity2", "Entity3"]
}
```