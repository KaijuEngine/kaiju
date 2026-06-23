---
title: Vulkan Validation Layers | Kaiju Engine
---

The renderer requests `VK_LAYER_KHRONOS_validation` automatically in `debug`
builds (`useValidationLayers = build.Debug` in `src/rendering/vk_config.go`). When
the layer is active, the Vulkan driver turns GPU misuse — a wrong image layout, a
destroyed object still in use, an out-of-date swap chain — into a precise, logged
`VUID-…` error. When it is **absent**, that same misuse instead surfaces as a raw
`SIGSEGV`/`SIGBUS` deep inside the driver (`callVkQueueSubmit` and friends), which is
far harder to diagnose.

## macOS: routing through the Vulkan loader (required for layers)

On macOS the engine loads **MoltenVK directly** by default (`-lMoltenVK`,
`dlopen("libMoltenVK.dylib")`). That bypasses the Vulkan **loader**
(`libvulkan.dylib`) — and **layers are inserted by the loader**, so they can never be
enumerated on the direct path. You will see:

```
Could not find validation layer  layer=VK_LAYER_KHRONOS_validation
```

To run validation you must route through the loader. This is **opt-in** so the
default/release path is unchanged — set `KAIJU_VULKAN_USE_LOADER=1`, which makes
`getDefaultProcAddr` prefer `libvulkan.dylib` (and makes the engine request the
`VK_KHR_portability_enumeration` instance extension the loader needs to expose
MoltenVK; without it `vkCreateInstance` returns `VK_ERROR_INCOMPATIBLE_DRIVER`/`-9`).

### Setup (Homebrew)

```sh
brew install vulkan-loader molten-vk vulkan-validationlayers vulkan-tools

export KAIJU_VULKAN_USE_LOADER=1
# point the loader at the MoltenVK ICD manifest (path from `find $(brew --prefix) -name MoltenVK_icd.json`)
export VK_ICD_FILENAMES=/usr/local/etc/vulkan/icd.d/MoltenVK_icd.json
# and at the validation layer manifest
export VK_ADD_LAYER_PATH="$(brew --prefix)/share/vulkan/explicit_layer.d"
# optional, if dlopen can't find libvulkan.dylib:
# export DYLD_LIBRARY_PATH="$(brew --prefix)/lib"
```

Verify the loader+ICD first with `vulkaninfo` (it should list your GPU, no `-9`).
Then run a `debug` build of the app from the same shell — startup logs
`enabling the validation layers` instead of the warning, and `VUID-…` errors print
before any crash. (On Windows/Linux the loader is used already; just install the SDK
and set `VK_ADD_LAYER_PATH`.)

## Reproducing GPU bugs

With validation active, drive the scenario under test (e.g. rapidly drag-resize the
window). Misuse is reported by VUID, object handle, and call site instead of crashing
— for example the depth/stencil resize crash was found this way:

```
VUID-VkImageMemoryBarrier-image-03320: depth/stencil image (D24_UNORM_S8_UINT)
  transitioned with aspectMask = DEPTH_BIT only
VUID-vkCmdDraw-None-09600: vkQueueSubmit … expects VkImage (aspectMask = STENCIL_BIT)
  to be DEPTH_STENCIL_ATTACHMENT_OPTIMAL — instead current layout is UNDEFINED
```

## macOS: Metal API Validation (complementary)

Metal API Validation catches Metal-level misuse the Vulkan layers can't see (e.g.
`CAMetalLayer`/drawable threading). It needs no loader setup — run with
`MTL_DEBUG_LAYER=1 METAL_DEVICE_WRAPPER_TYPE=1` (or enable it in the Xcode scheme).
It is what first surfaced the climbing `spvDescriptorSet0` buffer offset that pointed
at the descriptor-pool growth during resize.
