---
name: ui-system-overview
description: 'How Kaiju''s custom UI system works: it manually parses HTML/CSS into its own engine-driven panel/widget tree (not a real browser). Covers the fit-content layout model, scroll containers, and where the parsing/layout code lives.'
triggers:
    - kaiju UI
    - kaiju layout
    - scroll container not working
    - panel fit-content
    - markup css
    - overflow-y
    - editor HTML UI
    - details panel scroll
    - hierarchy scroll
---

# Kaiju UI System Overview

Kaiju does **not** use a real browser. It has its own custom UI engine: HTML files and CSS files are parsed by Kaiju's own HTML/CSS parser (`src/engine/ui/markup/...`) into an entity/panel tree that the engine lays out and renders itself. This means "it's just CSS" fixes often are NOT enough — you must reason about how each CSS property is translated into engine panel state. Conversely, the engine parses only the CSS subset it implements; unsupported properties are ignored.

## Where the code lives

- **HTML templates**: `src/editor/editor_embedded_content/editor_content/editor/ui/workspace/*.go.html` (e.g. `stage_workspace_details.go.html`). These are split per-panel (details, hierarchy, content) with matching `.css` files in the same directory.
- **CSS property handlers**: `src/engine/ui/markup/css/properties/css_*.go` — each maps one CSS property to panel/engine state; the registered names are in `css_property.go`.
- **Panel/layout/scrolling core**: `src/engine/ui/panel.go` (fit-content, bounds, maxScroll, scroll direction, overflow).
- **Default stylesheets**: `src/engine/ui/markup/css/default.css`.
- **Reference docs already in the repo**: `kaiju/AGENTS.md` (Layout System, Panel Operations, Scrolling sections) and `kaiju/agent_skill/kaijuengine-game-dev/reference/ui.md`.

## The fit-content model (critical)

- Every panel/`div` defaults to **fit-content on BOTH axes** — `fitContent = ContentFitBoth` (`panel.go` `panelData.initialize`). A `<div>` with no width/height grows to wrap its children.
- `width`/`height` CSS set a definite size and disable fit on that axis (via `panel.DontFitContentWidth()/DontFitContentHeight()`).
- `height: fit-content` / `width: fit-content` re-enable fit (`FitContentWidth()/FitContentHeight()`).
- `panelPostLayoutUpdate()` (`panel.go`) computes fit sizes and **clamps** a fit-content element to its parent's *available* size when the parent has a definite size. This is correct CSS-ish behavior but is the #1 source of "it won't grow / content is clipped" bugs.

## Scrolling — how it actually works

Scrolling requires the **scroll container** to have a **definite height** and its **content to actually overflow** it:

- CSS `overflow-y: scroll` (or `auto`) maps to `panel.SetScrollDirection(... | PanelScrollDirectionVertical)` + `SetOverflow(OverflowScroll)` + `GenerateScissor()` (see `css_overflow_y.go`). Setting a definite-scroll overflow also calls `DontFitContentHeight()` so the container stops fitting to content.
- The container's `maxScroll` is computed as `contentBounds - containerSize` in `panelPostLayoutUpdate`. If content never exceeds the container, `maxScroll.Y = 0` → no scrollbar and **no wheel scrolling**.
- `panel.SetScrollY(v)` requests a scroll and is clamped to `maxScroll`. `ScrollY()`/`ScrollX()` read current pos; `RunAfterFrames` is sometimes needed to apply scroll after layout.

### The classic bug pattern (details panel)

A `div` that is a **fit-content wrapper** (default) placed as the **only direct child** of a definite-height scroll container gets **clamped to the container's height** by the fit-content clamp. The container then sees no overflow → scrollbar gone and content clipped at the bottom. You cannot scroll to it.

**Fixed pattern (mirror the hierarchy panel):** make the *wrapper itself* the definite-height scroll container with a definite width/height:

```css
#detailsArea { /* fixed panel, definite height via sideBarStandard */ }
.detailsBody {
    width: 100%;
    height: 100%;
    overflow-y: scroll;
}
```

Then the fit-content children are free to overflow it and a scrollbar/wheel works. This matches the hierarchy panel, where `#entityList` is `height: calc(100% - 3em); overflow-y: scroll`.

### Important gotchas

- If a scroll container directly contains many **small fit-content rows** (hierarchy), no single child exceeds the container, so the clamp never collapses it and it overflows/scrolls normally. The collapse only bites when one big fit-content wrapper is the direct child.
- If you move which element scrolls, update any Go `SetScrollY(0)` reset calls to target the **new scroll container** (e.g. after moving scroll from `detailsArea` to its first child `detailsBody`).
- Scrollbars are separate `PanelScrollDirection` gating; `overflow: hidden` disables scroll and clips.

## Editor UI patterns worth knowing

- Editor panels are absolutely positioned in the document; e.g. `#detailsArea` is `position:absolute; right:0; top:var(--ed-menu-bar-height); width:18%` with its height supplied by classes like `sideBarStandard` (70%) / `sideBarTall` (100%) from `stage_workspace_panels.css`.
- Template-driven lists: a template element (e.g. `#boundEntityDataTemplate`) is duplicated at runtime by the Go workspace UI code (`doc.DuplicateElement*`); runtime-created children are added while hidden then shown. Keep templates hidden.
- `document.Element` ↔ `UI`/`Panel` bridge: `element.UI.ToPanel()`, `element.UI.Hide()/Show()`, `doc.GetElementById(...)`, `element.Children`.

## When fixing a UI bug

1. Reproduce which element should scroll and its definite size.
2. Check for a fit-content wrapper between the scroll container and the content.
3. Prefer a targeted HTML/CSS-level fix (moving the scroll/definite height onto the right container) over changing engine `panel.go` clamp/scroll logic, because engine changes alter behavior app-wide.
4. Build (`go build ./...`) and run `go test ./engine/ui/... ./editor/...` to verify.
