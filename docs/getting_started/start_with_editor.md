---
title: Getting Started With Editor | Kaiju Engine
description: Guide to getting started with the Kaiju Engine editor, covering installation, project creation, and editor controls.
keywords: Kaiju Engine, editor, getting started, installation, tutorial, game development
---

# Getting started with the editor
The best way to get started with the engine is to get the editor up and running and create a project. I'd highly recommend watching through the Sudoku port series I've created on YouTube to learn the basics on how to use the engine/editor.

## Installing the editor
Kaiju is a portable program and doesn't require installation at this time. You can either [download a prebuilt version](https://github.com/KaijuEngine/kaiju/tags) or [build from source](../engine/build_from_source.md).

## Learn through the Sudoku series
<iframe width="560" height="315" src="https://www.youtube.com/embed/cmjX_M6lEZE?si=ZiCpQOjbgfp_9AV6" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>

## Launching the editor
When you launch the editor, you will be presented with the project select window. In this window, you can either select an existing project from the list of previously opened projects, or create a new project.

### Selecting a project folder
If you clicked on the "Select project folder" button, an overlay will pop up, allowing you to browse your file system and select a folder. Navigate into the folder you wish to select, and then click on the "Select" button in the top right.

If the folder is empty, a new project will be created inside of that folder. if there are content inside of the folder, Then the engine will try to determine if it is a kaiju engine project. If it is, it will be opened. If it's not, you will be presented with a warning that the selected folder is not a kaiju project.

You will then be loaded into the main editor window.

### Selecting an existing project
Back on the project select window, there is a list of existing projects if you have previously opened any. By clicking on any of the labels with the project name you're interested in, it will immediately be opened. If that project no longer exists, you will get a warning, and the project will be removed from the list.

You will then be loaded into the main editor window.

## Special terms
**Stage** - A collection of entities that are to be loaded, others may call it a "map", "scene", "level", etc. Stages help you build out your map in "stages", they can be merged together at runtime. *The term "Stage" is also a throw-back to what we would call maps/levels for games in the 90s*

**Template** - A singular entity, it's transform, shader data, any entity data attached to it, and all of the child entities likewise. In other environments people would call these "prefabs" or "blueprints". When a template is updated, all usages of the template across the game are updated as well.

**Table of Contents** - A collection of content ids that are grouped together for easy referencing. You can use a friendly `string` name to access various content found in the table at runtime. This can help reduce the need to have `const` string ids to content in your game code.

## Editor Controls
The primary editor window gives you access to all other editor windows. The main window is also the primary viewport for your game stage (level/map/scene).

| Shortcut       | Description                                |
|----------------|--------------------------------------------|
| `Alt + LMB`    | Rotate viewport                            |
| `MMB`	         | Pan viewport                               |
| `Space + LMB`  | Pan viewport                               |
| `Alt + RMB`    | Zoom viewport                              |
| `Scroll`       | Zoom viewport                              |
| `F`            | Focus the selection                        |
| `G`            | Grab/move selection                        |
| `R`            | Rotate selection                           |
| `S`            | Scale selection                            |
| `X`            | Locks transform mod to X axis              |
| `Y`            | Locks transform mod to Y axis              |
| `Z`            | Locks transform mod to Z axis              |
| `C`            | Toggle content panel                       |
| `H`            | Toggle hierarchy panel                     |
| `D`            | Toggle details panel                       |
| `Ctrl + S`     | Save the current stage                     |
| `Ctrl + T`     | Create template from selected              |
| `Ctrl + P`     | Parent selection [1](#note_1)              |
| `F5`           | Build and run a debug build [2](#note_2)   |
| `Ctrl + F5`    | Build and run a release build [3](#note_3) |

### Notes
<a id="note_1"></a>
[1] Parenting selection will parent all selected entities to the last selected entity. If there is only 1 entity selected when parenting, then it will be removed from it's parent and moved to the root.

<a id="note_2"></a>
[2] If a stage is currently open, that stage will be automatically loaded into by the debug instance that runs.

<a id="note_3"></a>
[3] This will start from the main entry point of the game, it will not load the current stage.
