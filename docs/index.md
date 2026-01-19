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
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng" />
		</video>
	</div>
	<div>
		The editor is a testament to the engine's flexibility, because the editor itself is a game running in the engine. Built entirely using Kaiju, it demonstrates the power and versatility of the engine while providing a full-featured development environment. From asset management and scene composition to shader editing and particle system design, the editor showcases what's possible when you build your tools with your own technology. This approach ensures that any feature available to game developers is also available to the editor itself, creating a tight feedback loop between engine development and practical usage. The editor serves as both a powerful tool for game creation and a living proof of concept for the engine's capabilities.
	</div>
</div>

## 2D

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.apng" />
		</video>
	</div>
	<div>
		Making 2D games is as simple as switching the editor to "2D" mode. The engine handles sprite rendering, layering, and all the essentials you need for 2D game development with the same high-performance Vulkan backend that powers the 3D features. Whether you're creating pixel art platformers, hand-drawn adventures, or modern 2D action games, Kaiju provides the tools to bring your 2D vision to life with minimal friction. The 2D workflow integrates seamlessly with the engine's animation system, particle effects, and UI tools, allowing you to create rich, polished 2D experiences. You can even mix 2D and 3D elements in the same scene when creative opportunities arise, giving you the flexibility to push beyond traditional 2D boundaries.
	</div>
</div>

## 3D

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.apng" />
		</video>
	</div>
	<div>
		A completely custom built math library backs the 3D rendering, optimized specifically for game development needs and performance. The engine leverages Vulkan for high-performance graphics, giving you access to modern rendering techniques and the full power of contemporary GPUs. From simple 3D scenes to complex environments with dynamic lighting and shadows, the engine provides the foundation for stunning visual experiences. The 3D system is designed to be both powerful and accessible, with sensible defaults that let you get started quickly while still providing the low-level control needed for advanced rendering techniques. Whether you're building stylized worlds or aiming for photorealistic visuals, the rendering pipeline adapts to your artistic vision without getting in the way.
	</div>
</div>

## Particle systems

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.apng" />
		</video>
	</div>
	<div>
		Compose multiple particle emitters into a system for stunning visual effects that bring your game to life. Create everything from simple sparkles and trails to complex explosions, smoke plumes, magical effects, and environmental atmosphere by combining emitters with different behaviors, textures, and physical properties. The particle system is designed for performance, allowing you to create rich visual feedback without compromising frame rates, even with hundreds of thousands of particles on screen. Each emitter can be individually configured with properties like velocity, color gradients, size curves, and lifespan, then combined into systems that create cohesive, impressive effects. The visual editor makes it easy to tweak and preview particle effects in real-time, so you can iterate quickly and achieve exactly the look you're going for.
	</div>
</div>

## Animation

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/animation.apng" />
		</video>
	</div>
	<div>
		Full skeletal skinning, 2D sprite sheets, flip books, and material animations give you complete control over motion in your games. Animate characters with bone-based rigs for smooth, realistic movement, or use the skeletal system for procedural animation and inverse kinematics. Create smooth sprite-based animations with sprite sheets and flip books for 2D characters and effects, with support for variable frame rates and animation blending. Bring materials to life with animated properties like scrolling textures, pulsing emissive values, and morphing shader parameters. The animation system is flexible enough to handle any style of game, from realistic character movement to stylized effects and everything in between. Animations can be triggered programmatically, blended together for smooth transitions, and synchronized with game events to create responsive, dynamic experiences.
	</div>
</div>

## UI

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/ui.apng" />
		</video>
	</div>
	<div>
		A very fast, completely custom-built, retained-mode UI with the option of using HTML/CSS for markup provides the best of both worlds. This hybrid approach gives you the performance benefits of a retained-mode system—where the UI state is managed efficiently in memory—while allowing the rapid prototyping and familiar workflow of HTML/CSS when you need it. Build responsive menus, heads-up displays, interactive interfaces, and complex UI layouts that feel smooth and look great, all rendered with the same high-performance Vulkan backend as your game. The UI system integrates seamlessly with the engine's data binding capabilities, making it easy to connect your game state to visual elements without writing boilerplate code. Whether you're creating minimalist HUDs or elaborate menu systems with animations and transitions, the UI framework provides the tools you need while staying out of your way.
	</div>
</div>

## Physics

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/physics.apng" />
		</video>
	</div>
	<div>
		Simulate your worlds with 3D physics, including rigid body dynamics, collision detection, and physically accurate responses. Create realistic object interactions, implement character controllers with proper collision handling, or build complex physics-based puzzles and mechanics that respond naturally to player actions. The physics engine integrates seamlessly with the rendering system, making it easy to create physically interactive environments where objects tumble, stack, and collide with convincing realism. Support for various collision shapes—from simple boxes and spheres to complex mesh colliders—gives you the flexibility to balance performance with precision. The physics system can be tuned for different gameplay styles, whether you need arcade-style responsiveness or simulation-level accuracy, and integrates with the animation system for ragdoll physics and dynamic character responses.
	</div>
</div>

## Live shader updates

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.apng" />
		</video>
	</div>
	<div>
		Easily visualize your <a href="https://www.khronos.org/opengl/wiki/Core_Language_(GLSL)">GLSL</a> shader code in real time as you write it, dramatically accelerating your shader development workflow. The live shader update system provides immediate feedback, letting you see the visual results of your code changes without any recompilation or relaunching. No more waiting through long build cycles just to see if that gradient looks right or that distortion effect works as intended—see your changes instantly and iterate quickly on visual effects. The system works with both vertex and fragment shaders, supporting the full GLSL specification and providing helpful error messages when syntax issues arise. This tight feedback loop enables rapid experimentation and learning, making shader programming more accessible while still providing the power and control that experienced graphics programmers expect. Whether you're creating stylized toon shaders, complex post-processing effects, or experimental visual techniques, the live update system keeps you in the creative flow.
	</div>
</div>

## Audio

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/audio.apng" />
		</video>
	</div>
	<div>
		Play sounds and music, even in 3D space, powered by <a href="https://solhsa.com/soloud/">SoLoud</a>-a robust, proven audio engine that handles the complexities of game audio. Position audio sources anywhere in your game world for immersive spatial audio that responds to listener position and orientation, creating convincing soundscapes where sounds naturally fade with distance and pan based on direction. Play traditional 2D sounds for UI feedback, background music, and ambient effects that stay consistent regardless of camera position. The audio system handles streaming for large music files, manages multiple simultaneous sounds efficiently, and provides a straightforward API for common audio tasks like volume control, playback speed adjustment, and sound filtering. With support for various audio formats and the ability to create audio groups for mixing control, you have all the tools needed to create a polished audio experience that enhances your game's atmosphere and feedback.
	</div>
</div>

## Cross platform

<div class="indexHighlight">
	<div>
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/cross_platform.png" />
	</div>
	<div>
		Create on Windows, Linux and Mac. Deploy to Windows, Linux, Mac, and Android with more platforms coming soon. Build your game once and reach players across multiple platforms without maintaining separate codebases or dealing with platform-specific quirks. The engine abstracts platform differences in rendering, input handling, file systems, and other low-level details, allowing you to focus on your game logic while Kaiju handles the platform-specific implementation details. The same code runs across all supported platforms, with automatic handling of differences in graphics APIs, window management, and system integration. This write-once, deploy-many approach dramatically reduces the complexity of multi-platform development and ensures that your game delivers a consistent experience regardless of where players encounter it. Future platform support is actively in development, expanding the potential audience for your games even further.
	</div>
</div>

## Development velocity

<div class="indexHighlight">
	<div>
		<video autoplay muted loop playsinline width="100%">
		<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.mp4" type="video/mp4">
		<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/development_velocity.apng" />
		</video>
	</div>
	<div>
		Unmatched edit-build-launch speed keeps you in the creative flow. Iterate quickly with incredibly fast build times that keep you focused on making your game rather than waiting for compilation. The engine is designed from the ground up for rapid development cycles—make a change, build in seconds (not minutes), and see the results immediately in your running game. This fast iteration loop is critical for maintaining creative momentum and allows for an experimental, playful approach to game development where you can try ideas quickly without the friction of long waits. Combined with hot-reloading capabilities for many asset types and the live shader update system, you spend more time creating and less time waiting. The <a href="https://go.dev/">Go</a> language's fast compilation times and the engine's efficient build system work together to provide one of the fastest development experiences available in modern game engines, letting you maintain the flow state that's essential for creative work.
	</div>
</div>

## Support the project
If you like what you see, and want to support the project's continued development, please consider [becoming a sponsor](https://github.com/sponsors/BrentFarris).

<iframe src="https://github.com/sponsors/BrentFarris/button" title="Sponsor BrentFarris" height="32" width="114" style="border: 0; border-radius: 6px;"></iframe>

## Join the community
- [GitHub repository](https://github.com/KaijuEngine/kaiju)
- [Kaiju creator on X/Twitter](https://twitter.com/ShieldCrush)
- [Discord server](https://discord.gg/8rFPEu8U52)
