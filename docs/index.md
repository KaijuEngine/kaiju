---
title: Kaiju Engine
---

# Kaiju Engine
An extremely fast, open source game engine and editor, written in [Go](https://go.dev/) backed by [Vulkan](https://www.vulkan.org/).

Kaiju is a powerful game development platform where all game scripting and logic is written in **[Go](https://go.dev/)**, combining the performance of a compiled language with the simplicity of modern syntax. Whether you prefer working with the **Kaiju Editor** (a visual interface with integrated Go code editing) or using the **Kaiju Engine** directly (pure Go code), you have the flexibility to build games your way.

The engine supports both **2D and 3D game development** with a comprehensive feature set including physics simulation, particle systems, skeletal animation, a custom UI framework with HTML/CSS support, spatial audio, and live shader updates. Built on [Vulkan](https://www.vulkan.org/) for high-performance rendering, Kaiju delivers fast build times and rapid iteration cycles, making it ideal for developers who want to spend more time creating and less time waiting.

## Editor

<div class="indexHighlight">
	<div>
		The editor is a testament to the engine's flexibility, because the editor itself is a game running in the engine.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng" />
		</video>
	</div>
</div>

## 2D

<div class="indexHighlight">
	<div>
		Making 2D games is as simple as switching the editor to "2D" mode.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.apng" />
		</video>
	</div>
</div>

## 3D

<div class="indexHighlight">
	<div>
		A completely custom built math library backs the 3D rendering.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.apng" />
		</video>
	</div>
</div>

## Particle systems

<div class="indexHighlight">
	<div>
		Compose multiple particle emitters into a system for stunning visual effects.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.apng" />
		</video>
	</div>
</div>

## Animation

<div class="indexHighlight">
	<div>
		Full skeletal skinning, 2D sprite sheets, flip books, and material animations.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.apng" />
		</video>
	</div>
</div>

## UI

<div class="indexHighlight">
	<div>
		A very fast, completely custom-built, retained-mode UI with the option of using HTML/CSS for markup.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.apng" />
		</video>
	</div>
</div>

## Physics

<div class="indexHighlight">
	<div>
		Simulate your worlds with 3D physics.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.apng" />
		</video>
	</div>
</div>

## Live shader updates

<div class="indexHighlight">
	<div>
		Easily visualize your GLSL shader code in real time.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.apng" />
		</video>
	</div>
</div>

## Audio

<div class="indexHighlight">
	<div>
		Play sounds and music, even in 3D space, powered by Soloud.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.apng" />
		</video>
	</div>
</div>

## Cross platform

<div class="indexHighlight">
	<div>
		Create on Windows, Linux and Mac.<br/>
		Deploy to Windows, Linux, Mac, and Android (more platforms added soon).
	</div>
	<div>
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/cross_platform.png" />
	</div>
</div>

## Development velocity

<div class="indexHighlight">
	<div>
		Unmatched edit-build-launch speed. Iterate quickly with incredibly fast build times.
	</div>
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.apng" />
		</video>
	</div>
</div>

## Sponsor the project
If you like what you see, and want to support the project's continued development, please consider [becoming a sponsor](https://github.com/sponsors/BrentFarris).

<iframe src="https://github.com/sponsors/BrentFarris/button" title="Sponsor BrentFarris" height="32" width="114" style="border: 0; border-radius: 6px;"></iframe>

## Join the community
- [GitHub repository](https://github.com/KaijuEngine/kaiju)
- [Kaiju creator on X/Twitter](https://twitter.com/ShieldCrush)
- [Discord server](https://discord.gg/8rFPEu8U52)

## Sponsors - Thank you!
<table id="sponsors">
	<tr>
		<th>Name</th>
		<th>GitHub</th>
	</tr>
</table>
