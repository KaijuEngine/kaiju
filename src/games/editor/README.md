# Kaiju Engine Editor
The engine runtime and editor have the same goal of being as fast and responsive
as possible. In the initial iteration of the editor I was simply adding things
in as I thought of them. This created a non-cohesive code-base and also UI/UX.
The goal of this new version of the editor is to be cohesive and having each
part thought out before writing any code.

*Note: "the developer" is used to describe the person interacting with the
editor who is actively developing a game/application.*

## Go code design
Every attempt is made to make the code as performant as possible and generate
as little of memory garbage as possible. All new code must be thoroughly planned
and designed before being written. This can be via technical design doc,
flow charts, and/or any other type of specification document. All public
functions must have clean, readable, thorough, and expressive comments to
describe the intent. Please review the
[CONTRIBUTING](https://github.com/KaijuEngine/kaiju/blob/master/CONTRIBUTING.md)
document for coding rules.

## Window design
Having floating windows is at times useful, but they create a very clunky way
for the developer to interact with the editor. For this reason, no external
popup windows will be permitted. Virtual overlays like confirmations, progress
bars, and other "obstructive" elements will be presented within the main window.

## Workspaces
Developers will be presented with "workspaces" for the different tasks that
they are focused on. For example, an "Animation", or "Stage", or "Content" work
space that the developer can focus in on. Having custimizable UI is nice, and
potentially a task for the future, but it greatly distracts from focusing on
creating great engine/editor features and dramatically bloats the code.

### Project workspace
The project workspace allows the developer to manage existing projects or
create a new project. This workspace will have a list of projects that the
editor is aware of in the center/left of the view. Selecting a project will open
the details about the project in the details panel on the right. The developer
will also have the option to delete the workspace from the details panel. At the
top of the workspace will be a button that will allow the developer to create a
new project. Clicking this button will make the file browser overlay present
itself. Upon selecting a new folder location, the project will be created and
the stage workspace will be presented.

### Stage workspace
The stage workspace presents developers with an asset browser, a hierarchy, and
a details panel. When selecting an entity in the hierarchy or within the stage
viewport, it will present that entity's information in the details panel.

#### Stage - Asset browser
The asset browser allows developers to search, filter, and drag-drop items from
within onto the stage, hierarchy, or details panel. The asset browser will be
positioned along the bottom of the window.

#### Stage - Hierarchy
The hierarchy shows all of the entities within a stage with their parent/child
relationship apparent. It also allows developers to select entities directly
within the hierarchy, rather than needing to navigate to it on the stage.
Entites can be parented, re-parented, selected, and deleted from within the
hierarchy. The hierarchy will be positioned along the right of the window.

#### Stage - Details panel
The details panel shows information about the selected entity or the selected
asset. Entities will present information related to their transformation and any
components assigned to them. Assets will show information about their current
configuration, the "compression" of a texture, for example. The details panel
will be positioned along the right of the window.

### Content workspace
The content workspace allows developers to manage the various content in their
game. It consists of a single large content search/preview area on the left of
the screen, and a smaller details window on the right of the screen. The
developer will be presented with a button to import new content on on the larger
left panel of the screen. Content can also be deleted by selecting on the assets
and clicking on the "Delete" button at the bottom of the details window. When
selecting an asset, it's details will be shown to modify things like categories,
tags, compression, etc.

## Content
Assets will automatically be placed into folders matching their asset type. So,
textures will go into textures, materials into materials, stages into stages,
etc. Developers will not create sub-folders or deal with any assets directly
on the file system. Instead, they can use categories and tags to create virtual
folders that can be filtered/searched. This prevents assets from moving around
and for the editor to keep track of the linkage of that asset location to it's
references. Content should be sortable by import date to make it easier to find
newly imported assets.

Likewise, developers can not change the name of any of the asset files, they
will have a GUID name assigned on import that should never be touched. Instead
the asset can have a virtual name associated with it, stored the same way that
category and tag assignments are. Upon import, the asset will take on a virtual
name equal to the file name that was imported. Content will be imported into the
`database/content` folder.

## Database
There will be various information that needs to be stored about content and the
developer's project. This information will be stored locally both to the file
as well as in a cache [SQLite](https://www.sqlite.org/) database.

### Database - Asset configuration
Developer-assigned information like name, category, tags, etc. Will be stored
into a compressed JSON file format with the same GUID as the target asset. This
file will have the extension `.json` and reside in the `database/configuration`
folder matching the `database/content` folder structure. These files are to be
committed to version control as they can be used to build the database cache.

### Database - Cache (SQLite)
The cache database is to speed up the process of search by mirroring all of the
data found in the asset configuration files. It will also store any runtime 
information to speed up the interface and usability, by storing things like the
BVH structures for assets for example. This file is not to be committed to
version control as it's a binary and can be re-constructed by scanning all of
the asset configuration files as well as the assets to rebuild the cache. The
cache database will be stored in the `database/cache.db` file.

### Database - Folder layout
- root
	- database
		- cache.db
		- config
			- * (matches content structure)
		- content
			- audio
				- music
				- sound
			- font
			- mesh
			- ui
				- html
				- css
			- render
				- material
				- spv
			- texture
		- src
			- font
				- charset
			- plugin (editor extensions)
			- render
				- shader (raw shader source code)

## Project
The "project" referrs to the game/application that the developer is using the
editor to create. The editor can not be used without first selecting a project.
This means that when the editor starts up, it should either be loading an
existing project, either from last time, the command line, or the developer
opening the project directly from their file browser. If the editor is otherwise
started without such a project, the developer will be prompted with the
"project" workspace. This workspace can not be exited until a project is
either selected or created.

## Overlays
Some UI views do not fit neatly within the system as a "workspace" and therefore
are labeled as "overlays". These overlays are used to select or present
contextual information about the action of the developer or of the editor. Every
overlay should block input to the rest of the editor while they are presented.

### File browser overlay
The file browser overlay allows the developer to select a file, a folder, or
multiplies of them. This overlay will take up the majority of the screen and
have a panel on the left for quick access to common locations and a center panel
with a path input bar and a list of files and folders within that path.

*Note: In the future we'll add a search input bar to the top next to the path
input box to make it easier for the developer to search the current folder for
files.*

### Confirm overlay
The confirm overlay is a simple overlay that presents 2 options, typically
"Okay" or "Cancel". The overlay has a title and a description as well. The
title, description, confirm, and cancel texts should all be settable upon
invoking the overlay.

### Input overlay
The input overlay allows the developer to input a string into an input box and
submit it to the invoker. The overlay has a title and a description as well. The
title, description, input placeholder text, and input default text texts should
all be settable upon invoking the overlay.

### Progress bar overlay
The progress bar overlay is used to present the developer with information on
the current progress of an action. This overlay contains a progress bar across
the center as well as a label at the bottom to show the status of what is being
worked through. Optionally this overlay can include a title and description that
can be set by the invoker to describe the action being processed.
