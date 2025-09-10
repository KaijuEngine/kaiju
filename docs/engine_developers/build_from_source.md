---
title: Build from source | Kaiju Engine
---

# Build from source

Below are instructions on how to build the engine from source. Please take care to ensure you're using the Kaiju Engine Go compiler. It is modified for speed (relating to games) and has some features enabled that are currently disabled in Go until the next release.

## Prerequisites
To start, make sure you have the [Vulkan SDK](https://vulkan.lunarg.com/sdk/home) installed for your system.

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

## Issues with SoLoud audio library linking?
If you are having trouble linking with the soloud library (`libs/libsoloud_*.a`), then you'll need to rebuild the library files to link against. It is likely that you're using a different compiler than the original (and it's a C++ library). Below are the instructions on how to build the library. Once built, copy the library `.a` file into the `libs/` folder and overwrite the existing ones.

**Windows**
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud
cd contrib
mkdir build
cd build
cmake .. -G "MinGW Makefiles" .. -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_BACKEND_WASAPI=ON -DSOLOUD_C_API=ON
cmake --build . --config Release
```

**Linux**
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud
cd contrib
mkdir build
cd build
cmake .. -G "Unix Makefiles" -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_BACKEND_ALSA=ON -DSOLOUD_C_API=ON
cmake --build . --config Release
```

## Building content
The source code is not deployed with the project template files generated. So you will want to generate these files before you begin playing around with creating projects. To do this, go into the src folder and run the command below.
```bash
go run ./generators/project_template/main.go
```

This will generate the project template zip file. This zip file is extracted into the folder that you select when creating a new project. It has a copy of the source code and content. Also be sure that whenever you pull new changes in content from the repository, you run this command again to update the project template. This will also require you to re-extract the project template into your project folder.