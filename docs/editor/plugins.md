# Editor Plugins

The Kaiju editor can be extended with **editor plugins** – small Go packages that are compiled together with the editor and can add new UI, tools, or modify the editor behaviour.

---

## Plugin folder layout

Each plugin lives in its own folder under the **plugins** directory (the location is resolved by `editor_plugin.PluginsFolder`). A valid plugin folder must contain two files:

* `plugin.json` – a JSON configuration describing the plugin.
* `plugin.go` – the Go entry‑point that implements the `editor_plugin.Plugin` interface.

The folder may contain additional source files, assets, etc., but the two files above are required for the editor to recognise the plugin.

---

## `plugin.json` schema

```json
{
	"Name": "My Awesome Plugin",
	"PackageName": "my_awesome_plugin",
	"Description": "A short description of what the plugin does",
	"Version": 0.001,
	"Author": "Your Name",
	"Website": "https://github.com/your/repo",
}
```

| Field | Type | Description |
|-------|------|-------------|
| **Name** | string | Human readable name shown in the *Plugin Settings* UI. |
| **PackageName** | string | The Go package name used for the plugin source files. Must be a valid identifier and unique across plugins. |
| **Description** | string | Short description displayed in the UI. |
| **Version** | number (float) | Plugin version – shown in the UI, not used by the engine. |
| **Author** | string | Author name. |
| **Website** | string (optional) | URL opened when the *Website* link is clicked in the UI. |

---

## `plugin.go` – the entry point

When a new plugin project is created via **Create Plugin Project** (see below) a stub file is generated from the constant `editorPluginGo` in `editor_plugin_manager.go`. The stub looks like this:

```go
package rename_me

import (
		"kaiju/editor"
		"kaiju/editor/editor_plugin"
)

// A unique key for the plugin – you can use a URL or any string that will not clash with other plugins.
const pluginKey = "https://github.com/KaijuEngine/kaiju"

type Plugin struct {}

func init() { editor.RegisterPlugin(pluginKey, &Plugin{}) }

func (p *Plugin) Launch(ed editor_plugin.EditorInterface) error {
		// TODO: implement plugin behaviour
		return nil
}
```

### Required parts

1. **Package name** – replace `rename_me` with the `PackageName` from `plugin.json`.
2. **`pluginKey`** – must be unique. A URL is a common choice but any string works.
3. **`init` function** – registers the plugin with the editor via `editor.RegisterPlugin`.
4. **`Launch` method** – receives an `EditorInterface` that gives access to the host, settings, events, project, history, etc. Implement your plugin logic here.

### `EditorInterface` methods

| Method | Returns | Description |
|--------|---------|-------------|
| `Host()` | `*engine.Host` | Low‑level host used for rendering, input, etc. |
| `BlurInterface()` / `FocusInterface()` | – | Called when the editor loses or gains focus. |
| `Settings()` | `*editor_settings.Settings` | Access to the editor's global settings. |
| `Events()` | `*editor_events.EditorEvents` | Subscribe to editor events (e.g., file opened, stage changed). |
| `History()` | `*memento.History` | Undo/redo history. |
| `Project()` | `*project.Project` | The currently loaded project. |
| `ProjectFileSystem()` | `*project_file_system.FileSystem` | File system abstraction for the project. |
| `StageView()` | `*editor_stage_view.StageView` | UI view of the current stage. |

---

## Creating a new plugin project

The editor UI provides a **Create plugin project** command (found under the *Kaiju engine logo* popup). Internally it calls `editor_plugin.CreatePluginProject(path)` which:

1. Creates an empty folder.
2. Writes a default `plugin.json`.
3. Writes the stub `plugin.go`.

You can also invoke the function programmatically:

```go
import "kaiju/editor/editor_plugin"

err := editor_plugin.CreatePluginProject("C:/path/to/your/plugin")
```

---

## Enabling / disabling plugins

Plugins are listed in the **Plugin Settings** pane of the editor (accessed via the *Plugin Settings* button in the workspace UI). Each entry shows the name, description, author, version and a checkbox to enable the plugin.

After toggling a checkbox you must click **Recompile Editor**. The editor will rebuild itself with the selected plugins compiled in. The UI prevents recompilation if no plugin state has changed.

---

## Loading plugins at runtime

When the editor starts it loads all plugins whose `Enabled` flag is `true` in their `plugin.json`. The registration performed in the plugin’s `init` function adds the plugin to the global registry. The editor then calls the plugin’s `Launch` method, passing the `EditorInterface`.

---

## Example: a minimal "Hello World" plugin

```go
package hello

import (
		"fmt"
		"kaiju/editor"
		"kaiju/editor/editor_plugin"
)

const pluginKey = "github.com/yourname/kaiju-hello"

type Plugin struct{}

func init() { editor.RegisterPlugin(pluginKey, &Plugin{}) }

func (p *Plugin) Launch(ed editor_plugin.EditorInterface) error {
		// Print a message to the console when the plugin is loaded.
		fmt.Println("Hello from Kaiju editor plugin!")
		// You could also add UI elements, register shortcuts, etc.
		return nil
}
```

Corresponding `plugin.json`:

```json
{
	"Name": "Hello Plugin",
	"PackageName": "hello",
	"Description": "Prints a greeting when the editor starts.",
	"Version": 0.001,
	"Author": "Your Name",
	"Website": "",
}
```

Place both files in a folder under the editor’s `plugins` directory, enable it in the UI, and recompile the editor.

---

## Frequently asked questions

* **Do I need to rebuild the editor for every change?**
	Yes. Plugins are compiled into the editor binary. After editing source code you must click **Recompile Editor**.
* **Can a plugin be disabled without recompiling?**
	No. The `Enabled` flag is read at compile time. Disabling a plugin requires a recompilation.
* **Where are the compiled plugins stored?**
	They become part of the `kaiju.exe` binary generated by the `go build` task.

---

## Files to review for more details

| File | Description |
|------|-------------|
| `menu_bar.go.html` | Creating plugin project |
| `editor_plugin_manager.go` | Functions for creating plugin folders, config files, and entry point generation |
| `editor_plugin.go` | Definition of the EditorPlugin interface and EditorInterface methods |
| `settings_workspace.go.html` | HTML UI layout for the Plugin Settings workspace pane |
| `settings_workspace.go` | Go implementation handling plugin UI, toggling, and recompilation logic |
