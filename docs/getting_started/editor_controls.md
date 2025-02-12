---
title: Editor Controls | Kaiju Engine
---

# Editor Controls
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

## Notes
[1] Parenting selection will parent all selected entities to the last selected entity. If there is only 1 entity selected when parenting, then it will be removed from it's parent and moved to the root.
[2] If a map is currently open, that map will be automatically loaded into in the debug instance that runs.
[3] This will start from the main entry point of the game, it will not load the current map.