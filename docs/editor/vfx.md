---
title: VFX Workspace | Kaiju Engine Editor
---

# VFX Workspace
The VFX (Visual Effects) Workspace in the Kaiju Engine Editor is a specialized environment for creating and editing particle systems. Particle systems are used to generate visual effects such as smoke, fire, explosions, rain, and other dynamic graphical elements composed of many small particles.

## Overview
A particle system consists of one or more emitters, each responsible for spawning and controlling particles. The VFX Workspace provides a real-time preview of your particle effects in the stage view, allowing you to see changes immediately as you adjust parameters.

## Accessing the VFX Workspace
To open the VFX Workspace, click on the "VFX" workspace tab in the main menu bar. The workspace consists of three main areas:
- **Stage View**: The central 3D view where particle effects are previewed
- **Left Panel**: Particle system and emitter management
- **Right Panel**: Emitter configuration settings

## Creating a New Particle System
1. Switch to the VFX Workspace.
2. Click on the "New" button on the left to create a new system
3. The new particle system will open automatically, or you can open an existing one by selecting it from the Content Workspace.

## Managing Emitters

### Adding an Emitter
1. In the left panel, click the "Add Emitter" button.
2. A new emitter will be added to the particle system with default settings.
3. Select the emitter to configure its properties in the right panel.

### Selecting an Emitter
- Click on an emitter name in the left panel to select it.
- The right panel will display all configurable properties for that emitter.

### Deleting an Emitter
- Click the "X" button next to an emitter name in the left panel.
- Confirm the deletion when prompted.

## Emitter Configuration
The right panel displays all configurable properties for the selected emitter, organized in collapsible sections.

### Basic Properties
- **Texture**: The texture applied to particles (select from content)
- **Spawn Rate**: How often particles are spawned (particles per second)
- **Particle Life Span**: How long each particle lives (in seconds)
- **Life Span**: How long the emitter runs (0 for infinite)
- **Offset**: Position offset from the emitter's origin

### Movement and Direction
- **Direction Min/Max**: Minimum and maximum initial direction vectors
- **Velocity Min/Max**: Minimum and maximum initial speed ranges

### Appearance
- **Opacity Min/Max**: Initial opacity range for particles
- **Color**: Base color of particles
- **Fade Out Over Life**: Whether particles fade to transparent over their lifetime

### Path Functions
- **Path Function Name**: Select a predefined path (e.g., "Circle", "None")
- **Path Function Offset**: Starting point on the path
- **Path Function Scale**: Size multiplier for the path
- **Path Function Speed**: Speed of movement along the path

### Emission Control
- **Burst**: Whether to emit all particles at once
- **Repeat**: Whether the emitter repeats after its life span

## Previewing Effects
Particle systems are automatically previewed in the stage view. The emitter is positioned at the camera's look-at point, allowing you to see effects in real-time as you adjust parameters.

- Move the camera to change the viewing angle
- Effects update immediately when you change settings
- Use the stage controls to navigate around the effect

## Saving Particle Systems
1. Make sure you have a name entered in the "Particle system" field at the top of the left panel.
2. Click the "Save" button to save your changes.
3. The particle system will be saved as content and can be used in your game.

## Advanced Features

### Custom Path Functions
You can register custom path functions using `vfx.RegisterPathFunc()`. These functions define how particles move over time, allowing for complex behaviors like spirals, waves, or custom patterns.

### Performance Considerations
- Higher spawn rates and longer life spans increase performance requirements
- Use texture atlasing for multiple particle types

## Technical Details
Particle systems are stored as JSON specifications containing an array of emitter configurations. Each emitter manages its own particle pool and rendering data. The system supports instanced rendering for efficient GPU utilization.

Emitters can be configured with various properties to control particle behavior, appearance, and movement patterns. Path functions allow for procedural animation of particle positions over time.

The VFX (Visual Effects) Workspace in the Kaiju Engine Editor is a specialized environment for creating and editing particle systems. Particle systems are used to generate visual effects such as smoke, fire, explosions, rain, and other dynamic graphical elements composed of many small particles.
