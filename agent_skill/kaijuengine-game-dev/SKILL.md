---
name: kaijuengine-game-dev
description: >-
  Essential reference for building games and tools on the Kaiju Engine (Go module
  `kaijuengine.com`, Vulkan-backed). Covers the GameInterface bootstrap, the Host
  runtime and frame/update order, entities and the custom matrix Transform, the
  custom `kaijuengine.com/matrix` math library (never use gonum/mathgl), the
  Vulkan rendering/Drawing system, the HTML/CSS-like UI system (no JS runtime),
  build tags, and integration testing. Use this whenever working in a kaijuengine
  project — writing or modifying Go that imports `kaijuengine.com/...`, scaffolding
  a new game, editing engine or editor source, or answering how-to/architecture
  questions about entities, rendering, materials/shaders, or the UI — even when the
  user doesn't explicitly say "Kaiju". Prefer this over generic guidance for
  anything touching the kaiju engine tree.
---

# Kaiju Engine — Game Development

Kaiju Engine is a 2D/3D game engine in Go (`kaijuengine.com`, Go 1.25.0+) backed
by Vulkan, with a custom math library, a hierarchical entity system, and a
web-inspired UI. Cross-platform: Windows, Linux, macOS, Android.

A game is a type implementing `bootstrap.GameInterface` (`ContentDatabase`,
`PluginRegistry`, `Launch`). The engine creates a `*engine.Host` — the central
runtime mediator — and calls `Launch(host)`. Your code creates entities, attaches
drawings, and registers per-frame update functions on the host.

> **Source of truth:** this skill is distilled from the engine's `AGENTS.md`. When
> it disagrees with the current source, trust the source — and verify a symbol
> exists before relying on it, since some details here describe internals
> conceptually and may drift from the live API.

## Critical rules (do not violate)

1. **Use `kaijuengine.com/matrix` for ALL math.** Never import gonum, mathgl, or
   any external math library. Use `matrix.Float` (configurable, default float32)
   instead of bare `float32`/`float64` in engine-facing code.
2. **Attach the entity transform to the drawing**: `Transform: &entity.Transform`.
   This is the link that makes a drawing follow its entity. Without it the drawing
   won't track the entity.
3. **Create entities with the work group**: `engine.NewEntity(host.WorkGroup())` —
   enables concurrent transform updates.
4. **Clean up in `OnDestroy`**: destroy `ShaderData` and remove updaters
   (`host.Updater.RemoveUpdate(&id)`) to avoid leaks. `RemoveUpdate` takes a
   pointer to the id.
5. **UI markup has NO JavaScript.** `.go.html` files are Go templates; an
   `onclick="fn"` maps to a SINGLE Go function in the funcMap. Chained JS-style
   calls (`a(); b()`) are invalid. CSS is custom-parsed; manage active states with
   CSS classes via `document.SetElementClasses(elm, ...)`.
6. **Content paths are relative to the working directory at runtime.** Games load
   content via `assets.NewFileDatabase("game_content")`; the editor places content
   in `database/content` by UUID.
7. **`deltaTime` is in seconds.** `host.Runtime()` is seconds since start;
   `host.Frame()` is the current frame number.

## Engine Go conventions

Follow the repo's Go style guide (also in the user's global instructions):
builder-is-implementation pattern, public API surface is interfaces/methods only
(no exported structs), locks are a last resort (prefer channels or
`github.com/puzpuzpuz/xsync/v4`), and platform build-tag files
(`_darwin.go`/`_linux.go`/`_windows.go`) hold ONLY what differs — shared structs
and constructors live in a single untagged file.

## Quick start — minimal game

```go
package main

import (
    "reflect"
    "kaijuengine.com/bootstrap"
    "kaijuengine.com/engine"
    "kaijuengine.com/engine/assets"
)

type Game struct{ host *engine.Host }

func (Game) PluginRegistry() []reflect.Type        { return nil }
func (Game) ContentDatabase() (assets.Database, error) {
    return assets.NewFileDatabase("game_content")
}
func (g *Game) Launch(host *engine.Host) {
    g.host = host
    // create entities, add drawings, register updates here
    id := host.Updater.AddUpdate(g.update)
    _ = id
}
func (g *Game) update(deltaTime float64) { /* game logic; deltaTime in seconds */ }

func getGame() bootstrap.GameInterface { return &Game{} }
```

See `reference/making-a-game.md` for the full entity+drawing example.

## Frame / update order

Each frame the host runs, in order:

1. **FrameRunner** — functions scheduled via `RunAfterFrames` / `RunNextFrame`
   (run after window poll, before the UI)
2. **UIUpdater** — UI update logic
3. **UILateUpdater** — late UI updates
4. **Update** — main game logic (`host.Updater`)
5. **LateUpdate** — physics, collision (`host.LateUpdater`)
6. **EndUpdate** — internal frame preparation (input state transitions)

## Reference index (read on demand)

- **`reference/making-a-game.md`** — GameInterface, Launch steps, the complete
  entity-with-drawing example, and common patterns (update registration,
  frame-safe ops). Read when scaffolding a game or wiring entities to drawings.
- **`reference/host-entities-transform.md`** — Host responsibilities and methods,
  the Entity type (hierarchy, named data, activation), and the matrix Transform
  (local vs world, direction vectors, dirty-flag system). Read for entity
  hierarchy, positioning, or host caches.
- **`reference/rendering.md`** — Vulkan architecture, the caches, building
  Drawings, pre-built meshes, and terrain texture painting / splat-map contract.
  Read for rendering, meshes, materials, shaders, or terrain.
- **`reference/ui.md`** — the UI system: Manager, element types, events, layout,
  panel/label/input ops, dirty flags, and the HTML/CSS markup system. Read for any
  UI work in a game or the editor.
- **`reference/matrix.md`** — the custom math library: vector/matrix/quaternion
  types and common operations. Read before doing any math.
- **`reference/building-and-testing.md`** — build prerequisites, build tags, build
  commands, content layout, and the integration-testing framework (screenshots +
  video). Read to build, run, or visually test.

## Verifying visually

- For automated visual checks, use the **integration testing** framework
  (screenshots / video) — see `reference/building-and-testing.md`.
- To drive a *running* game interactively (screenshot + inject input), build with
  `-tags ai_driver` and use the **`kaiju-aidriver`** skill.
