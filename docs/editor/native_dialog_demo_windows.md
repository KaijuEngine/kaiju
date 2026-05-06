# Windows Native Dialog

The editor currently includes Windows-only native dialog when built with the `filedialog` tag (for example: `-tags="editor,filedialog"`).

## Summary

- Dialog requests execute on a worker thread
- Completed dialog results are queued and processed on the window polling thread
- Callback processing is driven by `Window.Poll()` through `filesystem.ProcessDialogCallbacks()`
- If `Root` is set, folder navigation is blocked outside that root and accepted selections are revalidated before results are returned
- Start folder rule:
`CurrentDirectory` is used only when it is inside `Root`; otherwise the dialog starts from `Root`

## Usage Snippets

Use these snippets from game/editor code where you already have `host *engine.Host`.

### Open File Dialog

`ok` callback is required for wrapper helpers; `cancel` is optional.

```go
import "kaijuengine.com/platform/filesystem"

err := host.Window.OpenFileDialog(
	"", // startPath: initial directory (empty lets OS choose a default)
	[]filesystem.DialogExtension{
		{Name: "Go Files", Extension: ".go"},
		{Name: "All Files", Extension: ".*"}, // maps to *.*
	},
	func(path string) { // required ok callback - called when user confirms selection
		println("selected file:", path)
	},
	nil, // optional cancel callback
)
if err != nil {
	// Immediate setup/launch failure (not a user cancel)
	println("open file dialog failed:", err.Error())
}
```

### Save File Dialog

`ok` callback is required; `cancel` is optional.

```go
import "kaijuengine.com/platform/filesystem"

err := host.Window.SaveFileDialog(
	"",           // startPath: initial directory
	"output.txt", // default save file name
	[]filesystem.DialogExtension{
		{Name: "Text Files", Extension: ".txt"},
		{Name: "All Files", Extension: ".*"},
	},
	func(path string) { // required ok callback
		println("save path:", path)
	},
	func() { // optional cancel callback
		println("save canceled or failed")
	},
)
if err != nil {
	println("save file dialog failed:", err.Error())
}
```

### Open Folder Dialog

`ok` callback is required; `cancel` is optional.

```go
err := host.Window.OpenFolderDialog(
	"", // startPath: initial folder
	func(path string) { // required ok callback
		println("selected folder:", path)
	},
	func() { // optional cancel callback
		println("folder selection canceled or failed")
	},
)
if err != nil {
	println("open folder dialog failed:", err.Error())
}
```

### Advanced Native Dialog Request

```go
import "kaijuengine.com/platform/filesystem"

root, err := filesystem.GameDirectory()
if err != nil {
	// Fallback: no explicit root/current directory constraints
	root = ""
}

req := filesystem.NativeDialogRequest{
	Mode:             filesystem.NativeDialogModeOpenFiles, // multi-select files
	Title:            "Import Assets",
	CurrentDirectory: root,
	Root:             root, // navigation + final-selection constraint
	ShowHidden:       true,
	Filters: []filesystem.DialogFilter{
		{Name: "Go and Text Files", Patterns: []string{"*.go", "*.txt"}},
		{Name: "Images", Patterns: []string{"*.png", "*.jpg", "*.jpeg"}},
		{Name: "All Files", Patterns: []string{"*.*"}},
	},
	Options: []filesystem.DialogCustomOption{
		{Name: "Recursive import", Default: 1}, // checkbox option
		{Name: "Import mode", Values: []string{"Copy", "Reference", "Link"}, Default: 0}, // combo option
	},
	WindowHandle: host.Window.PlatformWindow(),
}

err = filesystem.OpenNativeDialogWindow(req, func(result filesystem.NativeDialogResult) {
	switch result.Status {
	case filesystem.NativeDialogStatusAccepted:
		println("accepted files:", len(result.Paths))
		println("selected filter index:", result.SelectedFilterIndex)
		// If selected options are not returned by the native layer, Kaiju fills defaults from req.Options.
		println("selected options keys:", len(result.SelectedOptions))
	case filesystem.NativeDialogStatusCancel:
		println("dialog canceled")
	case filesystem.NativeDialogStatusFailed:
		println("dialog failed")
	}
})
if err != nil {
	println("advanced dialog failed:", err.Error())
}
```
