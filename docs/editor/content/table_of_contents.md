---
title: Table of Contents | Kaiju Engine Editor
---

# Table of Contents
A Table of Contents (TOC) in the Kaiju Engine Editor is a content type that allows you to create and manage collections of content items. It serves as a way to group related assets, entities, or other content for better organization and reference within your project.

## Overview
Each Table of Contents contains a map of entries, where each entry has a unique key (name), an associated content ID, and a display name. This structure enables quick lookup and management of grouped content.

## Creating a Table of Contents
1. In the Content Workspace, select one or more content items you want to include.
2. Right-click on the selected items and choose "Create Table of Contents" from the context menu.
3. Enter a name for the Table of Contents when prompted.
4. The TOC will be created and added to your project's content.

Note: All selected content must have unique names. If duplicates exist, you'll be prompted to resolve them before creation.

## Viewing and Editing a Table of Contents
1. Locate the Table of Contents file in the Content Workspace (it will have a Table of Contents icon).
2. Right-click on it and select "Show Table of Contents" from the context menu.
3. An overlay will appear displaying all entries with their keys and IDs.
4. To remove an entry, click the "X" button next to it.
5. Click outside the overlay to close and save changes.

## Adding Content to an Existing Table of Contents
1. Select the content items you want to add.
2. Right-click on an existing Table of Contents in the Content Workspace.
3. Choose "Add Selected to Table of Contents" from the context menu.
4. The selected items will be added with unique names (duplicates will be suffixed with "_1", etc.).

## Features
- **Unique Naming**: Ensures no duplicate keys within a TOC.
- **Content Reference**: Each entry links to actual content via ID.
- **Editable**: Add or remove entries as needed.
- **Persistent**: Changes are saved to the project's content database.

## Technical Details
Table of Contents are stored as JSON files in the project's content directory. The structure includes:

- `Entries`: A map where keys are strings and values are `TableEntry` objects.
- `TableEntry`: Contains `Id` (content ID) and `Name` (display name).

The TOC system integrates with the project's content database and file system, allowing for efficient serialization and deserialization.
