# Kaiju Engine
Kaiju is a 2D/3D game engine written in Go (Golang) backed by Vulkan. The goal of the engine is to use a modern, easy, systems level programming language, with a focus on simplicity, to create a new kind of game engine.

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

## ⚠️ WORK IN PROGRESS ⚠️
For the latest updates, please join the [Discord](https://discord.gg/HYj7Dh7ke3) or check my [Twitter/X](https://twitter.com/ShieldCrush).

Please review the ad-hoc [editor readme](https://github.com/KaijuEngine/kaiju/blob/master/src/editor/README.md)

## Compiling the engine
Please see the [documentation](https://kaijuengine.org/engine_developers/build_from_source/) on how to get started and compile the engine

## Editor previews
[complete-editor-refactor-progress.mp4](https://github.com/user-attachments/assets/00291482-1624-4cfb-b59f-c0395a5f1075)

[wip-file-browser-overlay.mp4](https://github.com/user-attachments/assets/ba4f4049-ded5-4638-8bdd-c09daf7869ce)

[updated-content-window.mp4](https://github.com/user-attachments/assets/1d1d773a-6a7e-4986-b984-e08fb303617c)

[content-tagging.mp4](https://github.com/user-attachments/assets/15122db6-efda-4458-bf69-f384def5aa31)

[status-bar-update.mp4](https://github.com/user-attachments/assets/6f3d6511-5db0-405f-b264-af041c199bd0)

[status-bar-update.mp4](https://github.com/user-attachments/assets/09ee5e96-6c47-47d5-9e66-6e92888562ce)
