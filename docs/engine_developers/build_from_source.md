---
title: Kaiju Engine | Build from source
---

# Build from source

Below are instructions on how to build the engine from source. Please take care to ensure you're using the Kaiju Engine Go compiler. It is modified for speed (relating to games) and has some features enabled that are currently disabled in Go until the next release.

## Prerequisites
To start, make sure you have the [Vulkan SDK](https://vulkan.lunarg.com/sdk/home) installed for your system.

I have made modifications to the Go compiler to increase the performance of the engine, for this reason you'll need to build the engine with the Kaiju Engine Go compiler

- Download the [Kaiju Engine Go compiler](https://github.com/KaijuEngine/go/tree/kaiju-go1.22) (release version 1.22)
  - This should be placed along side the Kaiju Engine repository
- Checkout kaiju-go1.22, run `git checkout kaiju-go1.22`, inside of the repository folder
- Ensure you have the standard Go compiler installed (Go builds Go)
- Run the `make` script file for your platform inside of the `src` directory
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

## Linux development
- Ensure you've got `gcc` installed
- Ensure you've got the X11 libs installed (xlib)
- Pull the repository
- To build the exe, run `go run build/build.go`
  - Make sure to use the Kaiju Engine Go compiler

## Debug in VSCode
- Open the project in VSCode
- Press Ctrl+Shift+P and type "Choose Go Environment"
  - Select the Kaiju Engine Go compiler `bin` folder
- Select one of the debug options
- Press F5

## Building content
The source code is not deployed with the project template files generated. So you will want to generate these files before you begin playing around with creating projects. To do this, go into the src folder and run the command below.
```bash
go run ./generators/project_template/main.go
```

This will generate the project template zip file. This zip file is extracted into the folder that you select when creating a new project. It has a copy of the source code and content. Also be sure that whenever you pull new changes in content from the repository, you run this command again to update the project template. This will also require you to re-extract the project template into your project folder.