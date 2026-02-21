---
title: Getting Started Without Editor | Kaiju Engine
description: Guide to building and running a Kaiju game using only Go code, without the graphical editor.
keywords: Kaiju, getting started, no editor, Go, build, run, engine
---

# Getting started without the editor (pure code)

This guide explains how to build and run a Kaiju game **without** using the graphical editor. The workflow relies on the `main.test.go` file, which is compiled when the **`editor` build tag is *not* present. By following the steps below you can create a game from scratch using only Go code.

---
## Prerequisites

* **Go 1.22+** installed and available on your `PATH`.
* A recent **Git** client to clone the repository.
* A C compiler is required because the engine relies on CGo.

---
## 1. Clone the repository

```powershell
git clone https://github.com/kaijuengine/kaiju.git
cd kaiju/src
```

---
## 2. Build the engine *without* the `editor` tag

The file `src/main.test.go` contains the build constraint:

```go
//go:build !editor
```

This means it is compiled **only** when the `editor` tag is **absent**. The default VS Code tasks in this repository build with `-tags=editor`. To build a binary that runs the pure‑code path, invoke `go build` **without** that flag:

```powershell
# Build from the src directory and output the executable one level up
go build .
```

The resulting `kaiju` executable lives in the repository root.

---
## 3. Run the binary

The executable name differs by platform. After building, you will have either `kaiju.exe` on Windows or `kaiju` on macOS/Linux.

* **Windows**
	```powershell
	.\kaiju.exe
	```
* **macOS / Linux**
	```bash
	./kaiju
	```

On the first run the engine will detect that the `game_content` directory is missing and will automatically copy the stock assets from the embedded editor content (see **Step 4**). After the copy completes the game window will appear, showing a simple rotating sphere – the default demo implemented in `main.test.go`.

---
## 4. What `main.test.go` does – a walkthrough

The file provides a minimal, fully‑functional game implementation. Below is a detailed description of each required piece so you can recreate it in your own project.

### 4.1 Package imports

```go
import (
	"kaiju/bootstrap"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)
```

These packages give you access to the engine core, asset handling, math utilities, logging, and the standard library for file I/O.

### 4.2 Constants – paths used by the engine

```go
const rawContentPath = `editor/editor_embedded_content/editor_content`
const gameContentPath = `game_content`
```

* `rawContentPath` points to the embedded editor assets shipped with the repository.
* `gameContentPath` is the directory the engine expects to find game‑specific assets (textures, meshes, shaders, etc.). If it does not exist, the engine will call `gameCopyEditorContent()` to populate it with the default assets.

### 4.3 The `Game` type

```go
type Game struct {
	host *engine.Host
	ball *engine.Entity
}
```

* `host` gives you access to the engine subsystems (rendering, material cache, etc.).
* `ball` is a simple entity we create to demonstrate rendering.

### 4.4 Required interface methods

Kaiju expects a type that implements the `bootstrap.GameInterface`. The following methods satisfy that contract:

* **`PluginRegistry()`** – returns a slice of plugin types. The example returns an empty slice because no custom plugins are needed.
* **`ContentDatabase()`** – creates (or copies) the asset database. It checks for `gameContentPath` and calls `gameCopyEditorContent()` if missing, then returns `assets.NewFileDatabase(gameContentPath)`.
* **`Launch(host *engine.Host)`** – called once the engine is ready. Here we:
	1. Store the host.
	2. Create a sphere mesh.
	3. Retrieve a basic shader and material.
	4. Build a `rendering.Drawing` that ties the mesh, material, and shader data together.
	5. Register the drawing with the host and schedule the `update` method to be called each frame.

### 4.5 Updating your game
Notice the **`update(deltaTime float64)`** function, it is a helper function (not part of the `GameInterface`) that is registered with the host updater to run each frame. It provides a simple animation that moves the sphere in a sinusoidal pattern. This function is registered with the `host.Updater` in the `Launch`.

### 4.6 Helper functions

* **`getGame()`** – required by the bootstrap package; it returns a pointer to a `Game` instance.
* **`gameCopyEditorContent()`** – copies the default editor assets into `game_content`. It walks the `editor/editor_embedded_content/editor_content` directory, skips the `editor` and `meshes` sub‑folders, and writes each file to the target directory.

---
## 5. Creating your own game from scratch

1. **Create a new Go file** (e.g., `mygame.go`) in the `src` folder.
2. **Define a struct** that holds any state you need (similar to `Game`).
3. **Implement the four interface methods** listed in section 4.4. You can reuse most of the example code and replace the sphere with your own entities, shaders, or assets.
4. **Add your assets** to a new folder (e.g., `my_game_content`). Update the constants accordingly:

```go
const rawContentPath = `editor/editor_embedded_content/editor_content`
const gameContentPath = `my_game_content`
```

5. **Build and run** the binary exactly as in Step 2 and Step 3.
   
	**Important:** If you create your own game implementation, you should either replace or delete the existing `src/main.test.go` file to avoid the example code being compiled alongside your own code.

---
## 6. Frequently asked questions

| Question | Answer |
|---|---|
| *Do I need the editor at all?* | No. The engine can run completely head‑less; the editor is only a convenience for asset editing.
| *Do I need the stock engine content?* | Yes. The engine expects the default assets (textures, shaders, etc.) that are shipped with the repository. They are copied automatically on first run if `game_content` is missing.
## 7. Custom asset database (optional)

The engine uses an `assets.Database` implementation to load textures, meshes, shaders, and other resources. The example uses the built‑in file‑based database (`assets.NewFileDatabase`). If you prefer a different storage mechanism (e.g., embedded assets, network‑based loading, or a custom format), you can implement the `assets.Database` interface yourself and return it from `ContentDatabase()`.

Typical steps for a custom database:
1. Create a type that satisfies the methods defined in `kaiju/engine/assets/database.go`.
2. Implement asset lookup, loading, and any caching you need.
3. In `ContentDatabase()`, return an instance of your custom type instead of `assets.NewFileDatabase`.

This flexibility allows you to integrate the engine with existing pipelines or package assets in a way that best fits your project.
| *What if I want to keep the editor assets but add my own?* | Place your custom assets in `game_content` (or a sub‑folder) and reference them by path in your code. The engine will prioritize files in `game_content` over the embedded defaults.
