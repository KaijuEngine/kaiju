---
title: UI Workspace | Kaiju Engine Editor
description: Documentation for the UI Workspace in the Kaiju Engine Editor, covering features, usage, and troubleshooting.
keywords: UI Workspace, Kaiju Engine, editor, UI design, live preview, HTML, CSS, data binding, responsive design
---

# UI Workspace
The UI Workspace in the Kaiju Engine Editor is a specialized environment for designing, previewing, and testing user interface layouts. It provides a live preview of HTML-based UI files, allowing developers to see changes in real-time as they edit the markup and styles.

## Overview
The UI Workspace bridges the gap between static HTML editing and runtime UI behavior. It supports dynamic data binding, CSS styling, and responsive design testing through aspect ratio controls. This workspace is essential for creating polished user interfaces for games and applications built with Kaiju Engine.

## Accessing the UI Workspace
To open the UI Workspace:

1. Switch to the Content Workspace.
2. Locate an HTML file in your project's content.
3. Right-click on the HTML file and select "View in UI workspace" from the context menu.

Alternatively, you can directly switch to the UI workspace via the workspace selector and then load an HTML file.

## Interface Overview
The UI Workspace consists of:

- **Preview Area**: The central area displaying the live UI preview.
- **Help Panel**: Initial instructions when no file is loaded.
- **Menu Bar**: Controls at the bottom for editing and configuration.

## Loading UI Files

### From Content Workspace
1. Navigate to your HTML content file.
2. Right-click and choose "View in UI workspace".
3. The file will automatically load in the UI Workspace.

### Manual Loading
If needed, you can load HTML files directly through the workspace interface (though the primary method is through the Content Workspace).

## Preview Features

### Live Reloading
The UI Workspace automatically detects changes to your HTML and CSS files:

- Edit your HTML file in an external editor.
- Save the file.
- The preview updates within 1 second to reflect your changes.

### Dynamic Data Binding
Load mock data to test dynamic content:

1. Click the "Load mock data" button in the menu bar.
2. Select a JSON file containing sample data.
3. The UI will re-render with the bound data.

Mock data files should be JSON objects that match the structure expected by your HTML templates.

### Aspect Ratio Control
Test responsive design with different screen ratios:

- **Width Ratio**: Set the target width ratio (e.g., 16 for 16:9).
- **Height Ratio**: Set the target height ratio (e.g., 9 for 16:9).
- The preview area scales to fit the specified aspect ratio within the workspace.

Leave both fields at 0 for full-window preview.

## Editing UI Files

### Opening in Editor
1. Click the "Edit HTML" button in the menu bar.
2. Your configured code editor (default: VS Code) will open with the HTML file loaded.
3. Make changes and save - the preview will update automatically.

### External Editor Configuration
Configure your preferred editor in the Settings Workspace under Editor Settings > External Editors > Code Editor.

## Best Practices

### File Organization
- Keep HTML, CSS, and mock data files organized in your project's content structure.
- Use relative paths for CSS and asset references.

### Data Binding
- Create comprehensive mock data files that cover all possible UI states.
- Test with various data sizes to ensure performance.

### Responsive Design
- Regularly test different aspect ratios to ensure your UI works across devices.
- Use CSS media queries for responsive layouts.

## Troubleshooting

### Preview Not Updating
- Ensure your HTML and CSS files are saved.
- Check that file paths are correct and files exist.
- Verify that the editor has write access to the files.

### Data Binding Issues
- Confirm your JSON mock data matches the expected structure.
- Check for JSON syntax errors.
- Ensure data property names match your HTML template bindings.

### Aspect Ratio Problems
- Enter positive numbers for both width and height ratios.
- If the preview appears too small, check your window size and zoom level.

## Technical Details
The UI Workspace uses Kaiju Engine's markup system to render HTML with:

- **Document Parsing**: HTML is parsed and converted to engine UI elements.
- **CSS Support**: Cascading stylesheets are applied for visual styling.
- **Data Binding**: JSON data is bound to HTML elements using template syntax.
- **Live Monitoring**: File system watchers detect changes for automatic reloading.

UI files are stored as standard HTML in the project's content directory and can be loaded at runtime using the engine's asset system.
