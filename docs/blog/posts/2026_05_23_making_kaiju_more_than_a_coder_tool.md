---
title: Making Kaiju More Than A Coder Tool
description: Adding tools to enable artists and make level designing faster, fun, and more fluid.
tags: game engine, level design, tools
category: Editor
date: 2026-05-23
---

# Making Kaiju More Than A Coder Tool
<div class="blog-author">
    <span class="author-name">
		Brent Farris
	</span>
    <span class="author-date">
		May 23rd, 2026
	</span>
</div>

---

## Improving engine usability
For a long time I've been heavily focused on code, performance, and what the experience of the engine is to programmers/developers. The last few months I've decided that I want to improve the tools not only for my own use, but also for artists and designers too. This means adding quality of life things, making the UI a little more usable, creating tools for artists, preparing for localization, etc.

## UI Updates
![New project UI](https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/new_project.png)

First impressions are important and having something nice to look at when you initially open the engine becomes more important over time. To achieve this new look, there are several updates to the UI shader to improve outline quality, precision, border shape/size, and other finer details. Recent updates have pushed our CSS implementation much further than it had been originally. Recent updates also bring the new `<kaiju-include>` tag, as seen in `stage_workspace.go.html`, to allow you to break up your large HTML UI files into managable pieces.

One new re-usable UI gizmo that we have now is also the color picker.

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/color_picker.mp4" type="video/mp4">
</video>

## Stage Tools
Vertex snapping is here! You can hold down the "V" key on your keyboard, while you have entities selected, and you'll enter into the vertex snapping mode. This will allow you to select a vertex you wish to snap from within your selection, and drag it over other verts in the scene to snap to them.

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/vertex_snapping.mp4" type="video/mp4">
</video>

Next up, I added the ability to duplicate objects by holding Shift and click-dragging the translation gizmo. This also works with snapping when you hold the CTRL/Mod key.

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/drag_duplicate.mp4" type="video/mp4">
</video>

I've also added a setting in the Editor Settings to change your gizmo invocation controls from G/R/S for translate, rotate, and scale; to W/E/R (typical in software like Maya).

![Alternative transform gizmo hotkeys](https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/w_e_r_transform_tools.png)

I have introduced the ability to create primitives through the `Create` drop down in the menu bar. Also, speaking of the menu bar, you can now click to expand a section, then move your mouse back and forth through the other options to expand them; instead of needing to click each one individually.

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/auto_expand_menus.mp4" type="video/mp4">
</video>

## Physics System
Kaiju now has it's own multi-threaded physics engine built-in, directly in Go. We no longer have a dependency on Bullet3. This implementation is much further along than the previous Bullet3 implementation, so it's a fantasic improvement. The physics system has both constraints and terrain collision support as well.

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/graviton_initial_test.mp4" type="video/mp4">
</video>

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/graviton_constraints.mp4" type="video/mp4">
</video>

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/graviton_terrain.mp4" type="video/mp4">
</video>

## Terrain Editor
Kaiju now has the start of a new Terrain editor. This terrain editor allows you to paint terrain heights, it's auto-chunked for rendering performance, and you have the ability to paint textures on it. Currently it does not support foliage, but that will be coming soon.

<video autoplay muted loop playsinline max-width="100%">
	<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/blog/2026_05_23/terrain_sculpting.mp4" type="video/mp4">
</video>

## Integration Testing
You can now create integration tests in Kaiju. This allows you to create tests that actually run integrated tests (non unit test) for complex game or software systems. Something I often use this for is generating screenshots of UI elements, shader results, or other visual elements for further processing and iteration. Check out the `integrationtest` command-line arg, there is a sample `-integrationtest=screenshot` you can run. _Note that this only runs in debug builds, you can change that if you need it for release builds._

## Localization
This is a very tiny mention, but Kaiju now has a `localization` package in the code. This sets the ground work for implementing localized keyboard input, as well as prepping for a string-based localization tabling system.

## Thank you!
Thank you to all of the contributors who have stedily made Kaiju even better, improved performance, improved UX, added features, and generally have been a real blessing! The continual support and excitement from people really help inspire me to keep on moving along.
