---
title: Working with VFX in Kaiju Engine
description: An overview of the visual effects (VFX) subsystem, how particle systems and emitters are implemented, and how to use them in the Kaiju Engine editor.
tags: vfx, particles, engine, graphics, go
image: images/vfx_overview.png
date: 2026-01-30
---

# Working with VFX in Kaiju Engine


---

## Introduction

Visual effects (VFX) are a cornerstone of modern game development, providing everything from subtle smoke trails to spectacular fireworks.  Kaiju Engine's VFX subsystem is designed to be lightweight, extensible, and tightly integrated with the editor, allowing designers to craft and iterate on particle effects in real time.

In this post we'll explore the architecture of the VFX system, dive into the key data structures (`Particle`, `Emitter`, and `ParticleSystem`), and show how to edit emitters using the built‑in editor UI.  We'll also provide a simple firework example, debugging tips, and guidance on extending the system with custom path functions.

---

## VFX Architecture Overview

At a high level the VFX pipeline consists of three main pieces:

1. **Particle** - a lightweight struct that stores transform, velocity, opacity and lifespan.
2. **Emitter** - owns a list of particles, spawns them according to a configuration, and updates them each frame.
3. **ParticleSystem** - aggregates one or more emitters and provides a single interface for the renderer.

Relevant source files include:

- `src/rendering/vfx/particle.go` - defines the `Particle` type and its update logic.
- `src/rendering/vfx/emitter.go` - implements spawning, path functions, and per‑particle data.
- `src/rendering/vfx/emitter_path_funcs.go` - registers built‑in path functions (e.g., `Circle`).
- `src/editor/editor_workspace/vfx_workspace/vfx_workspace.go` - UI glue that lets you edit emitters in the editor.

---

## Particle Structure

```go
type particleTransformation struct {
    Position matrix.Vec3
    Rotation matrix.Vec3 // TODO:  This can be 1D for billboarded particle
    Scale    matrix.Vec3 // TODO:  This can be 2D for billboarded particle
}

type Particle struct {
    Transform       particleTransformation
    Velocity        particleTransformation
    OpacityVelocity float32
    LifeSpan        float32
}
```

The `update` method advances the particle based on its velocity and reduces its remaining lifespan:

```go
func (p *Particle) update(deltaTime float64) {
    p.LifeSpan -= float32(deltaTime)
    t := &p.Transform
    v := &p.Velocity
    t.Position.AddAssign(v.Position.Scale(matrix.Float(deltaTime)))
    t.Rotation.AddAssign(v.Rotation.Scale(matrix.Float(deltaTime)))
    t.Scale.AddAssign(v.Scale.Scale(matrix.Float(deltaTime)))
}
```

---

## Emitters

An `Emitter` holds a slice of `Particle` objects and a configuration struct (`EmitterConfig`).  The config controls texture, spawn rate, particle lifespan, direction ranges, velocity ranges, color, and optional path functions.

Key fields in `EmitterConfig`:

```go
type EmitterConfig struct {
    Texture          content_id.Texture
    SpawnRate        float64
    ParticleLifeSpan float32
    LifeSpan         float64
    Offset           matrix.Vec3
    DirectionMin     matrix.Vec3
    DirectionMax     matrix.Vec3
    VelocityMinMax   matrix.Vec2
    OpacityMinMax    matrix.Vec2
    Color            matrix.Color
    PathFuncName     string `options:"PathFuncName"`
    PathFunc         func(t float64) matrix.Vec3 `visible:"hidden"`
    PathFuncOffset   float64
    PathFuncScale    float32
    PathFuncSpeed    float32
    FadeOutOverLife  bool
    Burst            bool
    Repeat           bool
}
```

The emitter spawns particles based on `SpawnRate` and applies the optional path function to offset the whole system over time.

---

## Path Functions

Path functions let you move an entire emitter along a curve.  The engine ships with a `Circle` function, but you can register your own.

```go
func init() {
    RegisterPathFunc("None", nil)
    RegisterPathFunc("Circle", pathFuncCircle)
}

func pathFuncCircle(t float64) matrix.Vec3 {
    // Normalise t to the range [0,1]
    for t < 0 { t += 1 }
    for t > 1 { t -= 1 }
    angle := matrix.Float(2 * math.Pi * t)
    var pos matrix.Vec3
    pos.SetX(matrix.Cos(angle))
    pos.SetZ(matrix.Sin(angle))
    return pos
}
```

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.mp4" type="video/mp4">
	<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.apng" />
</video>

---

## Editing VFX in the Editor

The VFX editor UI is defined in `editor/ui/workspace/vfx_workspace.go.html`.  It provides two panels:

- **Left panel** - list of emitters, system name, and add/save buttons.
- **Right panel** - per‑emitter data bindings (texture, color, direction, etc.).

![VFX Workspace UI](https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_01_30/vfx_workspace_ui.png)

### Adding a New Emitter

Click the **Add Emitter** button to create a new emitter with a default `EmitterConfig`:

```go
w.addEmitter(vfx.EmitterConfig{
    Texture:          "smoke.png",
    SpawnRate:        0.05,
    ParticleLifeSpan: 2,
    Color:            matrix.ColorWhite(),
    DirectionMin:     matrix.NewVec3(-0.3, 1, -0.3),
    DirectionMax:     matrix.NewVec3(0.3, 1, 0.3),
    VelocityMinMax:   matrix.Vec2One().Scale(1),
    OpacityMinMax:    matrix.NewVec2(0.3, 1.0),
    FadeOutOverLife:  true,
    PathFuncScale:    1,
    PathFuncSpeed:    1,
})
```

You can then edit each field in the right‑hand panel.  When you're satisfied, click **Save** - the workspace serialises the `ParticleSystemSpec` to JSON and writes it back to the project's content database.

---

## Still a work in progress

### Known limitations

* **CPU‑only simulation** - At the moment particles are updated on the CPU.  This works well for modest counts, but large fire‑works or dense smoke quickly become a bottleneck.
* **Path functions are static** - The built‑in `Circle` is the only non‑trivial path function.  Custom functions can be registered, but there is no UI for authoring functions in-editor.
* **Limited editor feedback** - The VFX workspace shows the raw config values, but does not visualise the spawn area, direction cones, or velocity ranges directly in the viewport.

### Planned improvements

* **GPU particle pipelines** - Off‑load the `update` and `spawn` logic to a compute shader.
* **Rich path‑function editor** - Expose some curve editors in the UI to set path curves.
* **Live preview helpers** - Visual gizmos for spawn cones, velocity vectors, and opacity envelopes to make tweaking feel immediate.

### Contributing

The VFX subsystem is deliberately lightweight, but I welcome extensions. To add a new path function:

1. Implement the function in `src/rendering/vfx/emitter_path_funcs.go` following the `pathFuncCircle` example.
2. Register it with `RegisterPathFunc("MyPath", myPathFunc)` inside the same file.
3. Submit a pull request with tests that verify the function's output range.

If you encounter bugs or have ideas for new emitter features, open an issue on the repository or join the discussion in Discord.
