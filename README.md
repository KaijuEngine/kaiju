# Kaiju Engine
Kaiju Engine is a 2D/3D game engine being developed in the Go programming language. It started as a personal hobby game engine written in C, but we have decided to be re-written in Go.

If you'd like occasional updates on what I'm doing here, I post them [Twitter/X](https://twitter.com/KaijuCoder)

Our discord server is located here: [Discord](https://discord.gg/HYj7Dh7ke3)

## ⚠️ WORK IN PROGRESS ⚠️
For the latest updates, please join the [Discord](https://discord.gg/HYj7Dh7ke3) or check my [Twitter/X](https://twitter.com/KaijuCoder).

# Developing from source
Below are instructions on how to build the engine from source. Please take care to ensure you're using the Kaiju Engine Go compiler. It is modified for speed (relating to games) and has some features enabled that are currently disabled in Go until the next release.

## Prerequisites
I have made modifications to the Go complier to increase the performance of the engine, for this reason you'll need to build the engine with the Kaiju Engine Go compiler
- Download the [Kaiju Engine Go compiler](https://github.com/KaijuEngine/go/tree/kaiju-go1.22) (release version 1.22)
  - This should be placed along side the Kaiju Engine repository
- Ensure you have the standard Go compiler installed (Go builds Go)
- Run the make script file inside of the `src` directory
  - This will build the Kaiju Engine Go compiler into the `bin` directory

## Windows Development
- Download mingw into `C:/`
  - I use [x86_64-13.2.0-release-win32-seh-msvcrt-rt_v11-rev1.7z
](https://github.com/niXman/mingw-builds-binaries/releases)
- Add the `bin` folder to your environment path
  - Mine is `C:\mingw64\bin`
- Pull the repository
- To build the exe, run `go run build/build.go`
  - Make sure to use the Kaiju Engine Go compiler

### Debug in VSCode
- Open the project in VSCode
- Press Ctrl+Shift+P and type "Choose Go Environment"
  - Select the Kaiju Engine Go compiler `bin` folder
- Select one of the Windows debug options
- Press F5

## Linux development
- Ensure you've got `gcc` installed
- Ensure you've got the X11 libs installed (xlib)
- Pull the repository
- To build the exe, run `go run build/build.go`
  - Make sure to use the Kaiju Engine Go compiler

### Debug in VSCode
- Open the project in VSCode
- Press Ctrl+Shift+P and type "Choose Go Environment"
  - Select the Kaiju Engine Go compiler `bin` folder
- Select one of the X11 debug options
- Press F5
