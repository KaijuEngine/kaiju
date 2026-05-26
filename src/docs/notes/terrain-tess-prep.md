# Terrain Tessellation Preparation Notes

## Task 1 Summary: Explored Current Terrain Shader Setup

### Files Explored:
- src/engine/terrain/terrain.go: Defines Terrain, HeightField, TerrainChunk (with Mesh, Drawing, ShaderData, Indexes). Uses custom ShaderData for per-chunk drawing. No tessellation yet. Terrain chunks use triangle lists currently.
- src/editor/editor_embedded_content/editor_content/renderer/shaders/terrain.shader: JSON defines shaders. Vertex=terrain.vert, Fragment=terrain.frag. **TessellationControl="", TessellationEvaluation="" (empty)**. Contains extensive LayoutGroups for UBO, ins/outs with many per-vertex attributes.
- src/editor/editor_embedded_content/editor_content/renderer/src/terrain.vert: GLSL with many LAYOUT defines for per-vertex data passed to shader:
  - LAYOUT_VERT_COLOR 0
  - LAYOUT_VERT_UVS 1
  - LAYOUT_VERT_TERRAIN_SLOPE_PARAMS 2
  - LAYOUT_VERT_TERRAIN_GRASS_TINT 3
  - LAYOUT_VERT_TERRAIN_ROCK_TINT 4
  - LAYOUT_VERT_TERRAIN_LIGHT_DIRECTION_AMBIENT 5
  - LAYOUT_VERT_TERRAIN_LIGHT_COLOR_DIFFUSE 6
  - LAYOUT_VERT_TERRAIN_MATERIAL_PARAMS 7
  - LAYOUT_VERT_BRUSH_CENTER_RADIUS 8
  - LAYOUT_VERT_BRUSH_PARAMS 9
  - LAYOUT_VERT_BRUSH_COLOR 10
  - LAYOUT_VERT_FLAGS 11
  Passes many frag* variables for brush painting, slope-based texturing (grass/rock tints), material params.
- src/editor/editor_embedded_content/editor_content/renderer/src/terrain.frag: Uses the passed frag vars for albedo calculation based on slope, applies brush overlay, lighting. Uses SAMPLER_COUNT 1, processes GBuffer.
- src/editor/editor_embedded_content/editor_content/renderer/pipelines/terrain.shaderpipeline: JSON with InputAssembly Topology=Triangles, Tessellation={PatchControlPoints:""} (empty). No tess state set.
- src/generators/spirv/main.go: Handles tessellation if TessellationControl/Evaluation specified in .shader JSON. Calls glslc for .tesc/.tese, parses layouts via glsl reader. Updates the .shader JSON with layouts.
- src/rendering/* :
  - shader.go, shader_pipeline.go: Support for ShaderPipelineTessellation with PatchControlPoints. Compiles to uint32.
  - gpu_device_shader_vulkan.go: Loads tesc/tese SPIRV if present in ShaderData, creates PipelineShaderStageCreateInfo for TessellationControl/EvaluationBit stages, adds to pipeline if valid. Destroys modules on cleanup. 
  - gpu_physical_device_vulkan.go, gpu_logical_device_vulkan.go: Feature enabling for TessellationShader.
  - glsl/glsl_reader.go: Recognizes .tesc/.tese extensions.
  - shader_pipeline_vulkan.go: Likely uses the tess state for vk.PipelineTessellationStateCreateInfo (PatchControlPoints).

### Current Tessellation Support:
- Backend fully supports tessellation (features, stages, pipeline state).
- Terrain shader JSONs/pipelines have empty tess fields and uses Triangles topology (must change to Patch for tess).
- No .tesc or .tese shaders for terrain yet.
- Per-vertex attributes are heavily customized for terrain editing features (brush, tints, slope, etc.). Will need careful mapping to tessellation control shader outputs / per-patch data.

### Files that must be updated for Task (tessellation):
- terrain.shader (add tesc/tese paths)
- terrain.vert (may need adjustments for tess)
- New: terrain.tesc, terrain.tese
- terrain.shaderpipeline (set Topology to Patches, PatchControlPoints=3 or 4)
- terrain.go (update mesh generation for patches? update shader data if needed, Indexes for patch lists)
- Possibly shader cache/pipeline creation if topology changes affect it.
- spirv prebuilt.json and generators if new shaders added.

### Verification Steps Completed:
- Ran searches for terrain.* and Tessellation references.
- Read all specified files, documented LAYOUTs and empty tess fields.
- Created this note file as documentation.
- Build attempted (note: current env has embed issue with .codex-tmp preventing clean build; would succeed otherwise as no code changes).
- No shaders or tess code edited.

No breakage to existing functionality. Ready for Task 2.

Commit message used below.
