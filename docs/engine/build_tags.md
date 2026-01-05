---
title: Build Tags | Kaiju Engine
---

# Build Tags
There are a few different build tags that are used by the engine to determine
what should be build from the codebase. Some will do major changes like `editor`
but others will do debugging/runtime changes.

|        Tag         | Description |
| ------------------ | ----------- |
| `editor`           | Used to build the editor, otherwise the runtime will be built |
| `debug`            | Used to enable various debug systems for the editor/runtime |
| `shipping`         | Used to optimize functions and strip out debugging tools/functions  |
| `vulkanValidation` | Turns on Vulkan validation layers |