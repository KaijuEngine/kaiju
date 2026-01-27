---
title: Content Workspace | Kaiju Engine Editor
---

# Content Workspace
The Content Workspace is a core component of the Kaiju Engine Editor that allows developers to manage, organize, and preview project assets and content. It provides a user-friendly interface for importing, filtering, tagging, and manipulating various types of content such as textures, models, audio files, shaders, and more.

## Overview
The Content Workspace features a three-panel layout designed for efficient content management:

### Left Panel: Filters and Import
- **Content Search**: A text input field for searching content by name.
- **Import Content**: Button to open a file browser for importing new assets into the project.
- **List View Toggle**: Switch between grid and list view modes for content display.
- **Type Filters**: Buttons to filter content by type (e.g., Texture, Model, Audio).
- **Tag Filters**: Dynamic list of tags for filtering content by user-defined tags.

### Center Panel: Content List
- Displays imported content in a scrollable grid or list.
- Each entry shows a preview image, name, and type icon.
- Supports single and multi-selection of content items.
- Right-click context menu for additional actions.

### Right Panel: Content Details
- **Name Editor**: Input field to rename selected content.
- **Tags Management**: 
  - List of current tags with remove buttons.
  - Input field to add new tags with auto-completion hints.
- **Audio Player** (for audio content): Play/pause controls, time display, and seek slider.
- **Action Buttons**:
  - **Open in Editor**: Launch the appropriate editor for the content type.
  - **Reload**: Reimport the content from its source file.
  - **Delete**: Remove the content from the project.

## Key Features

### Importing Content
1. Click the "Import content..." button in the left panel.
2. Select one or more files from the file browser.
3. The editor will process the files, create appropriate configurations, and add them to the content database.
4. Imported content appears in the center panel and is ready for use in the project.

Supported content types include:
- Textures (images)
- 3D Models
- Audio files (music and sound effects)
- Shaders
- Fonts
- And more...

### Filtering and Searching
- **Text Search**: Type in the search box to filter by content name (case-insensitive partial matches).
- **Type Filtering**: Click type buttons to show/hide content of specific types.
- **Tag Filtering**: Click tag buttons to filter content by tags.
- Filters can be combined for precise content discovery.

### Content Selection
- Click content entries to select them (single selection).
- Hold Ctrl/Cmd to add/remove from selection (multi-selection).
- Selected content shows in the right panel for editing.

### Editing Content Properties
- **Renaming**: Edit the name in the right panel's input field.
- **Tagging**: 
  - Add tags by typing in the "Add new tag..." field.
  - Auto-completion suggests existing tags.
  - Remove tags by clicking the X button next to each tag.

### Audio Content Playback
For audio content (music and sound effects):
- Audio player appears in the right panel when audio content is selected.
- Play/Stop button to control playback.
- Time display shows current position and total duration.
- Seek slider to jump to different parts of the audio.

### Table of Contents
The Content Workspace supports creating and managing "Table of Contents" for organizing related content:

- Select multiple content items.
- Use context menu or keyboard shortcut to create a new Table of Contents.
- Tables of Contents are special content types that group related assets.
- View and edit Table of Contents contents through the overlay interface.

### Deleting Content
- Select content to delete.
- Click the "Delete" button in the right panel.
- Confirm deletion in the prompt (note: this may affect references in the project).

## Technical Details

### Content Database
All content is managed through the Content Database system:
- **Import Process**: Files are processed and converted to engine-compatible formats.
- **Caching**: Content metadata is cached for fast access.
- **Dependencies**: The system tracks content relationships and dependencies.
- **Reimporting**: Source files can be reimported to update content.

### File Structure
Imported content creates several files:
- **Config File**: JSON configuration with metadata, tags, and settings.
- **Content File**: Processed asset data in engine format.
- **Source Link**: Reference to original source file for reimporting.

### Integration with Editor
The Content Workspace integrates with other editor systems:
- **Asset Database**: Provides content to the runtime engine.
- **Scene Editor**: Content can be dragged into scenes.
- **Property Editors**: Content references appear in component properties.
- **Build System**: Content is included in project builds.

## Best Practices
1. **Organize with Tags**: Use consistent tagging to categorize content.
2. **Use Descriptive Names**: Give content meaningful names for easy identification.
3. **Regular Cleanup**: Remove unused content to keep the project tidy.
4. **Version Control**: Commit content changes along with source files.
5. **Backup Sources**: Keep original source files safe for reimporting.
