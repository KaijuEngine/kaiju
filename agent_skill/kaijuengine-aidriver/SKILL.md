---
name: kaijuengine-aidriver
description: >-
  Drive an already-running Kaiju game through its built-in AI Driver HTTP server
  (localhost, enabled by the `ai_driver` build tag): capture screenshots of the
  game window and inject mouse/keyboard input (click, type, scroll, key presses),
  loop screenshot then reason then act then re-screenshot, and quit the game
  gracefully. Use this whenever the user wants to look at, inspect, interact with,
  control, or visually verify the running kaiju game — for example "screenshot the
  game", "click the start button", "type into the name field", "is the menu
  rendering correctly", or "check my UI change in the running app" — even if they
  don't mention ai_driver or curl. Prefer this over the generic run/verify skills
  for a kaiju game that exposes the AI Driver server, since it talks to the live
  process via curl rather than launching or rebuilding it. Requires the game built
  with the `ai_driver` tag.
---

# Kaiju AI Driver

You control a running kaiju game by taking screenshots and injecting
mouse/keyboard input through a localhost HTTP server compiled into the game. You
talk to it with `curl`. No authentication; it binds `127.0.0.1` only.

## Preconditions

The game must be running and built with the `ai_driver` tag:

```
go run -tags ai_driver .          # or: go build -tags ai_driver -o game . && ./game
```

On startup it logs `AI Driver started addr=127.0.0.1:7777`. Default port is
`7777`; override with the `AI_DRIVER_PORT` env var.

Confirm it is up before doing anything else:

```
curl -s http://127.0.0.1:7777/v1/health
```

If this fails to connect, the game is not running or was built without
`-tags ai_driver`. Ask the user to launch it that way; do not try to start it
yourself unless asked.

## Coordinate & timing contract (do not violate)

- **Coordinates are SCREENSHOT PIXELS, top-left origin (+x right, +y down).**
  Read a pixel straight off the screenshot you just looked at and send that exact
  pixel. The server converts to the engine's logical points for you — never
  multiply or divide coordinates yourself, even on a Retina display.
- **Input takes effect over frames, not milliseconds.** A click is a press then a
  release on a later frame; typing is one character per frame. After acting, let
  the frame settle so the screenshot reflects your input: pass `settle_frames`
  (POST) or `?settle=N` (screenshot). `settle_frames: 2` is a good default.
- Coordinates outside the framebuffer are rejected with `400 invalid_coordinate`.
  If the window may have resized, re-check `/v1/state` and re-screenshot.

## Core loop

1. **Learn the geometry** (once, and again after any resize):

   ```
   curl -s http://127.0.0.1:7777/v1/state
   ```

   Returns `window` (logical size), `framebuffer` (screenshot pixel size),
   `scale` (e.g. 2.0 on Retina), `focused`, and `frame`. The screenshot will be
   `framebuffer` pixels; send coordinates in that space.

2. **See the game** — capture a PNG to a temp file, then Read it:

   ```
   curl -s "http://127.0.0.1:7777/v1/screenshot?settle=1" --output /tmp/kaiju-shot.png
   ```

   Then use the Read tool on `/tmp/kaiju-shot.png` to view the frame.

3. **Reason** about what to do and note the target pixel from the image.

4. **Act and re-screenshot in one call** — POST input with
   `return_screenshot:true` and write the resulting frame straight to a file:

   ```
   curl -s -X POST http://127.0.0.1:7777/v1/input \
     -H 'Content-Type: application/json' \
     -d '{"settle_frames":2,"return_screenshot":true,
          "actions":[{"type":"mouse_click","button":"left","x":1280,"y":980}]}' \
     --output /tmp/kaiju-after.png
   ```

   Then Read `/tmp/kaiju-after.png` to verify the result.

5. Repeat until the task is done.

## Resizing the window

Resize the game window (dimensions are logical points, the same space as
`/v1/state`'s `window`). Rendering pauses for a couple of frames while the window
resizes and the swap chain rebuilds — the last frame is held on screen, so it's a
brief freeze, not a flicker. The call waits for the new size to become stable
before it returns, so the reported geometry (and any screenshot) already reflect
the resize; you don't need to poll, but re-read `/v1/state` if you cached sizes.

```
curl -s -X POST http://127.0.0.1:7777/v1/resize \
  -H 'Content-Type: application/json' \
  -d '{"width":1600,"height":900}'
```

It returns `{"ok":true,"frame":...,"window":{...},"framebuffer":{...},"scale":{...}}`
with the *actual* resulting size (the OS may clamp to screen or minimum bounds, so
the height you get back can be smaller than requested). Add `"return_screenshot":true`
(with `--output FILE`) to get the post-resize frame in one call, or
`"settle_frames":N` to wait extra frames before the read-back.

## Closing the game

When you are finished, shut the game down gracefully through the API instead of
killing the process:

```
curl -s -X POST http://127.0.0.1:7777/v1/quit
```

It responds `{"ok":true,"message":"host is shutting down"}` and the game closes
its window and exits on its own a couple of frames later. Do not use `pkill`.

## POST /v1/input reference

Body fields:

- `coordinate_space`: `"framebuffer"` (default; screenshot pixels) or `"window"`
  (logical points). Use the default.
- `settle_frames`: frames to advance after the actions (default applied as-is;
  use 1-2 so a returned/next screenshot reflects the input).
- `return_screenshot`: if `true`, the response body is the resulting PNG (use
  `--output FILE`); otherwise it is JSON `{ok, frame_after, actions_run, warnings}`.
- `actions`: array, run in order, each spanning one or more frames:

  | type          | fields                          | effect |
  |---------------|---------------------------------|--------|
  | `mouse_move`  | `x`, `y`                        | move cursor |
  | `mouse_down`  | `button`, `x?`, `y?`            | press and hold |
  | `mouse_up`    | `button`, `x?`, `y?`            | release |
  | `mouse_click` | `button`, `x?`, `y?`, `hold_frames?` | press then release |
  | `scroll`      | `dx`, `dy`, `x?`, `y?`          | scroll wheel |
  | `key_down`    | `key`                           | hold a key |
  | `key_up`      | `key`                           | release a key |
  | `key_press`   | `key`, `hold_frames?`           | tap a key |
  | `type_text`   | `text`                          | type a string (US-QWERTY) |
  | `wait_frames` | `frames`                        | idle N frames |

- `button`: `left`, `middle`, `right`, `x1`, `x2`.
- `key`: symbolic names, case-insensitive — `Return`/`Enter`, `Escape`/`Esc`,
  `Space`, `Tab`, `Backspace`, `Delete`, `Left`/`Right`/`Up`/`Down`, `A`..`Z`,
  `0`..`9`, `F1`..`F12`, `Home`, `End`, `PageUp`, `PageDown`, modifiers
  (`Shift`, `Ctrl`, `Alt`, `Cmd`).

Examples:

```
# Type into a focused field, then submit
curl -s -X POST http://127.0.0.1:7777/v1/input -H 'Content-Type: application/json' \
  -d '{"settle_frames":1,"actions":[{"type":"type_text","text":"Player1"},
       {"type":"key_press","key":"Return"}]}'

# Press Escape and capture the result
curl -s -X POST http://127.0.0.1:7777/v1/input -H 'Content-Type: application/json' \
  -d '{"settle_frames":2,"return_screenshot":true,
       "actions":[{"type":"key_press","key":"Escape"}]}' --output /tmp/kaiju-after.png
```

## Caveats

- `type_text` is best-effort and assumes a US-QWERTY layout; symbols outside that
  layout and IME input are not handled. For anything non-alphanumeric, prefer
  explicit `key_press` actions. Unmapped characters are skipped and reported in
  the JSON `warnings` field.
- If `/v1/state` shows `focused:false`, injected input may be ignored — ask the
  user to click the game window once to focus it.
- The screenshot is the last *presented* frame, so always use a `settle` of 1-2
  when you need to see the effect of input you just sent.
- The server only moves the in-game cursor/keyboard — it cannot drive other apps
  or the OS, so its blast radius is the game window alone.

## Debugging

```
curl -s http://127.0.0.1:7777/v1/help     # list endpoints
curl -s http://127.0.0.1:7777/v1/state    # geometry + focus + frame
```
