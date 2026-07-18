# Custom Math Library (`kaijuengine.com/matrix`)

**DO NOT use external math libraries (gonum, mathgl, etc.).** The engine has a
complete custom math library. Use `matrix.Float` for all floating-point in
engine-facing code (configurable precision, default float32).

```go
import "kaijuengine.com/matrix"

var pos matrix.Vec3 = matrix.NewVec3(1.0, 2.0, 3.0)
var mat matrix.Mat4 = matrix.Mat4Identity()
```

## Key types

- **Vectors**: `Vec2`, `Vec3`, `Vec4` (aliased array types, e.g. `Vec3` is
  `[3]matrix.Float`)
- **Matrices**: `Mat3`, `Mat4` (16-element arrays for 3D)
- **Quaternion**: efficient rotation handling
- **Float**: `matrix.Float` — configurable precision (default float32)

## Common functions

```go
// Vector creation
matrix.Vec3{x, y, z}        // a [3]matrix.Float literal
matrix.NewVec3(x, y, z)
matrix.Vec3Zero()
matrix.Vec3One()
matrix.Vec3Up(); matrix.Vec3Down(); matrix.Vec3Forward() // etc.

// Matrix creation
matrix.Mat4Identity()
matrix.Mat4Zero()

// Transformations
mat.Translate(position Vec3)
mat.Rotate(rotation Vec3) // Euler angles
mat.Scale(scale Vec3)

// Vector operations
vec.Add(other)
vec.Subtract(other)
vec.Multiply(scalar)
vec.Normal()
vec.Cross(other)
vec.Dot(other)
vec.Length()
```

Colors are also in this package (`matrix.ColorRed()`, `matrix.ColorWhite()`,
`color.ScaleAlpha(a)`, etc.).
