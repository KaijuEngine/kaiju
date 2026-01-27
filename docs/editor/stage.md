---
title: Stage Workspace | Kaiju Engine Editor
---

# Stage Workspace
The Stage Workspace is the primary editing environment in the Kaiju Engine Editor, providing a comprehensive interface for creating, manipulating, and organizing 3D and 2D scenes. It serves as the main viewport where developers can visually construct game levels, place entities, and fine-tune their properties in real-time.

## Overview
The Stage Workspace consists of several integrated panels and tools that work together to provide a complete scene editing experience:

- **Stage View**: The central 3D/2D rendering area where entities are displayed and manipulated
- **Hierarchy Panel**: A tree view showing the scene's entity structure
- **Details Panel**: Property editor for selected entities
- **Content Panel**: Asset browser for adding content to the scene
- **Transform Tools**: Gizmos for moving, rotating, and scaling entities
- **Camera Controls**: Navigation tools for viewing the scene from different angles

## Stage View
The Stage View is the heart of the Stage Workspace, displaying your scene in either 3D or 2D mode. You can toggle between camera modes using the dimension toggle button.

### Camera Modes
- **3D Mode**: Full 3D perspective view with turntable-style and fly camera controls
- **2D Mode**: Orthographic X/Y view for 2D game development

### Navigation
- **Mouse Controls**:
  - Alt + Left-click and drag: Rotate camera (3D) or pan (2D)
  - Alt + Right-click and drag: Pan camera
  - Mouse wheel: Zoom in/out
- **Keyboard Controls**:
  - Right-click + W/A/S/D: Move camera (when in fly mode)
  - Right-click + scroll wheel: Increase/decrease camera fly speed

## Hierarchy Panel
Located on the left side, the Hierarchy Panel displays all entities in your scene as a tree structure. This panel allows you to:

- Select entities by clicking on them
- Drag and drop entities to reorder them in the hierarchy
- Search for specific entities using the search bar

### Entity Management
- **Selection**: Click on an entity name to select it
- **Multi-selection**: Hold `Ctrl` to select multiple entities
- **Parenting**: Drag entities onto others to create parent-child relationships

## Details Panel
The Details Panel, positioned on the right side, shows properties of the currently selected entity or entities. It provides:

- **Transform Controls**: Position, rotation, and scale inputs
- **Entity Data**: Custom properties and components
- **Shader Parameters**: Material and shader settings
- **Content References**: Links to assets and resources

## Content Panel
At the bottom of the interface, the Content Panel serves as an asset browser where you can:

- Browse available assets by type and tags
- Search for specific content
- Drag assets directly into the scene
- Preview assets before placement

## Transform Tools
Transform tools allow precise manipulation of selected entities through visual gizmos:

### Translation Tool (Move)
- Red arrow: X-axis movement
- Green arrow: Y-axis movement
- Blue arrow: Z-axis movement

### Rotation Tool
- Red circle: X-axis rotation
- Green circle: Y-axis rotation
- Blue circle: Z-axis rotation

### Scaling Tool
- Red box: X-axis scaling
- Green box: Y-axis scaling
- Blue box: Z-axis scaling

### Tool Usage
- **Activation**: Press 1, 2, or 3 to select translate, rotate, or scale gizmo
- **Snapping**: Hold `Ctrl` for grid snapping during transforms

## Hotkeys

| Key | Action |
|-----|--------|
| `Tab` | Toggle between 3D/2D camera modes |
| `1` | Translation tool |
| `2` | Rotation tool |
| `3` | Scaling tool |
| `G` | Grab (blender style [X, Y, Z lock axis]) |
| `R` | Rotate (blender style [X, Y, Z lock axis]) |
| `S` | Scale (blender style [X, Y, Z lock axis]) |
| `F` | Focus camera on selection |
| `Delete` | Delete selected entities |
| `Ctrl+D` | Duplicate selected entities |
| `Ctrl+Z` | Undo |
| `Ctrl+Y` | Redo |

## First-Time Developer Experience (FTDE)
When opening the Stage Workspace for the first time, you'll see a welcome prompt that introduces you to the basic concepts and provides guidance on getting started with scene creation.
