---
title: Shader Definition | Kaiju Engine
---

# Shader Definition
Shaders in Vulkan can be a bit complex, so we have developed the foundational
tools to aid in the development and usage of shaders. In your content folder
there is a folder named "shaders", this folder should contain a folder named
"definitions". Inside of this folder is where you create a simple definition
file for your shader.

## Setting up a new shader definition
A shader definition contains all of the information about the shaders that will
be combined to create a shader pipeline. To start, you will need to create a
new shader definition `.json` file. This file should just contain one object
named "Vulkan" and inside of it, fill out the shaders that make up the pipeline.

So for example, the basic shader definition would be created as:
```json
{
	"Vulkan": {
		"Vert": "renderer/spv/basic.vert.spv",
		"Frag": "renderer/spv/basic.frag.spv",
		"Geom": "",
		"Tesc": "",
		"Tese": "",
	}
}
```

This tells the system what shaders you will need for this pipeline. With just
the description of which shaders to use, you are ready to generate the rest of
the information directly from your supplied shaders.

## Generate shader definition
To generate the shader definitions, you can run the go generator tool found in
```sh
src/generators/shader_definition/main.go
```

If you are in visual studio, you can run this by selecting the "Shader
Definition Generator" launch setting and running it. This will go through all
of the definition files within `content/renderer/definitions` and update the
`.json` files to contain all the information needed about the shader. With this,
the engine will know how to load up your shader and setup the pipeline layout
for your shader. It will also setup any named buffer objects so that you can
update the data within them in your code easily by using their name.