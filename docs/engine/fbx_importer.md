---
title: FBX importer | Kaiju Engine
---

# FBX importer

Kaiju's FBX importer is intended for binary FBX assets exported from DCC tools
such as Blender. It follows the binary record/property structure documented by
Blender's reverse-engineered FBX notes and matches key parser behavior used by
Assimp: FBX 7500 and newer use 64-bit record headers, arrays may be
zlib-compressed, and nested records terminate with null sentinels.

## Supported Files

- Binary FBX only. ASCII FBX is rejected with `ASCII FBX is not supported yet`
  and a warning is logged.
- FBX 7400 is covered by the `monkey.fbx` fixture and uses 32-bit record
  headers.
- FBX 7500 and newer are parsed with 64-bit record headers.
- Scalar property types: bool, int16, int32, int64, float32, float64, string,
  and raw byte blocks.
- Array property types: float32, float64, int32, int64, bool, and byte arrays,
  including raw and zlib-compressed encodings.

## Imported Features

- Mesh geometry from `Geometry` objects containing `Vertices` and
  `PolygonVertexIndex`.
- Triangle fan triangulation for polygons with three or more corners.
- Normals, UVs, and vertex colors from layer elements using `ByPolygonVertex`,
  `ByVertice`, `ByVertex`, or `ByControlPoint` mapping.
- `Direct` and `IndexToDirect` layer references.
- Generated face normals when normal layer data is absent.
- Model hierarchy and local translation, rotation, and scale.
- Global axis settings and unit scale conversion.
- Model geometric transforms baked into attached mesh vertices.
- Materials connected to geometry or model nodes.
- External texture paths and embedded `Video` texture content.
- Skin clusters with up to four normalized influences per vertex.
- Blend shape vertex offsets.
- Animation curves for local translation, rotation, and scale channels.

## Known Limitations

- ASCII FBX is not imported.
- Unsupported layer mapping or reference modes fail mesh conversion with a
  descriptive error after logging a warning.
- Unsupported object classes are indexed as generic objects and ignored by
  conversion; a warning is logged with the class, node class, name, and id.
- The importer does not evaluate FBX constraints, pivots, pre/post rotations,
  poses, cameras, lights, NURBS, patches, line geometry, or layered material
  semantics.
- Texture usage is mapped to Kaiju slots by common FBX property names. Unknown
  texture usages are assigned fallback slots.
- Animation import samples raw keyed curves. Advanced FBX interpolation modes
  and animation layer blending are not evaluated.
- UVs are imported in the FBX coordinate convention currently used by Kaiju's
  texture path; do not flip V without updating the regression checklist below.

## Parser Hardening

The binary parser validates record and property bounds before reading or
allocating. Claimed property counts cannot exceed the property-list byte length.
Array element counts are checked against element stride and a decoded-size cap
before raw allocation or zlib decompression. Compressed arrays are read through a
limited reader so malformed files cannot expand past the declared decoded byte
length.

Malformed binary tests cover:

- record end offsets outside the file or before the record header
- property-list lengths outside the containing record
- impossible property counts
- oversized claimed array counts
- compressed payload lengths outside the property list
- invalid compressed payloads
- unterminated null record sentinels
- bad nested sentinels

## Regression Checklist

Use `src/editor/editor_embedded_content/editor_content/meshes/monkey.fbx` as the
baseline fixture. It is binary FBX 7400, uses 32-bit record headers,
zlib-compressed arrays, and contains one `Geometry` plus one `Model`. Suzanne
has 507 source vertices, 500 polygons, and triangulates to 2,904 index entries.

Before treating a loader change as safe:

- Run `go test ./rendering/loaders ./rendering/loaders/fbx`.
- Run `go test ./rendering/loaders/fbx -bench=Monkey -run=^$` when parser or
  mesh conversion allocation behavior changes.
- Build from `src` with
  `go build -tags="debug,editor,filedrop,rawsrc" -o ../ ./`.
- Run `../kaijuengine.com.exe -integrationtest=fbx_monkey`.
- Inspect `integration_test_fbx_monkey.png`.

Visual checks:

- Winding: Suzanne should render front faces from the test camera, without
  appearing inside-out.
- UV vertical orientation: textured fixtures should keep expected V
  orientation. If V flipping changes, update importer code and tests together.
- Axis settings: non-default `GlobalSettings` should convert positions,
  directions, rotations, and scale consistently.
- Compressed arrays: the monkey fixture must continue to parse through the zlib
  array path and keep the expected 2,904 index count.

## References

- [Blender binary FBX notes](https://code.blender.org/2013/08/fbx-binary-file-format-specification/)
- [Assimp FBX binary tokenizer](https://codebrowser.dev/qt6/qtquick3d/src/3rdparty/assimp/src/code/AssetLib/FBX/FBXBinaryTokenizer.cpp.html)
