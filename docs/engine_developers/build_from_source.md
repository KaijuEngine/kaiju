---
title: Build from source | Kaiju Engine
---

# Build from source

Below are instructions on how to build the editor from source

## Prerequisites
To start, make sure you have the [Vulkan SDK](https://vulkan.lunarg.com/sdk/home) installed for your system.

## macOS Development
- Install Xcode Command Line Tools:
  - Run `xcode-select --install` in Terminal
- Install the [Vulkan SDK](https://vulkan.lunarg.com/sdk/home#mac) for macOS
  - After installation, add the Vulkan library path to your environment:
    ```sh
    export VULKAN_SDK=$HOME/Library/VulkanSDK/1.4.XXX.X/macOS  # Replace with your version and your sdk path
    ```
- Pull the repository
- Go into src: `cd src`
- To build the editor in debug mode, run:
  - `CGO_ENABLED=1 CGO_CFLAGS="-I$VULKAN_SDK/include" CGO_LDFLAGS="-L$VULKAN_SDK/lib -lMoltenVK -Wl,-rpath,$VULKAN_SDK/lib" go build -tags="debug,editor" -o ../bin/kaiju`
- To build the editor, run:
  - `CGO_ENABLED=1 CGO_CFLAGS="-I$VULKAN_SDK/include" CGO_LDFLAGS="-L$VULKAN_SDK/lib -lMoltenVK -Wl,-rpath,$VULKAN_SDK/lib" go build -ldflags="-s -w" -tags="editor" -o ../bin/kaiju`

**Note:** On macOS, the engine uses Cocoa/AppKit for windowing and MoltenVK for Vulkan support. Keyboard shortcuts use the Cmd key (not Ctrl) for operations like copy/paste (Cmd+C/Cmd+V).

## Windows Development
### Windows Requirements (Important)

Kaiju requires a **64-bit Go toolchain** on Windows.

Do **not** install the legacy 32-bit `windows-386` version of Go, as it will
fail when compiling the Vulkan backend.

Install the 64-bit distribution: goX.Y.Z.windows-amd64.msi

Download from: https://go.dev/dl/

You can verify your Go architecture by running:
```
go env GOARCH
```
Expected output: amd64

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

### Soloud MacOS
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud/contrib
mkdir build
cd build
cmake .. -G "Unix Makefiles" -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_BACKEND_COREAUDIO=ON -DSOLOUD_C_API=ON -DSOLOUD_STATIC=ON -DCMAKE_POLICY_VERSION_MINIMUM=3.5
cmake --build . --config Release
# Copy the library to kaiju
cp build/libsoloud.a /path/to/kaiju/src/libs/libsoloud_darwin.a
```

### Soloud Android (on Windows)
```sh
git clone https://github.com/jarikomppa/soloud.git
cd soloud
cd contrib
mkdir build
cd build
cmake .. -G "MinGW Makefiles" .. -DCMAKE_TOOLCHAIN_FILE=%NDK_HOME%/build/cmake/android.toolchain.cmake -DANDROID_ABI=arm64-v8a -DANDROID_PLATFORM=android-21 -DANDROID_STL=c++_static -DCMAKE_BUILD_TYPE=Release -DSOLLOUD_STATIC=1 -DSOLLOUD_BUILD_DEMOS=OFF -DSOLOUD_BACKEND_OPENSLES=ON -DSOLOUD_BACKEND_SDL2=OFF -DSOLOUD_C_API=ON -DSOLOUD_BACKEND_NULL=OFF -DSOLOUD_BACKEND_MINIAUDIO=OFF -DSOLOUD_BACKEND_WAVEOUT=OFF -DSOLOUD_BACKEND_XAUDIO2=OFF -DSOLOUD_BACKEND_WINMM=OFF -DSOLOUD_BACKEND_WASAPI=OFF -DSOLOUD_BACKEND_ALSA=OFF -DSOLOUD_BACKEND_COREAUDIO=OFF -DSOLOUD_BACKEND_OPENAL=OFF -DSOLOUD_WAV=ON -DSOLOUD_OGG=ON -DSOLOUD_MP3=ON -DSOLOUD_FLAC=OFF -DSOLOUD_OPUS=OFF -DSOLOUD_SPEECH=OFF -DSOLOUD_SFXR=OFF -DSOLOUD_AY=OFF -DSOLOUD_SID=OFF -DSOLOUD_VIC=OFF -DSOLOUD_TEDSID=OFF -DSOLOUD_MONOTONE=OFF -DSOLOUD_VIC=OFF -DSOLOUD_BASSBOOST=OFF -DSOLOUD_BIQUAD=OFF -DSOLOUD_DCREMOVAL=OFF -DSOLOUD_ECHO=OFF -DSOLOUD_FFT=OFF -DSOLOUD_FREEVERB=OFF -DSOLOUD_LOFI=OFF -DSOLOUD_WAVESHAPER=OFF
cmake --build . --config Release
```

## Building Bullet3
Currently the engine uses Bullet3 for the physics system. Below are instructions
on how to build the library for the engine.

### Bullet3 Windows
```sh
git clone https://github.com/bulletphysics/bullet3.git
cd bullet3
mkdir build_mingw_static
cd build_mingw_static
cmake .. -G "MinGW Makefiles" -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF -DBUILD_CPU_DEMOS=OFF -DBUILD_OPENGL3_DEMOS=OFF -DBUILD_BULLET2_DEMOS=OFF -DBUILD_EXTRAS=OFF -DBUILD_UNIT_TESTS=OFF -DUSE_GLUT=OFF -DBULLET2_MULTITHREADING=ON
mingw32-make -j$(nproc)
```

### Bullet3 Linux
```sh
git clone https://github.com/bulletphysics/bullet3.git
cd bullet3
mkdir build_static
cd build_static
cmake .. -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF -DBUILD_CPU_DEMOS=OFF -DBUILD_OPENGL3_DEMOS=OFF -DBUILD_BULLET2_DEMOS=OFF -DBUILD_EXTRAS=OFF -DBUILD_UNIT_TESTS=OFF -DUSE_GLUT=OFF -DINSTALL_LIBS=ON
make -j$(nproc)
```

### Bullet3 MacOS
```sh
git clone https://github.com/bulletphysics/bullet3.git
cd bullet3
mkdir build_static
cd build_static
cmake .. -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF -DBUILD_CPU_DEMOS=OFF -DBUILD_OPENGL3_DEMOS=OFF -DBUILD_BULLET2_DEMOS=OFF -DBUILD_EXTRAS=OFF -DBUILD_UNIT_TESTS=OFF -DUSE_GLUT=OFF -DINSTALL_LIBS=ON
make -j$(nproc)
```

## Compiling Android
To compile for android, you can go to the engine root folder and run:
```sh
go run src/generators/engine_builds/engine_builds_android/main.go
```
