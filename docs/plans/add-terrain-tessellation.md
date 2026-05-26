# Add Tessellation to Terrain Shader for Smoother Terrain Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Add Vulkan tessellation shaders (TCS and TES) to the terrain rendering pipeline to dynamically subdivide terrain meshes near the camera, resulting in smoother silhouettes, better normal interpolation, and reduced blockiness without increasing base mesh resolution.

**Architecture:** 
- Leverage existing SPIR-V generator support for TessellationControl/Evaluation fields in .shader JSONs.
- Change pipeline from Triangles to PatchList (3 control points) in terrain.shaderpipeline.
- Create terrain.tesc (control shader that computes per-patch tessellation factors based on distance to camera) and terrain.tese (evaluation shader that interpolates positions, normals, UVs using barycentric coordinates for smooth surface).
- Update terrain.vert to output all necessary per-control-point data (it becomes the per-vertex stage for patches).
- Update terrain.frag if needed for new inputs from TES.
- Add distance-based tessellation factor calculation (min/max levels, edge factors).
- Update terrain.go if mesh generation or drawing needs adjustment for tessellated topology (likely minimal).
- Use TDD where possible with existing terrain tests; add shader compilation test.
- Ensure editor terrain tools (painting) continue to work with updated shaders.

**Tech Stack:** Go, GLSL 460, Vulkan tessellation (triangles), matrix lib, existing kaiju.glsl helpers, SPIR-V compiler in generators/spirv.

**Constraints:** 
- Do not change custom math library usage.
- Maintain compatibility with existing terrain heightfield, painting, physics.
- Keep base mesh resolution the same (tessellation adds detail at runtime).
- Support both lit and unlit terrain variants.
- YAGNI: No PN-triangles or displacement mapping yet — simple barycentric interpolation + distance-based uniform tess factor for now.
- DRY: Extract common tessellation helpers into kaiju.glsl if possible.

---

### Task 1: Explore and Document Current Terrain Shader Setup

**Objective:** Understand current files, how shaders are compiled, loaded, and used for terrain to avoid breaking existing functionality.

**Files:**
- Modify: `src/engine/terrain/terrain.go` (read only for now)
- Read: `src/editor/editor_embedded_content/editor_content/renderer/shaders/terrain.shader`
- Read: `src/editor/editor_embedded_content/editor_content/renderer/src/terrain.vert`
- Read: `src/editor/editor_embedded_content/editor_content/renderer/src/terrain.frag`
- Read: `src/editor/editor_embedded_content/editor_content/renderer/pipelines/terrain.shaderpipeline`
- Read: `src/generators/spirv/main.go` (for tess support)
- Read: `src/rendering/*` files related to shader cache and pipeline creation

**Step 1: Run searches to map all terrain shader references**
```bash
search_files --pattern "terrain\.(shader|vert|frag|shaderpipeline)" --path src
search_files --pattern "Tessellation|PatchControlPoints|tesc|tese" --path src
```

**Step 2: Read key files and note current tess fields (they are empty)**
Use read_file tool on the paths above. Document in comments or a temp note file what per-vertex attributes are passed (many LAYOUT_ defines for slope, tints, brush, etc.).

**Step 3: Verify current build works**
```bash
cd src
go build -tags="debug,editor" -o ../kaiju.exe ./
```
Expected: Builds successfully. Run a test scene with terrain if available.

**Step 4: Commit**
```bash
git add -A
git commit -m "docs: explore current terrain shader setup before adding tessellation"
```

**Verification:** Have notes on all files that must be updated. No breakage.

---

### Task 2: Add Tessellation Helper Functions to kaiju.glsl

**Objective:** Add reusable functions for calculating tessellation factors based on distance to camera and barycentric interpolation to avoid code duplication.

**Files:**
- Modify: `src/editor/editor_embedded_content/editor_content/renderer/src/kaiju.glsl:append at end`

**Step 1: Write failing "test" by adding comment expecting functions (since no unit test for GLSL easily)**
Add comment block describing expected functions:
```glsl
// Tessellation helpers (to be implemented):
// float tessFactor(vec3 pos, vec3 cameraPos, float minTess, float maxTess, float distScale)
// vec3 barycentricInterpolate(vec3 a, vec3 b, vec3 c, vec3 bary)
// ... similar for vec2, vec4
```

**Step 2: Run "test"**
Manually inspect that comment is there (no compile yet).

**Step 3: Implement minimal helpers**
Add the following to kaiju.glsl:
```glsl
float calcTessFactor(vec3 worldPos, vec3 camPos, float minTess, float maxTess, float distNear, float distFar) {
    float dist = distance(worldPos, camPos);
    float factor = (distFar - dist) / (distFar - distNear);
    factor = clamp(factor, 0.0, 1.0);
    return mix(minTess, maxTess, factor * factor);  // quadratic falloff
}

vec3 interpolate3(vec3 a, vec3 b, vec3 c, vec3 bary) {
    return a * bary.x + b * bary.y + c * bary.z;
}

vec2 interpolate2(vec2 a, vec2 b, vec2 c, vec3 bary) {
    return a * bary.x + b * bary.y + c * bary.z;
}

// Add more as needed for normals, colors, etc.
```

**Step 4: Verify by checking syntax (no GLSL compiler easy, but check include in vert)**
Recompile shaders later in next tasks.

**Step 5: Commit**
```bash
git add src/editor/editor_embedded_content/editor_content/renderer/src/kaiju.glsl
git commit -m "feat: add tessellation helper functions to kaiju.glsl"
```

**Verification:** Helpers are general and can be used in TCS/TES.

---

### Task 3: Create Terrain Tessellation Control Shader (TCS)

**Objective:** Create terrain.tesc that sets tessellation levels per patch based on distance, passes through all per-vertex data.

**Files:**
- Create: `src/editor/editor_embedded_content/editor_content/renderer/src/terrain.tesc`

**Step 1: Write the full TCS code (no separate test, as it's shader)**
```glsl
#version 460
#define TESSELLATION_CONTROL_SHADER

#include "kaiju.glsl"

layout(vertices = 3) out;  // Triangle patches

// All the in/out from vertex (define matching inputs from vert)
in vec3 fragColor[];  // Note: use arrays for per-vertex in TCS
// ... define all other ins like fragNormal[], fragTexCoords[], fragSlopeParams[] etc. (match LAYOUTs)

// Uniforms for tess levels (add to UBO later if needed, or use constants)
uniform float minTessLevel = 1.0;
uniform float maxTessLevel = 8.0;
uniform float tessDistanceNear = 5.0;
uniform float tessDistanceFar = 50.0;

void main() {
    // Pass through control points
    gl_out[gl_InvocationID].gl_Position = gl_in[gl_InvocationID].gl_Position;
    // Copy all other varyings: fragColor[gl_InvocationID] = fragColor[gl_InvocationID]; etc for all

    // Calculate tess factors only once per patch (invocation 0)
    if (gl_InvocationID == 0) {
        vec3 camPos = cameraPosition.xyz;  // from UBO
        vec3 p0 = worldPositionFrom(gl_in[0].gl_Position); // helper
        vec3 p1 = worldPositionFrom(gl_in[1].gl_Position);
        vec3 p2 = worldPositionFrom(gl_in[2].gl_Position);
        
        float f0 = calcTessFactor(p0, camPos, minTessLevel, maxTessLevel, tessDistanceNear, tessDistanceFar);
        float f1 = calcTessFactor(p1, camPos, minTessLevel, maxTessLevel, tessDistanceNear, tessDistanceFar);
        float f2 = calcTessFactor(p2, camPos, minTessLevel, maxTessLevel, tessDistanceNear, tessDistanceFar);
        
        gl_TessLevelOuter[0] = f1;  // opposite to vertex 0?
        gl_TessLevelOuter[1] = f2;
        gl_TessLevelOuter[2] = f0;
        gl_TessLevelInner[0] = (f0 + f1 + f2) / 3.0;
    }
}
```
(Note: Complete all in/outs matching the many LAYOUT_FRAG_* from vert. Use the LAYOUT defines.)

**Step 2: Update terrain.shader to reference it**
Use patch tool or edit the JSON to set:
"TessellationControl": "terrain.tesc",
"TessellationControlFlags": "",
"TessellationEvaluation": "terrain.tese",  (will create next)
"TessellationEvaluationFlags": "",

**Step 3: Run generator to compile new SPIRV**
```bash
cd src/generators/spirv
go run main.go
```
Expected: Generates spv files without error for terrain.

**Step 4: Commit**
```bash
git add src/editor/editor_embedded_content/editor_content/renderer/src/terrain.tesc src/editor/editor_embedded_content/editor_content/renderer/shaders/terrain.shader
git commit -m "feat: add terrain.tesc for distance-based tessellation factors"
```

**Verification:** Shader compiles, no SPIRV errors. Check generated spv file exists.

---

### Task 4: Create Terrain Tessellation Evaluation Shader (TES)

**Objective:** Create terrain.tese that evaluates subdivided points using barycentric coords, recomputes normals if needed for smoothness.

**Files:**
- Create: `src/editor/editor_embedded_content/editor_content/renderer/src/terrain.tese`

**Step 1: Write the full TES code**
```glsl
#version 460
#define TESSELLATION_EVALUATION_SHADER

#include "kaiju.glsl"

layout(triangles, equal_spacing, cw) in;  // or fractional_odd for smoother

// Inputs from TCS (per patch)
in vec3 fragColor[];
// ... all other per-control point inputs

out vec3 fragNormal;  // etc for all outputs to fragment

void main() {
    vec3 bary = gl_TessCoord;
    
    // Interpolate all attributes
    vec4 pos = interpolate4(gl_in[0].gl_Position, gl_in[1].gl_Position, gl_in[2].gl_Position, bary);  // use helpers
    // Interpolate UV, colors, normals, all frag* varyings using interpolate2/3/4
    
    gl_Position = projection * view * model * pos;  // or use writeStandardPosition helper if updated
    
    // For smoother normals, optionally recompute from height or use interpolated + normalize
    fragNormal = normalize(interpolate3(fragNormal[0], fragNormal[1], fragNormal[2], bary));  // from inputs
    
    // Copy other interpolated values to outputs
    fragTexCoords = ... ;
    // etc.
    
    fragViewDir = cameraPosition.xyz - (model * pos).xyz;
}
```
(Full implementation matching all outputs from vert. Extend helpers in kaiju.glsl as needed in previous task.)

**Step 2: Update terrain_unlit.shader similarly for the unlit variant (or share if possible).**

**Step 3: Re-run SPIRV generator and verify**
```bash
cd src/generators/spirv && go run main.go
ls -l ../editor_embedded_content/editor_content/renderer/spv/*terrain*
```

**Step 4: Commit**
```bash
git add src/editor/editor_embedded_content/editor_content/renderer/src/terrain.tese src/editor/editor_embedded_content/editor_content/renderer/shaders/terrain_unlit.shader
git commit -m "feat: add terrain.tese for barycentric evaluation and smooth interpolation"
```

**Verification:** Both shaders compile to SPIRV. No duplicate definitions.

---

### Task 5: Update Shader Pipeline for Tessellation

**Objective:** Configure the Vulkan pipeline to use patch primitives and specify patch control points.

**Files:**
- Modify: `src/editor/editor_embedded_content/editor_content/renderer/pipelines/terrain.shaderpipeline`
- Modify: `src/editor/editor_embedded_content/editor_content/renderer/pipelines/terrain_unlit...` if separate

**Step 1: Update JSON with tessellation settings**
Change:
"InputAssembly": {"Topology": "PatchList", "PrimitiveRestart": false},
"Tessellation": {"PatchControlPoints": "3"},

Keep other settings (cull, depth, etc.).

**Step 2: Validate by regenerating pipelines or checking in code**
If there's a pipeline generator, run it. Otherwise, the rendering code reads these JSONs.

**Step 3: Update rendering code if topology change requires it**
Check `src/rendering/` for where shaderpipeline is parsed and if PatchList is supported. Add if missing (likely is since generator supports tess).

**Step 4: Test compilation**
Rebuild the engine.

**Step 5: Commit**
```bash
git add src/editor/editor_embedded_content/editor_content/renderer/pipelines/terrain.shaderpipeline
git commit -m "refactor: update terrain pipeline to use PatchList with 3 control points for tessellation"
```

**Verification:** Pipeline JSON is valid, engine builds, no Vulkan validation errors when running with terrain.

---

### Task 6: Integrate Tessellation into Terrain Go Code and ShaderData

**Objective:** Ensure terrain model creation, drawing, and shader data pass any new uniforms for tess levels. Update material if needed.

**Files:**
- Modify: `src/engine/terrain/terrain.go:640-720` (New function area)
- Modify: `src/registry/shader_data_registry/shader_data_terrain.go` if new fields needed for tess params
- Modify: `src/rendering/gpu_pipeline.go` or wherever pipelines are created if tess not fully supported

**Step 1: Add tess level uniforms to ShaderDataTerrain if not dynamic**
Add fields like MinTess, MaxTess, etc. to struct and update Size(), Create().

**Step 2: Set defaults in terrain creation**
In New or Load, set tess levels based on config (add to TerrainConfig).

**Step 3: Write test for new config**
Extend terrain_test.go with test that checks shader data has tess params.

**Step 4: Run tests**
```bash
cd src
go test ./engine/terrain -run TestTerrain -count=1
```

**Step 5: Commit**
```bash
git add src/engine/terrain/terrain.go src/registry/shader_data_registry/shader_data_terrain.go
git commit -m "feat: integrate tessellation params into terrain shader data and model creation"
```

**Verification:** Tests pass, terrain renders with tessellation (visually smoother near camera, higher poly count in wireframe debug).

---

### Task 7: Update Vertex Shader for Tessellation Compatibility and Cleanup

**Objective:** Adjust terrain.vert to work as control point shader, remove any incompatible code, ensure all data is passed to TCS/TES.

**Files:**
- Modify: `src/editor/editor_embedded_content/editor_content/renderer/src/terrain.vert`

**Step 1: Refactor main() to output to gl_out or use standard helpers updated for tess.**
Add #ifdef VERTEX_SHADER or keep but ensure compatibility.

**Step 2: Recompile shaders and test.**

**Step 3: Add debug mode for wireframe tessellated terrain (use editor debug).**

**Step 4: Run full engine test with terrain in editor.**

**Step 5: Commit and update docs**
Update AGENTS.md with notes on terrain tessellation.

**Verification:** Terrain looks smoother, no visual artifacts, performance acceptable (tess factor capped), painting still works.

---

### Task 8: Final Testing, Optimization, and Documentation

**Objective:** Verify end-to-end, optimize tess factors, document the feature.

**Files:**
- Modify: `docs/AGENTS.md`
- Add tests if needed for performance.

**Steps:**
1. Run editor, load terrain, toggle wireframe to see subdivision.
2. Test at different distances — smooth close, coarse far.
3. Benchmark performance impact.
4. Add config options for tess levels in TerrainConfig.
5. Update any editor terrain workspace if affected.
6. Final commit with "feat: add dynamic tessellation to terrain shader for smoother rendering"

**Verification Commands:**
```bash
cd src
go test ./engine/terrain -run=Test
go build -tags="debug,editor" ./
# Run game/editor and visually inspect
```

**Expected Outcome:** Smoother terrain with dynamic detail, maintain all existing features.

---

**Plan complete.** This provides bite-sized, testable tasks with exact code snippets, commands, and verification. Ready for implementation via subagent-driven-development or manual execution. Shall I proceed with executing this plan using delegate_task for each task with reviews?
