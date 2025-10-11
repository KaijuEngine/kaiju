# Kaiju Engine Editor
The engine runtime and editor have the same goal of being as fast and responsive
as possible. In the initial iteration of the editor I was simply adding things
in as I thought of them. This created a non-cohesive code-base and also UI/UX.
The goal of this new version of the editor is to be cohesive and having each
part thought out before writing any code.

## Window design
Having floating windows is at times useful, but they create a very clunky way
for the developer to interact with the editor. For this reason, no external
popup windows will be permitted. Virtual overlays like confirmations, progress
bars, and other "obstructive" elements will be presented within the main window.

## Work spaces
Developers will be presented with "work spaces" for the different tasks that
they are focused on. For example, an "Animation", or "Stage", or "Content" work
space that the developer can focus in on. Having custimizable UI is nice, and
potentially a task for the future, but it greatly distracts from focusing on
creating great engine/editor features and dramatically bloats the code.

### Stage work space
The stage work space presents developers with an asset browser, a hierarchy, and
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

### Content work space
The content work space allows developers to manage the various content in their
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

## Project folder layout
- root
	- database
		- cache.db
		- configuration
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
				- src
			- texture
		- src
			- font (pre-msdf font files)
			- plugin (editor extensions)
			- render
				- shaders (raw shader source code)
