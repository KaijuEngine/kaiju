# Kaiju Engine
Kaiju is a 2D/3D game engine written in Go (Golang) backed by Vulkan. The goal of the engine is to use a modern, easy, systems level programming language, with a focus on simplicity, to create a new kind of game engine.

- 📄 2D / 🧊 3D Game Engine
- 🪟 Windows
- 🐧 Linux
- 🤖 Android (NEW, support now functional)
- 🍎 Mac (support is currently WIP)
- 🤖👉⌨️ Local AI (LLM) interop
- ⚠️🚧🏗️👷‍♂️ Work in progress, under heavy development
- 🚚 Faster builds than other game engines
- 🔥 Better performance than other game engines (9x faster than Unity out of the box)
- 💾 Less memory than other engines

## Join the community
- [GitHub repository](https://github.com/KaijuEngine/kaiju)
- [Mailing list](https://www.freelists.org/list/kaijuengine) <- Recommended for detailed updates
- [Discord server](https://discord.gg/8rFPEu8U52)
- [Brent Farris on X/Twitter](https://twitter.com/ShieldCrush)

## Why Kaiju?
The current version of the base engine renders extremely fast, faster than most would think a garbage collected language could go. In my testing a release mode build of a game in Unity with nothing but a black background and a cube runs at about 1,600 FPS. In Kaiju, the same thing runs at around 5,400 FPS on the same machine. In fact, a complete game, with audio, custom cursors, real time PBR rendering with real time shadows, UI, and more runs at 2,712 FPS (in "debug" mode) [screenshots or it didn't happen](https://x.com/ShieldCrush/status/1943516032674537958).

## Why Go (golang)?
I love C, and because I love C and found out that Ken Thompson played a part in designing Go, I gave Go a chance. It has been such a joy to use and work with I decided to port my C game engine to Go. Go is a modern system-level language that allows me to write code the way I want to write code and even have the opportunity to do some crazy things if I want to (no strings attached). Also the simplicity and "just works" of writing Assembly code was a great boost to my happiness.

What's more, it's a language that other developers can easily learn and jump right into extending the engine/editor. No need for developers to re-figure out some bespoke macros or crazy templating non-sense. It's flat, easy, straight forward, and the foot-gun is hidden behind some walls, but there if you want it. Furthermore, developers can write their games in Go directly, no need for some alternative language that is different than the engine code (but we'll include Lua for modding).

## What about the Garbage Collector?!
I am creating this section because I get asked about it when I mention "Go", possibly not realizing that most public game engines use a garbage collector (GC).

The GC is actually a feature I'm happy with (shocker coming from a C guy). Well, the reason is simple, if you're going to make a game engine that the public will use and needs to be stable, you need a garbage collector. Unity has C# (and possibly an internal GC as well), Unreal has a GC (and it could use a tune up if you ask me), Godot has a GC albeit their scripting language or when you use C#. It is actually very important for public engines to have a GC because people are only human and make a lot of mistakes, mistakes they'll blame on you (the engine developer) before they blame themselves.

Coincidentally, the overall design I have for the engine, plays very well with the GC and last I measured, I have a net-0 heap allocation while running (may need a new review). If you don't abuse the GC, you shouldn't generally feel it, it runs concurrently as well.

I'll be the first to admit, I think the developers of Go can create a better GC than I can, and probably better than Unreal and Unity too.

## ⚠️ WORK IN PROGRESS ⚠️
Though the engine is production ready, the editor **_is not_**, feel free to join and contribute to it's development.

For the latest updates, please join the [Discord](https://discord.gg/HYj7Dh7ke3) or check my [Twitter/X](https://twitter.com/ShieldCrush).

Please review the ad-hoc [editor readme](https://github.com/KaijuEngine/kaiju/blob/master/src/editor/README.md)

## Compiling the engine
Please see the [documentation](https://kaijuengine.org/engine_developers/build_from_source/) on how to get started and compile the engine

## Editor previews

[full-project-run-cycle.mp4](https://github.com/user-attachments/assets/04c75879-23af-40fa-9773-33cd22cc9552)

[clanker.mp4](https://github.com/user-attachments/assets/6be56b37-589b-4197-86e7-18b1153f7e07)

[working-code-binding.mp4](https://github.com/user-attachments/assets/b7edcbfb-0c78-482f-8eb1-f40910fbaabf)

[content-tagging.mp4](https://github.com/user-attachments/assets/15122db6-efda-4458-bf69-f384def5aa31)

[status-bar-update.mp4](https://github.com/user-attachments/assets/6f3d6511-5db0-405f-b264-af041c199bd0)

[focus-and-transform-hotkeys](https://github.com/user-attachments/assets/95a9bcdc-55fe-4317-9200-412f84a494ce)
