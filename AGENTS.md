# Kaiju Engine - Agent Documentation

This file provides essential information for AI agents working with the Kaiju Engine codebase.

## Project Overview

Kaiju Engine is a 2D/3D game engine written in Go (Golang) backed by Vulkan. It features a custom-built math library, hierarchical entity system, and modular rendering architecture. The engine is cross-platform (Windows, Linux, Mac, Android).

- **Module**: `kaijuengine.com`
- **Go Version**: 1.25.0+
- **Build Tags**: `debug`, `editor`, platform-specific (`.windows.go`, `.darwin.go`, `.linux.go`, `.android.go`)

## Project Structure

```
kaiju/
├── src/
│   ├── bootstrap/           # Game initialization and bootstrapping
│   │   ├── bootstrap.go     # Main bootstrap entry point
│   │   └── game_interface.go # GameInterface that games must implement
│   │
│   ├── engine/              # Core engine systems
│   │   ├── host.go          # Central runtime mediator (Host)
│   │   ├── entity.go        # Game entities with Transform
│   │   ├── updater.go       # Update system for game loop
│   │   └── physics_system.go # Physics integration
│   │
│   ├── matrix/              # Custom math library (CRITICAL - see below)
│   │   ├── vec2.go, vec3.go, vec4.go  # Vector types
│   │   ├── mat3.go, mat4.go           # Matrix types
│   │   ├── quaternion.go              # Quaternion for rotations
│   │   ├── transform.go               # 3D transformations
│   │   ├── float.go, float32.go, float64.go # Float type definitions
│   │   └── matrix.simd.go, matrix.none.go   # SIMD optimizations
│   │
│   ├── klib/                # Utility library (NOT math - see below)
│   │   ├── slice.go         # Slice utilities
│   │   ├── strings.go       # String utilities
│   │   ├── memory.go        # Memory management
│   │   └── ...
│   │
│   ├── rendering/           # Vulkan rendering system
│   │   ├── gpu_*.go         # GPU abstractions (gpu_device.go, gpu_application.go, etc.)
│   │   ├── drawing.go       # Drawing queue and management
│   │   ├── mesh.go, mesh_cache.go     # Mesh handling
│   │   ├── shader.go, shader_cache.go # Shader management
│   │   ├── texture.go, texture_cache.go # Texture management
│   │   ├── material.go, material_cache.go # Material system
│   │   ├── font.go, font_cache.go     # Font handling
│   │   ├── light.go                    # Lighting system
│   │   ├── render_pass.go              # Render passes
│   │   └── loaders/                    # Asset loaders (OBJ, GLTF, etc.)
│   │
│   ├── platform/            # Platform-specific code
│   │   ├── windowing/        # Window creation and management
│   │   ├── audio/           # Audio (Soloud backend)
│   │   ├── hid/             # Input devices (keyboard, mouse, touch, controller, stylus)
│   │   ├── concurrent/      # Threading and work groups
│   │   └── filesystem/      # File system operations
│   │
│   ├── engine/ui/           # UI system (HTML/CSS-like declarative UI)
│   │   ├── ui.go            # Core UI types and base element
│   │   ├── ui_manager.go    # UI Manager (element lifecycle)
│   │   ├── panel.go         # Panel (container element)
│   │   ├── button.go        # Button element
│   │   ├── label.go         # Label (text display)
│   │   ├── input.go         # Input (text field)
│   │   ├── layout.go        # Layout system (positioning, sizing)
│   │   ├── event.go         # Event types
│   │   └── markup/          # HTML/CSS markup system
│   │
│   ├── editor/              # Built-in editor
│   ├── plugins/             # Lua plugin system
│   ├── registry/            # Type registries
│   └── main.test.go         # Example game implementation
│
├── docs/                   # Documentation
└── AGENTS.md              # This file
```

## CRITICAL: Custom Math Library

**DO NOT use external math libraries (e.g., gonum, mathgl).** The Kaiju Engine has a complete custom math library at `kaijuengine.com/matrix`.

### Usage

```go
import "kaijuengine.com/matrix"

// Use matrix.Float for all floating-point operations
// This is typically float32 but can be configured
var pos matrix.Vec3 = matrix.NewVec3(1.0, 2.0, 3.0)
var mat matrix.Mat4 = matrix.Mat4Identity()
```

### Key Types

- **Vector Types**: `Vec2`, `Vec3`, `Vec4` (aliased array types)
- **Matrix Types**: `Mat3`, `Mat4` (16-element arrays for 3D)
- **Quaternion**: For efficient rotation handling
- **Float Type**: `matrix.Float` - configurable precision (default float32)

### Common Functions

```go
// Vector creation
matrix.Vec3{x, y, z} // This is a [3]matrix.Float
matrix.NewVec3(x, y, z)
matrix.Vec3Zero()
matrix.Vec3One()
matrix.Vec3Up(), matrix.Vec3Down(), matrix.Vec3Forward(), etc.

// Matrix creation
matrix.Mat4Identity()
matrix.Mat4Zero()

// Transformations
mat.Translate(position Vec3)
mat.Rotate(rotation Vec3)  // Euler angles
mat.Scale(scale Vec3)

// Vector operations
vec.Add(otherVec3)
vec.Subtract(otherVec3)
vec.Multiply(scalar)
vec.Normal()
vec.Cross(otherVec3)
vec.Dot(otherVec3)
vec.Length()
```

## How to Make a Game

### Step 1: Implement GameInterface

Create a type that implements `bootstrap.GameInterface`:

```go
import (
    "kaijuengine.com/bootstrap"
    "kaijuengine.com/engine"
    "kaijuengine.com/engine/assets"
    "reflect"
)

type Game struct {
    host *engine.Host
    // Add your entities here
}

func (Game) PluginRegistry() []reflect.Type {
    return []reflect.Type{}
}

func (Game) ContentDatabase() (assets.Database, error) {
    // Return a database pointing to your game content
    return assets.NewFileDatabase("game_content")
}
```

### Step 2: Implement Launch

The `Launch` function initializes your game:

```go
func (g *Game) Launch(host *engine.Host) {
    g.host = host
    
    // Create your entities and add drawings
    // See detailed example below
}
```

### Step 3: Register Updates

Use the host's Updater to register game loop functions:

```go
updateId := host.Updater.AddUpdate(g.update)

func (g *Game) update(deltaTime float64) {
    // Game logic here
    // deltaTime is in seconds
}
```

### Step 4: Bootstrap

Return your game from a function called by the engine:

```go
func getGame() bootstrap.GameInterface { return &Game{} }
```

## Complete Example: Creating an Entity with Drawing

This is a distillation of `src/main.test.go`:

```go
package main

import (
    "kaijuengine.com/bootstrap"
    "kaijuengine.com/engine"
    "kaijuengine.com/engine/assets"
    "kaijuengine.com/matrix"
    "kaijuengine.com/registry/shader_data_registry"
    "kaijuengine.com/rendering"
    "math"
    "reflect"
)

type Game struct {
    host *engine.Host
    ball *engine.Entity
}

func (Game) PluginRegistry() []reflect.Type {
    return []reflect.Type{}
}

func (Game) ContentDatabase() (assets.Database, error) {
    // Use file database for game content
    return assets.NewFileDatabase("game_content")
}

func (g *Game) Launch(host *engine.Host) {
    g.host = host
    
    // 1. Create a mesh (sphere: radius, widthSegments, heightSegments)
    sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
    
    // 2. Create shader data for the material
    sd := shader_data_registry.Create("basic")
    sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()
    
    // 3. Create an entity with transform
    // IMPORTANT: Pass host.WorkGroup() for concurrent transform updates
    g.ball = engine.NewEntity(host.WorkGroup())
    
    // 4. Get material and texture from caches
    mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
    if err != nil {
        panic("Material not found - check asset database path")
    }
    tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
    if err != nil {
        panic("Texture not found - check asset database path")
    }
    
    // 5. Create the drawing
    // CRITICAL: Attach entity's Transform to Drawing - this links the drawing to the entity
    draw := rendering.Drawing{
        Material:   mat.CreateInstance([]*rendering.Texture{tex}),
        Mesh:       sphere,
        ShaderData: sd,
        Transform:  &g.ball.Transform,  // <-- KEY: Links drawing to entity transform
        ViewCuller: &host.Cameras.Primary,
    }
    
    // 6. Add drawing to the rendering system
    host.Drawings.AddDrawing(draw)
    
    // 7. Register update function for game loop
    updateId := host.Updater.AddUpdate(g.update)
    
    // 8. Cleanup when entity is destroyed
    g.ball.OnDestroy.Add(func() {
        sd.Destroy()
        host.Updater.RemoveUpdate(&updateId)
    })
}

func (g *Game) update(deltaTime float64) {
    // Animate the ball using sin wave
    x := math.Sin(g.host.Runtime())
    // SetPosition automatically updates the world matrix
    // The drawing automatically follows because it uses &g.ball.Transform
    g.ball.Transform.SetPosition(matrix.NewVec3(matrix.Float(x), 0, -3))
}

func getGame() bootstrap.GameInterface { return &Game{} }
```

### Key Points

1. **Entity Creation**: `engine.NewEntity(host.WorkGroup())` - the WorkGroup enables concurrent transform updates
2. **Transform Attachment**: `Transform: &g.ball.Transform` - this is the critical link that makes the drawing follow the entity
3. **Automatic Updates**: When you call `entity.Transform.SetPosition()`, the world matrix is marked dirty and automatically updated before rendering
4. **Cleanup**: Always clean up ShaderData in `OnDestroy` to prevent memory leaks

## Host System (`src/engine/host.go`)

The `Host` is the central mediator for the entire runtime:

### Key Responsibilities

- **Window Management**: `host.Window`
- **Entity Management**: Create, destroy, update entities
- **Rendering**: `host.Drawings`, `host.Cameras`
- **Caching**: `host.ShaderCache()`, `host.TextureCache()`, `host.MeshCache()`, `host.FontCache()`, `host.MaterialCache()`
- **Update Systems**: `host.Updater`, `host.LateUpdater`, `host.UIUpdater`, `host.UILateUpdater`

### Update Order

The host executes updates in this order each frame:

1. **FrameRunner**: Functions scheduled via `RunAfterFrames`
2. **UIUpdater**: UI update logic
3. **UILateUpdater**: Late UI updates
4. **Update**: Main game logic
5. **LateUpdate**: Physics, collision detection
6. **EndUpdate**: Internal frame preparation

### Key Methods

```go
// Create entities
entity := engine.NewEntity(host.WorkGroup())

// Register updates
updateId := host.Updater.AddUpdate(func(deltaTime float64) { ... })

// Schedule deferred execution
host.RunAfterFrames(10, func() { ... })
host.RunAfterTime(time.Second, func() { ... })

// Access caches
host.MeshCache()
host.TextureCache()
host.ShaderCache()
host.MaterialCache()
host.FontCache()

// Get runtime
runtime := host.Runtime()  // seconds since start
frame := host.Frame()      // current frame number
```

## Entity System (`src/engine/entity.go`)

`Entity` represents a game object with a Transform:

### Key Fields

```go
type Entity struct {
    Transform matrix.Transform  // Position, rotation, scale
    Parent    *Entity          // Parent in hierarchy
    Children  []*Entity        // Child entities
    OnDestroy events.Event     // Fired when entity is destroyed
    OnActivate events.Event    // Fired when entity activates
    OnDeactivate events.Event  // Fired when entity deactivates
}
```

### Key Methods

```go
// Creation
entity := engine.NewEntity(host.WorkGroup())

// Hierarchy
entity.SetParent(otherEntity)
entity.FindByName("name")

// State
entity.Activate()
entity.Deactivate()
entity.IsActive()

// Destruction
host.DestroyEntity(entity)  // Schedule for destruction at next frame

// Named data (arbitrary key-value storage)
entity.AddNamedData("key", someValue)
entity.NamedData("key")

// Drawing integration
entity.StoreShaderData(sd)  // Store render data on entity
entity.ShaderData()          // Retrieve render data
```

## Transform System (`src/matrix/transform.go`)

The `Transform` handles 3D transformations with hierarchical support:

### Key Fields

```go
type Transform struct {
    localMatrix  Mat4  // Local transformation
    worldMatrix Mat4  // World transformation (includes parent transforms)
    parent      *Transform
    children    []*Transform
    position    Vec3
    rotation    Vec3  // Euler angles
    scale       Vec3
}
```

### Key Methods

```go
// Setters
transform.SetPosition(pos Vec3)
transform.SetRotation(rot Vec3)  // Euler angles
transform.SetScale(scale Vec3)

// World-space setters (account for parent)
transform.SetWorldPosition(pos Vec3)
transform.SetWorldRotation(rot Vec3)
transform.SetWorldScale(scale Vec3)

// Getters
transform.Position()      // Local position
transform.WorldPosition() // World position (includes parent)
transform.Rotation()
transform.WorldRotation()
transform.Scale()
transform.WorldScale()

// Direction vectors
transform.Right()    // Local X axis
transform.Up()      // Local Y axis
transform.Forward() // Local Z axis

// Matrix access
transform.Matrix()           // Local matrix
transform.WorldMatrix()      // World matrix
transform.InverseWorldMatrix()

// Hierarchy
transform.SetParent(parentTransform)
transform.SetDirty()  // Mark for update (cascades to children)
```

### Dirty Flag System

Transforms use a dirty flag system for efficient updates:

1. When position/rotation/scale changes, `SetDirty()` is called
2. Dirty transforms are added to a WorkGroup for parallel processing
3. Matrices are updated once before rendering
4. Children are automatically marked dirty when parent changes

## Rendering System (`src/rendering/`)

The rendering system is Vulkan-based with a comprehensive caching layer.

### Architecture

```
GPUApplication
    └── GPUInstance
        └── GPUDevice
            ├── GPULogicalDevice
            │   └── SwapChain
            └── GPUPainter
                └── CommandBuffers
```

### Key Components

#### Caches
- **ShaderCache**: Compiles and caches GLSL shaders
- **TextureCache**: Loads and caches textures
- **MeshCache**: Loads and caches meshes
- **MaterialCache**: Manages materials
- **FontCache**: Handles font rendering

#### Drawing
- **Drawing**: Represents a single renderable object (Material + Mesh + Transform + ShaderData)
- **Drawings**: Queue of all drawings to render
- **RenderPass**: Single pass through the rendering pipeline

#### GPU Abstraction
- `gpu_application.go`: Vulkan application instance
- `gpu_device.go`: Logical device management
- `gpu_swap_chain.go`: Framebuffer swapping
- `gpu_physical_device.go`: GPU hardware detection
- Platform-specific implementations in `gpu_*_vulkan.go` files

### Creating Drawings

```go
// 1. Get mesh from cache
mesh := host.MeshCache().Mesh("path/to/mesh.obj") // Or a UUID key

// 2. Get material
mat, _ := host.MaterialCache().Material(assets.MaterialDefinitionBasic)

// 3. Create material instance
matInstance := mat.CreateInstance([]*rendering.Texture{texture})

// 4. Create shader data
sd := shader_data_registry.Create("basic") // Use the one that matches the shader

// 5. Create and add drawing
draw := rendering.Drawing{
    Material:   matInstance,
    Mesh:       mesh,
    ShaderData: sd,
    Transform:  &entity.Transform,
    ViewCuller: &host.Cameras.Primary,
}
host.Drawings.AddDrawing(draw)
```

### Pre-built Meshes

The engine provides utility functions for common shapes (in src/rendering/mesh.go):

```go
rendering.NewMeshSphere(cache, radius, widthSegments, heightSegments)
rendering.NewMeshCube(cache, size)
rendering.NewMeshPlane(cache, width, depth)
```

## UI System (`src/engine/ui/`)

The Kaiju Engine features an optional web-inspired UI system with HTML/CSS-like layout, full event handling, and a markup system for declarative UI creation. The UI renders on a separate orthographic camera (`host.Cameras.UI`) from the main 3D rendering. You can choose construct the UI without using any HTML/CSS markup, developers can create stylizers directly.

### Architecture

```
UI Manager (ui.Manager)
    └── UI Elements (Panel as base)
            ├── Button
            ├── Label
            ├── Input
            ├── Image
            ├── Checkbox
            ├── Slider
            ├── Select
            └── ProgressBar

Markup System (ui/markup/)
    ├── document/          # DOM-like document model
    ├── css/              # CSS parsing and styling
    └── spec_generator/  # Code generation
```

### UI Manager

The UI Manager is the container for all UI elements. It must be created and initialized before use:

```go
import "kaijuengine.com/engine/ui"

// Create manager
man := ui.Manager{}

// Initialize with host (REQUIRED before creating elements)
man.Init(host)

// Create UI elements
button := man.Add().ToButton()
```

### Key Methods

```go
// Add creates a new UI element (returns *UI)
man.Add()

// Remove destroys a UI element
man.Remove(uiElement)

// Clear removes all UI elements
man.Clear()

// Reserve pre-allocates element slots
man.Reserve(100)

// Hovered returns all panels currently being hovered
man.Hovered()
```

### Creating UI Elements

All UI elements are created through the manager using type assertions:

```go
// Create a panel
panel := man.Add().ToPanel()
panel.Init(nil, ElementTypePanel)

// Create a button (Panel with click behavior)
button := man.Add().ToButton()
button.Init(texture, "Click Me")

// Create a label
label := man.Add().ToLabel()
label.Init("Hello World")

// Create an input field
input := man.Add().ToInput()
input.Init("placeholder")

// Create an image
image := man.Add().ToImage()
image.Init(texture)

// Create a slider
slider := man.Add().ToSlider()
slider.Init()

// Create a checkbox
checkbox := man.Add().ToCheckbox()
checkbox.Init()

// Create a progress bar
progress := man.Add().ToProgressBar()
progress.Init()
```

### Element Types

| Type | Description | Base |
|------|-------------|------|
| `Panel` | Container element, base for most elements | UI |
| `Button` | Clickable button with hover states | Panel |
| `Label` | Text display | UI |
| `Input` | Text input field | Panel |
| `Image` | Texture display | UI |
| `Checkbox` | Toggle checkbox | Panel |
| `Slider` | Value slider | Panel |
| `Select` | Dropdown selector | Panel |
| `ProgressBar` | Progress indicator | Panel |

### Event Handling

UI elements support a comprehensive event system:

```go
// Event types
ui.EventTypeEnter       // Mouse enters element
ui.EventTypeExit        // Mouse exits element
ui.EventTypeClick       // Left click
ui.EventTypeRightClick  // Right click
ui.EventTypeDoubleClick // Double click
ui.EventTypeDown        // Mouse button down
ui.EventTypeUp          // Mouse button up
ui.EventTypeRightDown   // Right mouse button down
ui.EventTypeRightUp     // Right mouse button up
ui.EventTypeMiss        // Click missed element
ui.EventTypeDragStart   // Drag started
ui.EventTypeDragEnd     // Drag ended
ui.EventTypeDrop        // Drop occurred
ui.EventTypeScroll      // Mouse scroll
ui.EventTypeChange      // Value changed
ui.EventTypeFocus       // Element focused
ui.EventTypeBlur        // Element lost focus
ui.EventTypeSubmit      // Form submitted (Input)
ui.EventTypeKeyDown     // Key pressed
ui.EventTypeKeyUp       // Key released
```

#### Adding Event Handlers

```go
// Add click handler
button.Base().AddEvent(ui.EventTypeClick, func() {
    fmt.Println("Button clicked!")
})

// Add hover enter handler
button.Base().AddEvent(ui.EventTypeEnter, func() {
    fmt.Println("Mouse entered button")
})

// Add change handler (for inputs, sliders, etc.)
input.Base().AddEvent(ui.EventTypeChange, func() {
    fmt.Println("Input changed:", input.Value())
})

// Remove event handler
button.Base().RemoveEvent(ui.EventTypeClick, eventId)
```

### Layout System

The layout system controls element positioning and sizing:

```go
// Positioning modes
layout.SetPositioning(PositioningStatic)    // Normal flow
layout.SetPositioning(PositioningAbsolute)  // Fixed position
layout.SetPositioning(PositioningFixed)     // Relative to viewport
layout.SetPositioning(PositioningRelative)  // Relative to normal position
layout.SetPositioning(PositioningSticky)   // Sticky positioning

// Size (in pixels)
layout.Scale(width, height)      // Set width and height
layout.ScaleWidth(width)         // Set width only
layout.ScaleHeight(height)      // Set height only

// Margins (left, top, right, bottom)
layout.SetMargin(left, top, right, bottom)

// Padding (left, top, right, bottom)
layout.SetPadding(left, top, right, bottom)

// Border (left, top, right, bottom)
layout.SetBorder(left, top, right, bottom)

// Offset (for absolute positioning)
layout.SetOffset(x, y)

// Z-order
layout.SetZ(z float32)

// Auto-fit content to children
panel.FitContent()        // Fit width and height
panel.FitContentWidth()  // Fit width only
panel.FitContentHeight() // Fit height only
panel.DontFitContent()   // Disable auto-fit
```

### Panel Operations

```go
// Add child element
panel.AddChild(childUI)

// Remove child element
panel.RemoveChild(childUI)

// Background color
panel.SetColor(matrix.ColorWhite())

// Background texture
panel.SetBackground(texture)

// Border radius
panel.SetBorderRadius(topLeft, topRight, bottomRight, bottomLeft)

// Scrolling
panel.SetScrollDirection(PanelScrollDirectionVertical)  // Vertical
panel.SetScrollDirection(PanelScrollDirectionHorizontal) // Horizontal
panel.SetScrollDirection(PanelScrollDirectionBoth)     // Both
panel.ScrollX()      // Current X scroll position
panel.ScrollY()      // Current Y scroll position
panel.SetScrollY(50) // Set scroll position
panel.ResetScroll()  // Reset to top
```

### Label Operations

```go
// Text
label.SetText("New Text")
label.Text() // Get current text

// Font styling
label.SetFontSize(16.0)
label.SetFontFace(rendering.FontRegular)
label.SetFontWeight("bold")   // "normal", "bold", "bolder", "lighter"
label.SetFontStyle("italic")  // "normal", "italic"

// Text alignment
label.SetJustify(rendering.FontJustifyLeft)   // or Center, Right
label.SetBaseline(rendering.FontBaselineTop)  // or Center, Bottom

// Colors
label.SetColor(matrix.ColorWhite())     // Text color
label.SetBGColor(matrix.ColorBlack())   // Background color

// Word wrap
label.SetWrap(true)  // Enable word wrap (default)
label.SetWrap(false) // Disable word wrap

// Auto-width based on text
label.SetWidthAutoHeight(width)
```

### Input Operations

```go
// Get/set value
input.Value()        // Get current text
input.SetValue("text") // Set text

// Placeholder
input.Placeholder()        // Get placeholder text
input.SetPlaceholder("...") // Set placeholder

// Events
input.Base().AddEvent(ui.EventTypeSubmit, func() {
    // Handle enter key
})
input.Base().AddEvent(ui.EventTypeChange, func() {
    // Handle value change
})
```

### Dirty Flag System

UI elements use a dirty flag system to optimize updates:

```go
// Mark element as needing update
ui.SetDirty(DirtyTypeLayout)     // Layout changed
ui.SetDirty(DirtyTypeResize)     // Size changed
ui.SetDirty(DirtyTypeGenerated)  // Content regenerated
ui.SetDirty(DirtyTypeColorChange) // Color changed
```

### HTML/CSS Markup System

The UI system supports declarative UI creation via HTML and CSS:

```go
import (
    "kaijuengine.com/engine/ui/markup"
    "kaijuengine.com/engine/ui/markup/document"
)

// From asset file
doc, err := markup.DocumentFromHTMLAsset(&man, "ui/main.html", nil, nil)

// From string
doc := markup.DocumentFromHTMLString(&man, htmlStr, cssStr, nil, nil, nil)

// Access elements
element, _ := doc.GetElementById("myButton")
element.UI.ToButton().Base().AddEvent(ui.EventTypeClick, handler)

// Get root panel
rootPanel := doc.Elements[0].UI.ToPanel()
```

### Complete Example: Creating a Simple UI

```go
import (
    "kaijuengine.com/engine"
    "kaijuengine.com/engine/ui"
    "kaijuengine.com/matrix"
    "kaijuengine.com/rendering"
    "kaijuengine.com/registry/shader_data_registry"
)

type Game struct {
    host       *engine.Host
    uiMan      ui.Manager
    myButton   *ui.Button
    myLabel    *ui.Label
    myPanel    *ui.Panel
}

func (g *Game) Launch(host *engine.Host) {
    g.host = host
    
    // Initialize UI Manager
    g.uiMan.Init(host)
    
    // Create a root panel (fills window)
    g.myPanel = g.uiMan.Add().ToPanel()
    g.myPanel.Init(nil, ui.ElementTypePanel)
    g.myPanel.SetColor(matrix.ColorGray().ScaleAlpha(0.8))
    g.myPanel.SetMargin(10, 10, 10, 10)
    g.myPanel.SetPadding(10, 10, 10, 10)
    
    // Create a label
    g.myLabel = g.uiMan.Add().ToLabel()
    g.myLabel.Init("Hello, Kaiju!")
    g.myLabel.SetFontSize(24.0)
    g.myLabel.SetColor(matrix.ColorWhite())
    g.myPanel.AddChild(g.myLabel.Base())
    
    // Create a button
    g.myButton = g.uiMan.Add().ToButton()
    // Get texture from cache
    tex, _ := host.TextureCache().Texture("square", rendering.TextureFilterLinear)
    g.myButton.Init(tex, "Click Me")
    g.myButton.Base().Layout().SetMargin(0, 10, 0, 0)
    g.myPanel.AddChild(g.myButton.Base())
    
    // Add click event
    g.myButton.Base().AddEvent(ui.EventTypeClick, func() {
        g.myLabel.SetText("Button Clicked!")
    })
    
    // Add hover events
    g.myButton.Base().AddEvent(ui.EventTypeEnter, func() {
        g.myButton.SetColor(matrix.ColorGreen())
    })
    g.myButton.Base().AddEvent(ui.EventTypeExit, func() {
        g.myButton.SetColor(matrix.ColorWhite())
    })
}
```

### Update Cycle

UI updates run through the host's UI updater:

- **UIUpdater**: Runs before main game Update
- **UILateUpdater**: Runs after UIUpdater

The UI manager automatically registers with these updaters when `Init()` is called.

### Accessing UI Camera

The UI renders to a separate orthographic camera:

```go
// UI camera is available via host
uiCamera := host.Cameras.UI
```

## Building the Project

### Prerequisites
- Go 1.25.0+
- C build tools
- Vulkan SDK

### Build Command

```bash
cd src
go build -tags="debug,editor" -o ../ ./
```

### Build Tags

- `debug`: Include debug information
- `editor`: Build with editor support
- Platform-specific: `.windows.go`, `.darwin.go`, `.linux.go`, `.android.go`

### Content

Game content should be placed in a `game_content/` directory at runtime. The example in `main.test.go` shows how to use `assets.NewFileDatabase("game_content")`. When building a game from the editor, content is placed in the `database/content` directory by UUID (or custom name).

## Common Patterns

### Update Registration

```go
// Register update
id := host.Updater.AddUpdate(func(dt float64) {
    // Game logic
})

// Remove update (e.g., when entity destroyed)
host.Updater.RemoveUpdate(&id)
```

### Frame-Safe Operations

```go
// Run on next frame
host.RunNextFrame(func() { ... })

// Run after N frames
host.RunAfterFrames(60, func() { ... })

// Run after time duration
host.RunAfterTime(time.Second, func() { ... })
```

## Important Notes for Agents

1. **Always use `kaijuengine.com/matrix`** for math - never import external math libraries
2. **Use `matrix.Float`** instead of `float32` or `float64` for engine-compatible code
3. **Attach entity transforms to drawings** using `&entity.Transform` - this is the critical link
4. **Clean up resources** in `OnDestroy` handlers to prevent memory leaks
5. **Use the WorkGroup** when creating entities: `engine.NewEntity(host.WorkGroup())`
6. **Content paths** are relative to the working directory at runtime
7. **Update functions** receive `deltaTime` in seconds
