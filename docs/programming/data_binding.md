---
title: Data Binding | Kaiju Engine
---

# Binding data for the editor
In the Kaiju engine, it's possible to create structures that the editor is aware
of. These structures are required to be plain-ol-data (POD), but can also be
some specialized engine structures as well. These structures are meant to provide
a method of easily binding arbitrary data for your game code.

## Creating a structure
There are only 2 requirements for your structure, (1) it must be POD, and (2) it
must implement the interface:
```go
func Init(e *engine.Entity, host *engine.Host)
```

## Registering your structure
To register your structure, you'll need to add it to the engine registry. You
typically will do this in the `init` function of your package. The function you
will need to call is `engine.RegisterEntityData`. This will take in your POD
type and register it for the editor to use. For example:
```go
func init() {
	engine.RegisterEntityData(&MyModuleStructure{})
}
```

## Supported POD types
Below are a list of POD types that you can use for your structures.

|  Types  |         |           |            |
| ------- | ------- | --------- | ---------- |
| int     | int16   | int32     | int64      |
| uint    | uint16  | uint32    | uint64     |
| float32 | float64 | complex64 | complex128 |
| bool    | string  | EntityId  | uintptr    |

## Struct tag decorators and constraints
Struct tags allow developers to tell the editor how to treat each field in the
structure when the developer is interacting with it. They are defined using
standard go structure tagging:

```go
type SomeEntityDataModule struct {
	Speed     float32 `clamp:"3,1,30"`
	MaxCount  int     `default:"15"`
	MaxHeight float32 `default:"3.14"`
	IsPrimary bool    `default:"false"`
}
```

|   Key   |      Arguments       | Description |
| ------- | -------------------- | ----------- |
| clamp   | number,number,number | Clamps the value between 2 numbers: default, min, max |
| default | any                  | Sets the default/starting value |
