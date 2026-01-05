---
title: Getting Started | Kaiju Engine
---

# Getting started
TBD

## Installing the editor
Kaiju is a portable program and doesn't require installation at this time. You can either [download a prebuilt version](https://github.com/KaijuEngine/kaiju/tags) or [build from source](/engine/build_from_source/).

## Launching the editor
When you launch the editor, you will be presented with the project select window. In this window, you can either select an existing project from the list below or click the button to find or create a project.

### Selecting a project folder
If you clicked on the "Select project folder" button, a window will pop up, allowing you to browse your file system and select a folder. Navigate into the folder you wish to select, and then click on the "Select" button in the top right.

If the folder is empty, a new project will be created inside of that folder. if there are content inside of the folder, Then the engine will try to determine if it is a kaiju engine project. If it is, it will be opened. If it's not, you will be presented with a warning that the selected folder is not a kaiju project.

You will then be loaded into the main editor window.

### Selecting an existing project
Back on the project select window, there is a list of existing projects if you have previously opened any. By clicking on any of the labels with the project name you're interested in, it will immediately be opened. If that project no longer exists, you will get a warning, and the project will be removed from the list.

You will then be loaded into the main editor window.

### Editor Controls
The primary editor window gives you access to all other editor windows. The main window is also the primary viewport for your game stage (level/map/scene).

| Shortcut       | Description                       |
|----------------|-----------------------------------|
| `Alt + LMB`    | Rotate viewport                   |
| `MMB`	         | Pan viewport                      |
| `Space + LMB`  | Pan viewport                      |
| `Scroll`       | Zoom viewport                     |
| `F`            | Focus the selection               |
| `G`            | Grab/move selection               |
| `R`            | Rotate selection                  |
| `S`            | Scale selection                   |
| `X`            | Locks transform mod to X axis     |
| `Y`            | Locks transform mod to Y axis     |
| `Z`            | Locks transform mod to Z axis     |
| `Y`            | Open content browser              |
| `Z`            | Open content browser              |
| `Ctrl + S`     | Save the current stage            |
| `Ctrl + Space` | Open content browser              |
| `Ctrl + H`     | Open hierarchy window             |
| `Ctrl + P`     | Parent selection [1]              |
| `F5`           | Build and run a debug build [2]   |
| `Ctrl + F5`    | Build and run a release build [3] |

### Notes
[1] Parenting selection will parent all selected entities to the last selected entity. If there is only 1 entity selected when parenting, then it will be removed from it's parent and moved to the root.
[2] If a map is currently open, that map will be automatically loaded into in the debug instance that runs.
[3] This will start from the main entry point of the game, it will not load the current map.
