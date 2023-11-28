# Kaiju Engine
Kaiju Engine is a 2D/3D game engine being developed in the Go programming language. It started as a personal hobby game engine written in C, but we have decided to be re-written in Go.

## Go language
Being a stubborn C programmer for most of my programming life, it was not easy to make the decision to use Go for the engine, nor was it made lightly. The bridge from C to Go started when I found out it was partly designed by Ken Thompson (partly responsible for C itself), and that it was created out of the frustration of C++. We wish to develop a game engine with the maximum performance and simplicity of design, Go can do this for us. We're not here to get into a language flame war, so just trust us here on this and just have fun!

## Video History
[View the playlist](https://www.youtube.com/playlist?list=PLwZ7-gKDdxn4MdyH6-t0It1lGUOAJ0aKz)

The development of Kaiju is to be captured on video. Every PR will be required to have a video of the development process of that request. The goal of this is to be a tool for new developers to learn the process of developing features for a game engine, but also help maintainers to understand the thinking behind the code. It is often hard to know why choices were made just by looking at the final product of the code in the review, having a video log/history of the development process helps us all learn together.

We are well aware that requiring a video log for every PR is something that will put off developers, but we strongly believe that code is not just some assembly line to pump out a result, it's a beautiful tapestry of ideas, puzzles, and hard-made choices. We wish to share with everyone in the process and to learn from others as well.

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