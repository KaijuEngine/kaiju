---
title: Kaiju Engine
description: Kaiju Engine is an extremely fast, open source game engine and editor written in Go, powered by Vulkan, supporting 2D/3D, physics, UI, and live shader updates.
keywords: game engine, Go, Vulkan, 2D, 3D, physics, UI, shader, cross-platform
---

<div class="kaiju-landing" id="top">
	<section class="kl-hero">
		<div class="kl-container kl-hero-grid">
			<div class="kl-hero-copy">
				<p class="kl-eyebrow"><span></span> Open source 2D and 3D game engine</p>
				<h1>Kaiju Engine</h1>
				<p class="kl-lede">
					Build fast games in Go with a Vulkan renderer, a flexible editor, custom UI,
					physics, particles, animation, audio, and live shader iteration.
				</p>
				<div class="kl-actions">
					<a class="kl-btn kl-btn-primary" href="getting_started/start_with_editor/">Start with editor</a>
					<a class="kl-btn kl-btn-secondary" href="getting_started/start_without_editor/">Start without editor</a>
					<a class="kl-btn kl-btn-secondary" href="engine/build_from_source/">Build from source</a>
				</div>
				<div class="kl-platform-row">
					<span>Go</span>
					<span>Vulkan</span>
					<span>Windows</span>
					<span>Linux</span>
					<span>macOS</span>
					<span>Android</span>
				</div>
			</div>

			<div class="kl-engine-frame" aria-label="Kaiju editor preview">
				<div class="kl-frame-toolbar">
					<div class="kl-window-dots" aria-hidden="true"><span></span><span></span><span></span></div>
					<span>kaiju-editor / scene</span>
				</div>
				<div class="kl-viewport">
					<video autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng">
						<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.mp4" type="video/mp4">
						<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng" alt="Kaiju editor preview">
					</video>
				</div>
			</div>
		</div>
	</section>

	<section class="kl-proof">
		<div class="kl-container kl-proof-grid" aria-label="Kaiju engine capabilities">
			<span>2D</span>
			<span>3D</span>
			<span>Physics</span>
			<span>Particles</span>
			<span>UI</span>
			<span>Audio</span>
		</div>
	</section>

	<section class="kl-section" id="workflow">
		<div class="kl-container">
			<div class="kl-section-heading">
				<p class="kl-eyebrow"><span></span> Choose your workflow</p>
				<h2>Visual editing or code-first control.</h2>
				<p>
					Use the editor for scenes, assets, and iteration, or work directly with
					the engine from Go. Both paths use the same runtime.
				</p>
			</div>

			<div class="kl-workflow-grid">
				<a class="kl-card kl-card-large" href="getting_started/start_with_editor/">
					<span class="kl-card-icon">E</span>
					<h3>Start with the editor</h3>
					<p>A visual workspace for scenes, content, shader previews, and play testing.</p>
					<ul class="kl-check-list">
						<li>Integrated Go project workflow</li>
						<li>Stage, content, shading, VFX, and UI tools</li>
						<li>The editor itself runs inside Kaiju</li>
					</ul>
					<span class="kl-text-link">Open guide <span aria-hidden="true">-&gt;</span></span>
				</a>

				<a class="kl-card kl-card-large" href="getting_started/start_without_editor/">
					<span class="kl-card-icon">Go</span>
					<h3>Start without the editor</h3>
					<p>Use Kaiju as a Go engine when you want direct control over runtime code.</p>
					<ul class="kl-check-list">
						<li>Write gameplay and systems in Go</li>
						<li>Use engine caches, entities, transforms, and drawings</li>
						<li>Bring your own project structure</li>
					</ul>
					<span class="kl-text-link">Open guide <span aria-hidden="true">-&gt;</span></span>
				</a>
			</div>
		</div>
	</section>

	<section class="kl-section kl-news" id="news">
		<div class="kl-container">
			<div class="kl-section-heading">
				<p class="kl-eyebrow"><span></span> Latest updates</p>
				<h2>News</h2>
			</div>

			<div class="kl-blog-grid" id="kaiju-news-list" data-news-src="blog/posts.json">
				<p class="kl-news-loading">Loading news...</p>
			</div>
		</div>
	</section>

	<section class="kl-section" id="features">
		<div class="kl-container">
			<div class="kl-section-heading kl-center">
				<p class="kl-eyebrow"><span></span> Engine features</p>
				<h2>Everything you need to build games.</h2>
				<p>
					Kaiju combines a Go runtime with production-minded engine systems for
					rendering, tools, content, and fast iteration.
				</p>
			</div>

			<div class="kl-feature-grid">
				<a class="kl-card kl-feature-card" href="editor/stage/">
					<span>01</span>
					<h3>Editor</h3>
					<p>Build scenes and inspect assets in a visual editor powered by the engine.</p>
				</a>
				<a class="kl-card kl-feature-card" href="editor/vfx/">
					<span>02</span>
					<h3>Particles and VFX</h3>
					<p>Compose emitters into reusable particle systems for gameplay feedback.</p>
				</a>
				<a class="kl-card kl-feature-card" href="engine/ui/writing/">
					<span>03</span>
					<h3>Custom UI</h3>
					<p>Create retained-mode UI directly or with the HTML/CSS-like markup system.</p>
				</a>
				<a class="kl-card kl-feature-card" href="engine/physics_constraints/">
					<span>04</span>
					<h3>Physics</h3>
					<p>Simulate 3D worlds with physics systems that integrate with entities.</p>
				</a>
				<a class="kl-card kl-feature-card" href="editor/shading/">
					<span>05</span>
					<h3>Live shaders</h3>
					<p>Iterate on GLSL shader code and visualize changes without breaking flow.</p>
				</a>
				<a class="kl-card kl-feature-card" href="engine/build_from_source/">
					<span>06</span>
					<h3>Cross platform</h3>
					<p>Create on desktop and deploy to Windows, Linux, macOS, and Android.</p>
				</a>
			</div>
		</div>
	</section>

	<section class="kl-section" id="showcase">
		<div class="kl-container">
			<div class="kl-section-heading">
				<p class="kl-eyebrow"><span></span> Showcase</p>
				<h2>One runtime, multiple ways to create.</h2>
			</div>

			<div class="kl-showcase-layout">
				<div class="kl-showcase-main">
					<video id="kaiju-showcase-video" autoplay muted loop playsinline poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng">
						<source src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.mp4" type="video/mp4">
						<img src="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng" alt="Kaiju editor preview">
					</video>
					<div class="kl-showcase-gradient" aria-hidden="true"></div>
					<div class="kl-showcase-caption">
						<h3 id="kaiju-showcase-title">Editor overview</h3>
						<p id="kaiju-showcase-description">A visual workspace for scenes, assets, previews, and iteration.</p>
					</div>
				</div>

				<div class="kl-showcase-list" aria-label="Showcase videos">
					<button class="kl-showcase-thumb is-active" type="button" aria-pressed="true"
						data-title="Editor overview"
						data-description="A visual workspace for scenes, assets, previews, and iteration."
						data-video="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.mp4"
						data-poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/editor.apng">
						<span>Editor</span>
						<strong>Scene and content workflow</strong>
					</button>
					<button class="kl-showcase-thumb" type="button" aria-pressed="false"
						data-title="2D workflow"
						data-description="Switch to 2D mode for sprites, UI, animation, effects, and gameplay."
						data-video="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.mp4"
						data-poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/2d.apng">
						<span>2D</span>
						<strong>Sprites, animation, and UI</strong>
					</button>
					<button class="kl-showcase-thumb" type="button" aria-pressed="false"
						data-title="3D rendering"
						data-description="Build native 3D scenes backed by Kaiju's custom math library."
						data-video="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.mp4"
						data-poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/3d.apng">
						<span>3D</span>
						<strong>Scenes, lighting, materials</strong>
					</button>
					<button class="kl-showcase-thumb" type="button" aria-pressed="false"
						data-title="Particles and VFX"
						data-description="Compose multiple particle emitters into flexible visual effects."
						data-video="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.mp4"
						data-poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/particle_systems.apng">
						<span>VFX</span>
						<strong>Particles and effects</strong>
					</button>
					<button class="kl-showcase-thumb" type="button" aria-pressed="false"
						data-title="Live shader editing"
						data-description="Visualize GLSL shader code in real time while you iterate."
						data-video="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.mp4"
						data-poster="https://github.com/KaijuEngine/kaiju_media_files/raw/refs/heads/master/docs/index.md/live_shader.apng">
						<span>Shaders</span>
						<strong>Fast visual iteration</strong>
					</button>
				</div>
			</div>
		</div>
	</section>

	<section class="kl-section kl-velocity">
		<div class="kl-container kl-velocity-grid">
			<div class="kl-section-heading">
				<p class="kl-eyebrow"><span></span> Developer velocity</p>
				<h2>Fast iteration from code to play.</h2>
				<p>
					Kaiju is designed for short feedback loops: edit gameplay, build quickly,
					launch, test, and keep creating.
				</p>
				<a class="kl-btn kl-btn-secondary" href="engine/build_from_source/">Build from source</a>
			</div>

			<div class="kl-terminal-card">
				<div class="kl-terminal-toolbar"><span></span><span></span><span></span></div>
				<pre><code>cd src
go build -tags="debug,editor,filedrop" -o ../ ./

../kaijuengine.com.exe</code></pre>
			</div>
		</div>
	</section>

	<section class="kl-section kl-platforms">
		<div class="kl-container">
			<div class="kl-section-heading kl-center">
				<p class="kl-eyebrow"><span></span> Cross platform</p>
				<h2>Create on desktop. Deploy across platforms.</h2>
			</div>
			<div class="kl-platform-matrix">
				<div>
					<h3>Create on</h3>
					<div class="kl-platform-list"><span>Windows</span><span>Linux</span><span>macOS</span></div>
				</div>
				<div>
					<h3>Deploy to</h3>
					<div class="kl-platform-list"><span>Windows</span><span>Linux</span><span>macOS</span><span>Android</span></div>
				</div>
				<div>
					<h3>Built for</h3>
					<div class="kl-platform-list"><span>2D games</span><span>3D games</span><span>Tools</span><span>Prototypes</span></div>
				</div>
			</div>
		</div>
	</section>

	<section class="kl-section kl-start" id="download">
		<div class="kl-container kl-start-grid">
			<div class="kl-section-heading">
				<p class="kl-eyebrow"><span></span> Start building</p>
				<h2>Download Kaiju and create your first project.</h2>
				<p>
					Start with the editor for visual development, or use the engine directly
					from Go for a code-first workflow.
				</p>
				<div class="kl-actions">
					<a class="kl-btn kl-btn-primary" href="https://github.com/KaijuEngine/kaiju/releases" target="_blank" rel="noreferrer">Download latest build</a>
					<a class="kl-btn kl-btn-secondary" href="getting_started/start_with_editor/">Read docs</a>
				</div>
			</div>
			<ol class="kl-steps-card">
				<li><span>01</span><div><strong>Download Kaiju</strong><p>Grab the latest editor build or build the engine from source.</p></div></li>
				<li><span>02</span><div><strong>Create a project</strong><p>Set up your project, assets, scenes, and runtime configuration.</p></div></li>
				<li><span>03</span><div><strong>Build your first scene</strong><p>Add objects, materials, cameras, UI, physics, and effects.</p></div></li>
				<li><span>04</span><div><strong>Write gameplay in Go</strong><p>Use Go for systems, behavior, tools, and game logic.</p></div></li>
			</ol>
		</div>
	</section>

	<section class="kl-section kl-community" id="community">
		<div class="kl-container kl-community-card">
			<div>
				<p class="kl-eyebrow"><span></span> Community</p>
				<h2>Help shape the future of Kaiju.</h2>
				<p>
					Follow development, contribute code, report issues, share projects,
					and connect with other developers building in Go.
				</p>
			</div>
			<div class="kl-actions">
				<a class="kl-btn kl-btn-primary" href="https://github.com/KaijuEngine/kaiju" target="_blank" rel="noreferrer">GitHub</a>
				<a class="kl-btn kl-btn-secondary" href="https://discord.gg/8rFPEu8U52" target="_blank" rel="noreferrer">Discord</a>
				<a class="kl-btn kl-btn-secondary" href="https://github.com/sponsors/BrentFarris" target="_blank" rel="noreferrer">Sponsor</a>
			</div>
		</div>
	</section>

	<section class="kl-section kl-sponsors">
		<div class="kl-container">
			<div class="kl-section-heading">
				<p class="kl-eyebrow"><span></span> Sponsors</p>
				<h2>Thank you.</h2>
				<p>If you like what you see, please consider supporting Kaiju's continued development.</p>
				<iframe src="https://github.com/sponsors/BrentFarris/button" title="Sponsor BrentFarris" height="32" width="114" style="border: 0; border-radius: 6px;"></iframe>
			</div>
			<table id="sponsors" class="kl-sponsor-table">
				<tr>
					<th>Name</th>
					<th>GitHub</th>
				</tr>
			</table>
		</div>
	</section>
</div>
