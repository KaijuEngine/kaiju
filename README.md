# Kaiju Engine
Kaiju is a 2D/3D game engine written in Go (Golang) backed by Vulkan. The goal of the engine is to use a modern, easy, systems level programming language, with a focus on simplicity, to create a new kind of game engine.

## Join the community
- [GitHub repository](https://github.com/KaijuEngine/kaiju)
- [Mailing list](https://www.freelists.org/list/kaijuengine)
- [Discord server](https://discord.gg/8rFPEu8U52)
- [Brent Farris on X/Twitter](https://twitter.com/ShieldCrush)

## Why Kaiju?
The current version of the base engine renders extremely fast, faster than most would think a garbage collected language could go. In my testing a release mode build of a game in Unity with nothing but a black background and a cube runs at about 1,600 FPS. In Kaiju, the same thing runs at around 5,400 FPS on the same machine. In fact, a complete game, with audio, custom cursors, real time PBR rendering with real time shadows, UI, and more runs at 2,712 FPS (in "debug" mode) [screenshots or it didn't happen](https://x.com/ShieldCrush/status/1943516032674537958).

## Why Go (golang)?
I love C, and because I love C and found out that Ken Thompson played a part in designing Go, I gave Go a chance. It has been such a joy to use and work with I decided to port my C game engine to Go. Go is a modern system-level language that allows me to write code the way I want to write code and even have the opportunity to do some crazy things if I want to (no strings attached). Also the simplilcity and "just works" of writing Assembly code was a great boost to my happiness.

What's more, it's a language that other developers can easily learn and jump right into extending the engine/editor. No need for developers to re-figure out some bespoke macros or crazy templating non-sense. It's flat, easy, straight forward, and the foot-gun is hidden behind some walls, but there if you want it. Furthermore, developers can write their games in Go directly, no need for some alternative language that is different than the engine code (but we'll include Lua for modding).

## ⚠️ WORK IN PROGRESS ⚠️
For the latest updates, please join the [Discord](https://discord.gg/HYj7Dh7ke3) or check my [Twitter/X](https://twitter.com/ShieldCrush).

Please review the ad-hoc [editor readme](https://github.com/KaijuEngine/kaiju/blob/master/src/editor/README.md)

## Compiling the engine
Please see the [documentation](https://kaijuengine.org/engine_developers/build_from_source/) on how to get started and compile the engine

## Editor Preview
[monkey-around.webm](https://github.com/user-attachments/assets/fb4ff322-0c5b-49bb-afe8-b2659689618a)
