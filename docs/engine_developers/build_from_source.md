---
title: Build from source | Kaiju Engine
---

# Build from source

Below are instructions on how to build the editor from source

## Prerequisites
To start, make sure you have the [Vulkan SDK](https://vulkan.lunarg.com/sdk/home) installed for your system.

## Windows Development
- Download mingw into `C:/`
  - I use [x86_64-15.2.0-release-win32-seh-msvcrt-rt_v13-rev0.7z](https://github.com/niXman/mingw-builds-binaries/releases)
- Add the `bin` folder to your environment path
  - Mine is `C:\mingw64\bin`
- Pull the repository
- Go into src `cd src`
- To build the exe in debug mode, run:
  - `go build -tags="debug,editor" -o ../kaiju.exe ./`
- To build the exe, run:
  - `go build -ldflags="-s -w" -tags="editor" -o ../kaiju.exe ./`

## Linux development
- Ensure you've got `gcc` installed
- Ensure you've got the X11 libs installed (xlib)
- Pull the repository
- Go into src `cd src`
- To build the exe in debug mode, run:
  - `go build -tags="debug,editor" -o ../kaiju ./`
- To build the exe, run:
  - `go build -ldflags="-s -w" -tags="editor" -o ../kaiju ./`

## Building Soloud
Currently the engine uses Soloud for playing music and sound effects. Below are instructions on how to build the library for the engine.

### Soloud Windows
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud
cd contrib
mkdir build
cd build
cmake .. -G "MinGW Makefiles" .. -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_BACKEND_WASAPI=ON -DSOLOUD_C_API=ON
cmake --build . --config Release
```

### Soloud Linux
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud
cd contrib
mkdir build
cd build
cmake .. -G "Unix Makefiles" -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_BACKEND_ALSA=ON -DSOLOUD_C_API=ON
cmake --build . --config Release
```

### Soloud Android (on Windows)
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud
cd contrib
mkdir build
cd build
cmake .. -G "MinGW Makefiles" .. -DCMAKE_TOOLCHAIN_FILE=%NDK_HOME%/build/cmake/android.toolchain.cmake -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM=android-21 -DANDROID_STL=c++_static -DCMAKE_BUILD_TYPE=Release -DSOLLOUD_OPENSLES=1 -DSOLLOUD_STATIC=1 -DSOLLOUD_BUILD_DEMOS=0 -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_C_API=ON
cmake --build . --config Release
```

## Building content
The source code is not deployed with the project template files generated. So you will want to generate these files before you begin playing around with creating projects. To do this, go into the src folder and run the command below.
```bash
go run ./generators/project_template/main.go
```

This will generate the project template zip file. This zip file is extracted into the folder that you select when creating a new project. It has a copy of the source code and content. Also be sure that whenever you pull new changes in content from the repository, you run this command again to update the project template. This will also require you to re-extract the project template into your project folder.