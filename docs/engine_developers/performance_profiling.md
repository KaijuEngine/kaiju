---
title: Performance Profiling | Kaiju Engine
---

The engine currently uses [Gotraceui](https://gotraceui.dev/) for visualizing the pprof traces that are built into Go. There is some helper code to automatically trace the running engine and launch GotraceUI upon closing the game window.

Note that `debug` version of the editor (especially with tracing) is about 2x (or more) slower than the release runtime.

## Install Gotraceui
First, you'll need to install the Gotraceui application. This may mean that you need to install from the `master` branch to get the correct Go version support.

## Launch with tracing
If you are using VSCode, you can select the `Trace Editor` option from the debug options drop down list.

<img width="235" height="228" alt="image" src="https://github.com/user-attachments/assets/f78887a4-3e43-48c3-8d03-f7b263abdac8" />

Otherwise, you will need to use the `-trace` command line arg in a `debug` build. To get a `debug` build, you need to compile the editor/game using `-tags="debug,..."` command line arg in your call to `go build`.

This will begin tracing your game the moment the main starts up and end the moment you close the window. This creates a `trace.out` file in your current working directory that will be automatically passed to Gotraceui. You can supply this file to other tools if you prefer to analyze it in another way.

## Code instrumentation
Throughout the code, you will notice a call to `defer tracing.NewRegion("...").End()` at the top of functions. This is used to implement tracing of the game to show up in the trace file. This code is automatically stripped in shipping builds by the Go compiler. You will also notice that some functions don't have this markup, this is typically because the function is so very small and likely will be inlined and it doesn't call any other functions. If you call other functions, you typically want to add a trace to make the chain of events show up in the trace viewer.

_Note that `"..."` is typically in the format `package_name.FunctionName` if it is a package function, or `StructName.FunctionName` if it is a struct function._

## Gotraceui basics
Gotraceui's website has pretty good [documentation](https://gotraceui.dev/manual/latest/) on how to use the tool and that might be a good place to start. However, the bare minimum gist is to scroll down until you see the pink blocks in the timeline area.

<img width="438" height="443" alt="image" src="https://github.com/user-attachments/assets/eeb360f2-34dd-4222-80e6-6e512469effc" />

Here you can zoom in and see the stacks of function calls that the engine is doing. This will allow you to review any performance bottlenecks or unexpected calls.

_Note that you can scroll further down and see other pink boxes for the work being done on other threads._

<img width="787" height="698" alt="image" src="https://github.com/user-attachments/assets/8509d3d2-c596-4fcc-9dc3-92258fc45207" />
