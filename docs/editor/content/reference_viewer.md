---
title: Reference Viewer | Kaiju Engine Editor
---

# Reference Viewer
The Reference Viewer is an overlay in the Kaiju Engine Editor that displays all references to a specific content item within the project. This helps developers understand where assets, entities, or other content are being used, aiding in refactoring or deletion decisions.

## Accessing the Reference Viewer
The Reference Viewer can be opened by right-clicking on content in the Content Workspace and selecting "Show References" from the context menu. It can also be triggered programmatically via the editor's API using `editor.ShowReferences(id)`.

## Features
- **Search Functionality**: Automatically searches through stages, templates, materials, shaders, HTML, CSS, and code files for references to the selected content ID.
- **Hierarchical Display**: References are displayed in a tree structure, showing parent references and their sub-references.
- **Real-time Updates**: As references are found, they are added to the list in real-time.
- **Source Information**: Each reference shows the name and source type (e.g., "entity", "material").

## Usage
1. Select a content item in the Content Workspace.
2. Right-click and choose "Show References".
3. The overlay will appear, showing "Searching..." while it scans the project.
4. Once complete, the list of references will be displayed.
5. Click outside the overlay to close it.

## Technical Details
The Reference Viewer uses the project's reference finding system, which scans various file types and structures to locate usages of the content ID. This includes:

- Stage files (.stage)
- Template files
- Material definitions
- Shader files
- Table of Contents
- HTML and CSS files
- Source code files

References are found by matching the content ID in relevant fields and structures within these files.
