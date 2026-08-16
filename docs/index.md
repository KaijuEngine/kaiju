---
title: Kaiju Engine
description: Kaiju Engine is an extremely fast, open source game engine and editor written in Go, powered by Vulkan, supporting 2D/3D, physics, UI, and live shader updates.
keywords: game engine, Go, Vulkan, 2D, 3D, physics, UI, shader, cross-platform
---

<div class="index-logo" role="img" aria-label="Kaiju Engine"></div>

# Kaiju Engine

An extremely fast, open source game engine and editor, written in [Go](https://go.dev/) and backed by [Vulkan](https://www.vulkan.org/).

Kaiju is a game development platform where game scripting and logic is written in **[Go](https://go.dev/)**, combining the performance of a compiled language with the simplicity of modern syntax. You can work with the **Kaiju Editor**, a visual interface with integrated Go code editing, or use the **Kaiju Engine** directly from Go.

The engine supports both **2D and 3D game development**, including physics simulation, particle systems, skeletal animation, a custom UI framework with HTML/CSS support, spatial audio, and live shader updates. Kaiju is designed around fast build times and rapid iteration so you can spend more time creating and less time waiting.

## Getting started

There are two ways to start making games in Kaiju. Both paths use the same runtime, so choose the workflow that suits your project.

- [Start with the editor](getting_started/start_with_editor.md) for a visual workflow with scene, content, shading, VFX, and UI tools.
- [Start without the editor](getting_started/start_without_editor.md) to use Kaiju directly from a Go project.
- [Build Kaiju from source](engine/build_from_source.md) if you want to work on the engine or create your own editor build.
- [Download the latest release](https://github.com/KaijuEngine/kaiju/releases) to get a prebuilt version.

A typical first project follows four steps:

1. Download Kaiju or build it from source.
2. Create a project and organize its assets and scenes.
3. Build your first scene with objects, materials, cameras, UI, physics, and effects.
4. Write gameplay, systems, behavior, and tools in Go.

## News

Read the [Kaiju Engine blog](blog/index.md) for project updates and engineering deep dives.

## Editor

<div class="indexHighlight">
	<div>
		The editor is a testament to the engine's flexibility, because the editor itself is a game running in the engine. Use it to build scenes, inspect content, preview shaders and effects, and play test your project.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng" alt="The Kaiju editor">
		</video>
	</div>
</div>

## 2D

<div class="indexHighlight">
	<div>
		Making 2D games is as simple as switching the editor to 2D mode. Build with sprites, UI, animation, effects, and Go gameplay code.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.apng" alt="A 2D game in Kaiju">
		</video>
	</div>
</div>

## 3D

<div class="indexHighlight">
	<div>
		A completely custom-built math library backs Kaiju's native 3D rendering, scenes, lighting, and materials.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.apng" alt="A 3D scene in Kaiju">
		</video>
	</div>
</div>

## Particle systems

<div class="indexHighlight">
	<div>
		Compose multiple particle emitters into reusable systems for visual effects and gameplay feedback.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.apng" alt="Particle effects in Kaiju">
		</video>
	</div>
</div>

## Animation

<div class="indexHighlight">
	<div>
		Use full skeletal skinning, 2D sprite sheets, flip books, and material animations.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.apng" alt="Animation in Kaiju">
		</video>
	</div>
</div>

## UI

<div class="indexHighlight">
	<div>
		Kaiju includes a fast, custom-built, retained-mode UI with the option of using HTML/CSS-like markup. You can create game interfaces and engine tools with the same system.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.apng" alt="Kaiju's custom UI system">
		</video>
	</div>
</div>

## Physics

<div class="indexHighlight">
	<div>
		Simulate 3D worlds with physics systems that integrate with Kaiju entities and transforms.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.apng" alt="Physics simulation in Kaiju">
		</video>
	</div>
</div>

## Live shader updates

<div class="indexHighlight">
	<div>
		Edit GLSL shader code and visualize changes in real time while you iterate.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.apng" alt="Live shader editing in Kaiju">
		</video>
	</div>
</div>

## Audio

<div class="indexHighlight">
	<div>
		Play sounds and music, including audio positioned in 3D space, powered by SoLoud.
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.apng" alt="Spatial audio in Kaiju">
		</video>
	</div>
</div>

## Cross platform

<div class="indexHighlight">
	<div>
		Create on Windows, Linux, and macOS.<br>
		Deploy to Windows, Linux, macOS, and Android, with more platforms planned.
	</div>
	<div>
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/cross_platform.png" alt="Platforms supported by Kaiju">
	</div>
</div>

## Development velocity

<div class="indexHighlight">
	<div>
		Kaiju is designed for a short edit-build-launch loop. Change gameplay, build quickly, launch, test, and keep creating.

		<pre><code>cd src
go build -tags="debug,editor,filedrop" -o ../ ./

../kaijuengine.com.exe</code></pre>
	</div>
	<div>
		<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.apng">
			<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.mp4" type="video/mp4">
			<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.apng" alt="Kaiju's fast edit, build, and launch workflow">
		</video>
	</div>
</div>

## Sponsor the project

If you like what you see and want to support the project's continued development, please consider [becoming a sponsor](https://github.com/sponsors/BrentFarris).

<iframe src="https://github.com/sponsors/BrentFarris/button" title="Sponsor BrentFarris" height="32" width="114" style="border: 0; border-radius: 6px;"></iframe>

## Join the community

- [GitHub repository](https://github.com/KaijuEngine/kaiju)
- [Kaiju creator on X/Twitter](https://twitter.com/ShieldCrush)
- [Discord server](https://discord.gg/8rFPEu8U52)
