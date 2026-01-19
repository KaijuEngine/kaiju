# Kaiju Engine
Kaiju is a 2D/3D game engine written in Go (Golang) backed by Vulkan. The goal of the engine is to use a modern, easy, systems level programming language, with a focus on simplicity, to create a new kind of game engine.

- 📄 2D / 🧊 3D Game Engine
- 🖥️ Built-in editor 
- 🪟 Windows, 🐧 Linux, 🍎 Mac (support is [currently WIP](https://github.com/KaijuEngine/kaiju/pull/489)), 🤖 Android
- 🎊 Particle systems
- ⏯️ 2D/3D Animation
- 🎶 Music, SFX, and 3D sound sources
- 🔣 UI - Custom built system with optional HTML/CSS markup
- 🪄 Live shader updates on GLSL changes
- 🔌 Editor plugins support (via Go)
- 🚚 Faster builds than other game engines
- 🔥 Better performance than other game engines
- ⚠️🚧🏗️👷‍♂️ Work in progress, under heavy development

## Sponsor the project
If you like the project, and would like to support it, please consider [becomming a sponsor on GitHub](https://github.com/sponsors/BrentFarris).

## Join the community
- [GitHub repository](https://github.com/KaijuEngine/kaiju)
- [Mailing list](https://www.freelists.org/list/kaijuengine) <- Recommended for detailed updates
- [Discord server](https://discord.gg/8rFPEu8U52)
- [Brent Farris on X/Twitter](https://twitter.com/ShieldCrush)

## ⚠️ WORK IN PROGRESS ⚠️
Though the engine is production ready, the editor **_is not_**, feel free to join and contribute to its development.

For the latest updates, please join the [Discord](https://discord.gg/HYj7Dh7ke3) or check my [Twitter/X](https://twitter.com/ShieldCrush).

Please review the Ad-Hoc [editor readme](https://github.com/KaijuEngine/kaiju/blob/master/src/editor/README.md)

## Getting started building the engine/editor
You can choose to get running quickly by recursively cloning the repository. This will also download the [src/libs submodule](https://github.com/KaijuEngine/kaiju_prebuilts) that includes all the pre-built library files needed to compile. Or, you can build the dependency libraries yourself.

### Clone the Repository with pre-built libraries
```sh
git clone --recurse-submodules https://github.com/KaijuEngine/kaiju.git
```

If you have Go, C build tools, platform libs, and Vulkan setup, you can start by running:
```sh
cd src
go build -tags="debug,editor" -o ../ ./
```

*Or just open the repository in VSCode (or other IDE) and begin debugging it.*

If your environment isn't setup, check out [this doc](https://github.com/KaijuEngine/kaiju/blob/master/docs/engine/build_from_source.md#prerequisites) on getting it setup. You can skip the library building steps (Soloud and Bullet3), you already have this libs from the submodule clone.

### Clone the Repository without pre-built libaries
```sh
git clone https://github.com/KaijuEngine/kaiju.git
```

If you clone in this way, you'll need to manually build the library dependencies yourself. Please view [this doc](https://github.com/KaijuEngine/kaiju/blob/master/docs/engine/build_from_source.md#building-soloud) for how to build Soloud and [this doc](https://github.com/KaijuEngine/kaiju/blob/master/docs/engine/build_from_source.md#building-bullet3) for how to build Bullet3.






## Documentation
To run the documentation locally, you need to install `mkdocs` and `mkdocs-material`.

1. Install dependencies:
```sh
pip install mkdocs mkdocs-material
```

2. Serve the documentation in watch mode:
```sh
mkdocs serve -w ./docs
```


# Kaiju Engine
An open source game engine and editor, written in Go backed by Vulkan.

## An editor, built in the engine
The editor is a testament to the engine's flexibility, because the editor itself is a game running in the engine.

[editor.mp4](https://github.com/user-attachments/assets/d45511a2-2e22-4f47-a738-4affdd1cfc45)

## 2D
Making 2D games is as simple as switching the editor to "2D" mode.

[2d.mp4](https://github.com/user-attachments/assets/a3b1b53f-43ce-47bc-b1a7-1aa43c25e1a0)

## 3D
A completely custom built math library backs the 3D rendering.

[3d.mp4](https://github.com/user-attachments/assets/7b5b1eb3-06ba-4827-8399-525b40d1cf09)

## Particle systems
Compose multiple particle emitters into a system for stunning visual effects.

[particle_systems.mp4](https://github.com/user-attachments/assets/09331b78-f426-47c1-ba62-b1b896f5259a)

## Animation
Full skeletal skinning, 2D sprite sheets, flip books, and material animations.

[animation](https://github.com/user-attachments/assets/4e9bb101-cb09-40c3-bb03-f2a1207a04f9)

## UI
A very fast, completely custom-built, retained-mode UI with the option of using HTML/CSS for markup.

[ui.mp4](https://github.com/user-attachments/assets/468b64c9-fb30-4b8a-83cf-1c7feee1a119)

## Physics
Simulate your worlds with 3D physics.

[physics.mp4](https://github.com/user-attachments/assets/3bd43af8-169e-405b-bd6a-44fbfc939afd)

## Live shader updates
Easily visualize your GLSL shader code in real time.

[live_shader.mp4](https://github.com/user-attachments/assets/4b715014-ccc7-49f4-9740-d717a820665b)

## Development velocity
Unmatched edit-build-launch speed. Iterate quickly with incredibly fast build times.

[development_velocity.mp4](https://github.com/user-attachments/assets/36bd06e8-dbe0-40ae-ab6a-8e8515949942)

## Cross platform
Create on Windows, Linux and Mac.

Deploy to Windows, Linux, Mac, and Android (more platforms added soon).

<img width="1280" height="720" alt="cross_platform" src="https://github.com/user-attachments/assets/75e56325-54aa-4133-8902-f1fd987c44f3" />

## Audio
Play sounds and music, even in 3D space, powered by Soloud.

## Editor overview
[(YouTube) Kaiju Engine Editor Introduction](https://www.youtube.com/watch?v=cmjX_M6lEZE)

### Editor plugins
[kaiju-editor-plugins.mp4](https://github.com/user-attachments/assets/4c7b7c65-f77b-47de-8d45-175dcb421afa)

## Why Kaiju?
The current version of the base engine renders extremely fast, faster than most would think a garbage collected language could go. In my testing a release mode build of a game in Unity with nothing but a black background and a cube runs at about 1,600 FPS. In Kaiju, the same thing runs at around 5,400 FPS on the same machine. In fact, a complete game, with audio, custom cursors, real time PBR rendering with real time shadows, UI, and more runs at 2,712 FPS (in "debug" mode) [screenshots or it didn't happen](https://x.com/ShieldCrush/status/1943516032674537958).

## Why Go (golang)?
I love C, and because I love C and found out that Ken Thompson played a part in designing Go, I gave Go a chance. It has been such a joy to use and work with I decided to port my C game engine to Go. Go is a modern system-level language that allows me to write code the way I want to write code and even have the opportunity to do some crazy things if I want to (no strings attached). Also the simplicity and "just works" of writing Assembly code was a great boost to my happiness.

What's more, it's a language that other developers can easily learn and jump right into extending the engine/editor. No need for developers to re-figure out some bespoke macros or crazy templating nonsense. It's flat, easy, straight forward, and the foot-gun is hidden behind some walls, but there if you want it. Furthermore, developers can write their games in Go directly, no need for some alternative language that is different from the engine code (but we'll include Lua for modding).

## What about the Garbage Collector?!
I am creating this section because I get asked about it when I mention "Go", possibly not realizing that most public game engines use a garbage collector (GC).

The GC is actually a feature I'm happy with (shocker coming from a C guy). Well, the reason is simple, if you're going to make a game engine that the public will use and needs to be stable, you need a garbage collector. Unity has C# (and possibly an internal GC as well), Unreal has a GC (and it could use a tune up if you ask me), Godot has a GC albeit their scripting language or when you use C#. It is actually very important for public engines to have a GC because people are only human and make a lot of mistakes, mistakes they'll blame on you (the engine developer) before they blame themselves.

Coincidentally, the overall design I have for the engine plays very well with the GC and last I measured, I have a net-0 heap allocation while running (may need a new review). If you don't abuse the GC, you shouldn't generally feel it, it runs concurrently as well.

I'll be the first to admit, I think the developers of Go can create a better GC than I can, and probably better than Unreal and Unity too.

## Star history
[![Star History Chart](https://api.star-history.com/svg?repos=KaijuEngine/kaiju&type=Date)](https://star-history.com/#KaijuEngine/kaiju&Date)   
