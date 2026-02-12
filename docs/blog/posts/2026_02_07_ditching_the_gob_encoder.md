---
title: Ditching the GOB encoder (and making my own)
description: Why the Go gob encoder breaks with empty or identical structs and how a custom POD encoder solves the problem using fully‑qualified type names.
tags: go, serialization, gob, pod, reflection, engine
image: images/simd_matrix.png
date: 2026-02-07
---

# The Problem with `gob`

For a while now, the editor has been using the [gob package](https://pkg.go.dev/encoding/gob) encoder for entity data binding. This worked pretty well for the most part. However, while making a game, I created 2 entries in the gob encoder that were empty structures; this broke everything.

For a refresher of what the [gob encoder](https://pkg.go.dev/encoding/gob) is, it's basically a replacement for what people often use [protocol buffers](https://en.wikipedia.org/wiki/Protocol_Buffers) (aka ProtoBuf) for. Using reflection, the package is able to efficiently map your structure/data to a binary format while being able to change/modify your structure. This is useful for storing the data as a file format or sending it over the internet. This is particularly useful because the structure/type is irrelevant, all that matters is the fields/data.

Focusing on that last sentence there, you might be able to see where things start to go wrong. When serializing types hidden within interfaces, you need to register that type with a call to something like `gob.Register`. This will register your concrete type so that it can be spawned from the data. There is one fatal flaw, types with identical fields will collide and the package will panic.

Take for example these two structures:
```go
type First struct {
    Index int
}

type Second struct {
    Index int
}
```

As far as Go's gob is concerned, these are both the same structure. This is fine in most reasonable cases, but not for the case I was trying to achieve in Kaiju. My goal in Kaiju is to make it so that we can serialize some POD structures that implement the `EntityData` interface.

```go
type EntityData interface {
	Init(entity *Entity, host *Host)
}
```

There are some cases where I want entity data to just trigger something to happen when an entity spawns into the stage. These structures will be empty structures with a name. Something like this:

```go
type GameBoardSpawn struct {}
```

This is actually where I hit the wall, when I had two empty structures. I then tested two different structures with identical fields and types and it too threw a panic.

# Custom POD encoder
Due to this limitation, I was forced to create my own encoder that will do what I need it to in this scenario. It is by no means the best implementation, but it solves my problem and fixes the collision issue. It too supports the ability to move fields around, as well as adding and removing fields, without breaking the serialization. It accomplishes this using reflection and matching package/struct names and the field names within the struct. So, not the most performant, but the encoding is done offline and the decoding runs very fast as the structures are typically very small (being that they are plain old data). You can check out the [encoder](https://github.com/KaijuEngine/kaiju/blob/master/src/engine/encoding/pod/pod_encoder.go) and [decoder](https://github.com/KaijuEngine/kaiju/blob/master/src/engine/encoding/pod/pod_decoder.go) in the repository for more details.

Below is a breakdown of the pod encoder.

1. **Uses fully‑qualified type names** (package + struct name) as the unique key, eliminating collisions for empty or identical structs.
2. **Stores a type‑lookup table** and a field‑lookup table in the binary header, so the format is self‑describing.
3. **Skips unsupported kinds** (pointers, interfaces, channels, functions, unsafe pointers) during encoding, because they cannot be reliably reconstructed without additional context.
4. **Supports slices, arrays, and nested structs** while preserving field order independence—fields can be added, removed, or reordered without breaking compatibility.

How It Works (High‑Level Overview)
| Step | Encoder | Decoder |
|-------|-----------|------------|
| **Header** | Write a slice of type keys (`[]string`) and a slice of field names (`[]string`). | Read the two lookup tables. |
| **Value** | For each value, write a **type id** (`uint8`). If the value is a slice/array, write the element count first. | Read the type ID, resolve it via the lookup table, then decode accordingly. |
| **Structs** | Write the number of encodable fields, then for each field: (1) Write the field‑lookup index (`uint16`). (2) Recursively encode the field value. | Read the field count, then for each field: (1) Resolve the field name via the lookup table. (2) Locate the struct field by name (reflection). (3) Recursively decode the field value. |
| **Primitives** | Directly write the binary representation (or string length + bytes for strings). | Directly read the binary representation. |

The encoder/decoder are in the `engine/encoding/pod` folder.
- `pod_encoder.go` - builds the lookup tables, writes the header, and recursively encodes values.
- `pod_decoder.go` - reads the header, resolves types, and recursively reconstructs values.

Both files rely on a **global registry** (a `sync.Map`) that maps a qualified type name to its `reflect.Type`. Registration is done via `pod.Register` and unregistration via `pod.Unregister`.

This new encoder is actually very simple thanks to reflection. There is actually more test code than actual encode/decode implementation code, which is also nice.