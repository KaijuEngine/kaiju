# Kaiju Engine
Kaiju Engine is a 2D/3D game engine being developed in the Go programming language. It started as a personal hobby game engine written in C, but we have decided to be re-written in Go.

If you'd like occasional updates on what I'm doing here, I post them [Twitter/X](https://twitter.com/KaijuCoder)

Our discord server is located here: [Discord](https://discord.gg/HYj7Dh7ke3)

## Go language
Being a stubborn C programmer for most of my programming life, it was not easy to make the decision to use Go for the engine, nor was it made lightly. The bridge from C to Go started when I found out it was partly designed by Ken Thompson (partly responsible for C itself), and that it was created out of the frustration of C++. We wish to develop a game engine with the maximum performance and simplicity of design, Go can do this for us. We're not here to get into a language flame war, so just trust us here on this and just have fun!

## WORK IN PROGRESS
This engine is a work in progress. I'm currently porting and refactoring my code from various engines I've written in the past. For an overview of where the engine currently is, please check out the [announcement posts in the GitHub discussions](https://github.com/KaijuEngine/kaiju/discussions).

## Windows Development
- Download mingw into `C:/`
  - I use [x86_64-13.2.0-release-win32-seh-msvcrt-rt_v11-rev1.7z
](https://github.com/niXman/mingw-builds-binaries/releases)
- Add the `bin` folder to your environment path
  - Mine is `C:\mingw64\bin`
- Pull the repository
- Use `build/build.bat` to compile the executable to `bin/kaiju.exe`

### Debug in VSCode
- Open the project in VSCode
- Select one of the Windows debug options
- Press F5

## Linux development
- Ensure you've got `gcc` installed
- Ensure you've got the X11 libs installed (xlib)
- Pull the repository
- Use `build/build.sh` to compile the executable to `bin/kaiju`

### Debug in VSCode
- Open the project in VSCode
- Select one of the X11 debug options
- Press F5
