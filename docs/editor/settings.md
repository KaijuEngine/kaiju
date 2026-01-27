---
title: Settings Workspace | Kaiju Engine Editor
---

# Settings Workspace
The Settings Workspace in the Kaiju Engine Editor provides a centralized interface for configuring both project-specific and editor-wide settings. It allows you to customize the behavior of the editor, set up external tools, manage plugins, and configure project properties.

## Accessing the Settings Workspace
To open the Settings Workspace, click on the "Settings" workspace in the main menu bar. The workspace is divided into three main sections, accessible via the left panel: Project Settings, Editor Settings, and Plugin Settings.

## Project Settings
Project Settings are specific to the current project and affect how the project is built, run, and configured.

### General Settings
- **Name**: The display name of the project.
- **Entry Point Stage**: The initial stage file to load when running the project.
- **Archive Encryption Key**: Optional key for encrypting archived builds.

### Android Settings
- **Root Project Name**: The name used for the Android project root.
- **Application Id**: The unique identifier for the Android application (e.g., com.example.myapp).

## Editor Settings
Editor Settings control the behavior and appearance of the Kaiju Engine Editor itself. These settings are global and persist across projects.

### Display
- **Refresh Rate**: Target frame rate for the editor (default: 60 FPS).
- **UI Scroll Speed**: Speed of scrolling in UI elements (default: 20).

### External Editors
Configure paths to external applications for editing different file types:

- **Code Editor**: Path to your preferred code editor (default: "code" for VS Code).
- **Image Editor**: Path to image editing software.
- **Mesh Editor**: Path to 3D mesh editing software.
- **Audio Editor**: Path to audio editing software.

### Camera
Settings for the editor's camera controls:

- **Zoom Speed**: Speed of zooming in/out (default: 120).
- **Fly Speed**: Movement speed when flying (default: 10).
- **Fly X/Y Sensitivity**: Mouse sensitivity for camera rotation (default: 0.2).

### Snapping
Grid snapping increments for precise placement:

- **Translate Increment**: Snapping distance for position changes.
- **Rotate Increment**: Snapping angle for rotations.
- **Scale Increment**: Snapping factor for scaling.

### Build Tools
Paths to required build tools:

- **Android NDK**: Path to Android Native Development Kit (auto-detected if possible).
- **Java Home**: Path to Java installation (auto-detected from JAVA_HOME or common locations).

## Plugin Settings
The Plugin Settings section allows you to manage editor plugins.

### Managing Plugins
- **Open Plugins Folder**: Button to open the plugins directory in your file browser.
- **Recompile Editor**: Button to recompile the editor with current plugin settings (only needed when plugin enable/disable state changes).

### Plugin List
Each available plugin is displayed with:

- **Name**: Plugin name.
- **Description**: Brief description of the plugin's functionality.
- **Author**: Plugin author.
- **Version**: Plugin version.
- **Website**: Link to plugin website (if provided).
- **Enabled**: Checkbox to enable or disable the plugin.

## Saving Changes
Settings are automatically saved when you close the Settings Workspace. For plugin changes that require recompilation, use the "Recompile Editor" button after toggling plugin states.

## Technical Notes
- Editor settings are stored in a JSON file in the user's application data directory.
- Project settings are stored in the project's config file.
- Some settings have automatic detection (e.g., Android NDK and Java Home paths).
- Plugin changes require editor recompilation to take effect.
