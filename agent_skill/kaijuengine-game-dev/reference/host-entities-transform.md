# Host, Entities, and Transform

## Host (`src/engine/host.go`)

The `Host` is the central runtime mediator.

### Responsibilities

- **Window**: `host.Window`
- **Entities**: create / destroy / update
- **Rendering**: `host.Drawings`, `host.Cameras` (`Cameras.Primary`, `Cameras.UI`)
- **Caches**: `host.ShaderCache()`, `host.TextureCache()`, `host.MeshCache()`,
  `host.FontCache()`, `host.MaterialCache()`
- **Updaters**: `host.Updater`, `host.LateUpdater`, `host.UIUpdater`,
  `host.UILateUpdater`

### Frame / update order

1. **FrameRunner** — `RunAfterFrames` / `RunNextFrame` callbacks (after the window
   poll that ingests OS input, before the UI updates)
2. **UIUpdater**
3. **UILateUpdater**
4. **Update** — main game logic (`host.Updater`)
5. **LateUpdate** — physics, collision (`host.LateUpdater`)
6. **EndUpdate** — internal frame prep (input state transitions)

### Key methods

```go
entity := engine.NewEntity(host.WorkGroup())

id := host.Updater.AddUpdate(func(deltaTime float64) { ... })
host.Updater.RemoveUpdate(&id) // takes a *pointer* to the id

host.RunNextFrame(func() { ... })
host.RunAfterFrames(10, func() { ... })
host.RunAfterTime(time.Second, func() { ... })

host.MeshCache(); host.TextureCache(); host.ShaderCache()
host.MaterialCache(); host.FontCache()

host.Runtime() // seconds since start
host.Frame()   // current frame number
```

## Entity (`src/engine/entity.go`)

A game object with a Transform.

### Key fields

```go
type Entity struct {
    Transform    matrix.Transform // position, rotation, scale
    Parent       *Entity
    Children     []*Entity
    OnDestroy    events.Event
    OnActivate   events.Event
    OnDeactivate events.Event
}
```

### Key methods

```go
entity := engine.NewEntity(host.WorkGroup())

// Hierarchy
entity.SetParent(other)
entity.FindByName("name")

// State
entity.Activate(); entity.Deactivate(); entity.IsActive()

// Destruction (cleaned up at the next frame)
host.DestroyEntity(entity)

// Named data (arbitrary key-value)
entity.AddNamedData("key", value)
entity.NamedData("key")

// Drawing integration
entity.StoreShaderData(sd)
entity.ShaderData()
```

## Transform (`src/matrix/transform.go`)

Hierarchical 3D transform.

### Key fields (conceptual)

```go
type Transform struct {
    localMatrix Mat4
    worldMatrix Mat4 // includes parent transforms
    parent      *Transform
    children    []*Transform
    position    Vec3
    rotation    Vec3 // Euler angles
    scale       Vec3
}
```

### Key methods

```go
// Local setters
transform.SetPosition(pos Vec3)
transform.SetRotation(rot Vec3) // Euler angles
transform.SetScale(scale Vec3)

// World-space setters (account for parent)
transform.SetWorldPosition(pos Vec3)
transform.SetWorldRotation(rot Vec3)
transform.SetWorldScale(scale Vec3)

// Getters
transform.Position(); transform.WorldPosition()
transform.Rotation(); transform.WorldRotation()
transform.Scale();    transform.WorldScale()

// Direction vectors
transform.Right()   // local X
transform.Up()      // local Y
transform.Forward() // local Z

// Matrices
transform.Matrix(); transform.WorldMatrix(); transform.InverseWorldMatrix()

// Hierarchy
transform.SetParent(parent)
transform.SetDirty() // cascades to children
```

### Dirty-flag system

1. Changing position/rotation/scale calls `SetDirty()`.
2. Dirty transforms are added to a WorkGroup for parallel processing.
3. Matrices are recomputed once before rendering.
4. Children are marked dirty automatically when a parent changes.
