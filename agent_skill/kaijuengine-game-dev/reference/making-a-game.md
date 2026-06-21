# Making a Game

A game implements `bootstrap.GameInterface`. The engine bootstraps a `*engine.Host`
and calls your `Launch(host)`.

## Step 1 — Implement GameInterface

```go
import (
    "reflect"
    "kaijuengine.com/bootstrap"
    "kaijuengine.com/engine"
    "kaijuengine.com/engine/assets"
)

type Game struct {
    host *engine.Host
    // your entities/state here
}

func (Game) PluginRegistry() []reflect.Type { return []reflect.Type{} }

func (Game) ContentDatabase() (assets.Database, error) {
    // Database pointing at your game content (relative to the working directory).
    return assets.NewFileDatabase("game_content")
}
```

## Step 2 — Implement Launch

```go
func (g *Game) Launch(host *engine.Host) {
    g.host = host
    // create entities, add drawings, register updates (see full example below)
}
```

## Step 3 — Register updates

```go
updateId := host.Updater.AddUpdate(g.update)

func (g *Game) update(deltaTime float64) {
    // game logic; deltaTime is in seconds
}
```

## Step 4 — Bootstrap

```go
func getGame() bootstrap.GameInterface { return &Game{} }
```

## Complete example: entity with a drawing

A distillation of `src/main.test.go`. The key idea: create a mesh, material
instance, and shader data, create an entity, then bind them in a `Drawing` whose
`Transform` points at the entity's transform.

```go
package main

import (
    "math"
    "reflect"

    "kaijuengine.com/bootstrap"
    "kaijuengine.com/engine"
    "kaijuengine.com/engine/assets"
    "kaijuengine.com/matrix"
    "kaijuengine.com/registry/shader_data_registry"
    "kaijuengine.com/rendering"
)

type Game struct {
    host *engine.Host
    ball *engine.Entity
}

func (Game) PluginRegistry() []reflect.Type { return []reflect.Type{} }

func (Game) ContentDatabase() (assets.Database, error) {
    return assets.NewFileDatabase("game_content")
}

func (g *Game) Launch(host *engine.Host) {
    g.host = host

    // 1. Mesh (sphere: radius, widthSegments, heightSegments)
    sphere := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)

    // 2. Shader data for the material
    sd := shader_data_registry.Create("basic")
    sd.(*shader_data_registry.ShaderDataStandard).Color = matrix.ColorRed()

    // 3. Entity with transform. Pass host.WorkGroup() for concurrent updates.
    g.ball = engine.NewEntity(host.WorkGroup())

    // 4. Material + texture from caches
    mat, err := host.MaterialCache().Material(assets.MaterialDefinitionBasic)
    if err != nil {
        panic("Material not found - check asset database path")
    }
    tex, err := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
    if err != nil {
        panic("Texture not found - check asset database path")
    }

    // 5. Drawing. CRITICAL: Transform must point at the entity's transform.
    draw := rendering.Drawing{
        Material:   mat.CreateInstance([]*rendering.Texture{tex}),
        Mesh:       sphere,
        ShaderData: sd,
        Transform:  &g.ball.Transform, // <-- links the drawing to the entity
        ViewCuller: &host.Cameras.Primary,
    }

    // 6. Enqueue the drawing
    host.Drawings.AddDrawing(draw)

    // 7. Per-frame update
    updateId := host.Updater.AddUpdate(g.update)

    // 8. Cleanup on destroy (avoid leaks)
    g.ball.OnDestroy.Add(func() {
        sd.Destroy()
        host.Updater.RemoveUpdate(&updateId)
    })
}

func (g *Game) update(deltaTime float64) {
    x := math.Sin(g.host.Runtime())
    // SetPosition marks the world matrix dirty; the drawing follows because it
    // references &g.ball.Transform.
    g.ball.Transform.SetPosition(matrix.NewVec3(matrix.Float(x), 0, -3))
}

func getGame() bootstrap.GameInterface { return &Game{} }
```

### Key points

1. **Entity creation** — `engine.NewEntity(host.WorkGroup())`; the work group
   enables concurrent transform updates.
2. **Transform attachment** — `Transform: &entity.Transform` is the link that
   makes a drawing track its entity.
3. **Automatic updates** — calling `entity.Transform.SetPosition()` marks the
   world matrix dirty; it is recomputed before rendering.
4. **Cleanup** — always destroy `ShaderData` and remove updaters in `OnDestroy`.

## Common patterns

### Update registration

```go
id := host.Updater.AddUpdate(func(dt float64) { /* game logic */ })
host.Updater.RemoveUpdate(&id) // e.g. when the entity is destroyed
```

### Frame-safe / deferred operations

```go
host.RunNextFrame(func() { ... })            // next frame
host.RunAfterFrames(60, func() { ... })       // after N frames
host.RunAfterTime(time.Second, func() { ... }) // after a duration
```
