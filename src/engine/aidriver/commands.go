/******************************************************************************/
/* commands.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package aidriver

import (
	"fmt"
	"strconv"

	"kaijuengine.com/platform/hid"
)

// Coordinate spaces accepted by the input API.
const (
	spaceFramebuffer = "framebuffer"
	spaceWindow      = "window"
)

// Action type discriminators for the POST /v1/input payload.
const (
	actionMouseMove  = "mouse_move"
	actionMouseDown  = "mouse_down"
	actionMouseUp    = "mouse_up"
	actionMouseClick = "mouse_click"
	actionScroll     = "scroll"
	actionKeyDown    = "key_down"
	actionKeyUp      = "key_up"
	actionKeyPress   = "key_press"
	actionTypeText   = "type_text"
	actionWaitFrames = "wait_frames"
)

// Error envelope codes returned to clients.
const (
	codeInvalidRequest    = "invalid_request"
	codeInvalidCoordinate = "invalid_coordinate"
	codeInvalidKey        = "invalid_key"
	codeInvalidButton     = "invalid_button"
	codeCaptureFailed     = "capture_failed"
	codeBusy              = "busy"
	codeTimeout           = "timeout"
)

// apiError is the single error envelope marshaled to clients. It implements the
// error interface so it can flow through normal Go error handling.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *apiError) Error() string { return e.Message }

func reqErr(msg string) *apiError {
	return &apiError{Code: codeInvalidRequest, Message: msg, Status: 400}
}

func coordErr(x, y float32, w, h int, space string) *apiError {
	return &apiError{
		Code:    codeInvalidCoordinate,
		Message: fmt.Sprintf("coordinate (%g,%g) is outside the %s bounds %dx%d", x, y, space, w, h),
		Status:  400,
	}
}

// toLogical converts a coordinate from the given space into logical window
// points (what hid.Mouse.SetPosition expects), validating bounds. In framebuffer
// space the coordinate is in screenshot pixels and is divided by the per-axis
// scale (framebuffer/window) so a client can use raw pixels off the screenshot;
// in window space it is passed through unchanged.
func toLogical(space string, x, y float32, winW, winH, fbW, fbH int) (float32, float32, *apiError) {
	if space == spaceWindow {
		if x < 0 || y < 0 || x >= float32(winW) || y >= float32(winH) {
			return 0, 0, coordErr(x, y, winW, winH, space)
		}
		return x, y, nil
	}
	if x < 0 || y < 0 || x >= float32(fbW) || y >= float32(fbH) {
		return 0, 0, coordErr(x, y, fbW, fbH, space)
	}
	return x * float32(winW) / float32(fbW), y * float32(winH) / float32(fbH), nil
}

// inputRequest is the decoded body of POST /v1/input.
type inputRequest struct {
	CoordinateSpace  string        `json:"coordinate_space"`
	SettleFrames     int           `json:"settle_frames"`
	ReturnScreenshot bool          `json:"return_screenshot"`
	Actions          []inputAction `json:"actions"`
}

// inputAction is a single step in a batched input request. X/Y are pointers so
// "omitted" is distinguishable from the value 0.
type inputAction struct {
	Type       string   `json:"type"`
	X          *float32 `json:"x"`
	Y          *float32 `json:"y"`
	DX         float32  `json:"dx"`
	DY         float32  `json:"dy"`
	Button     string   `json:"button"`
	Key        string   `json:"key"`
	Text       string   `json:"text"`
	HoldFrames int      `json:"hold_frames"`
	Frames     int      `json:"frames"`
}

func (r *inputRequest) space() string {
	if r.CoordinateSpace == "" {
		return spaceFramebuffer
	}
	return r.CoordinateSpace
}

// validateInput checks everything that can be verified without engine state:
// known action types, valid key/button names, and required-field presence.
// Coordinate bounds depend on the live framebuffer scale and are checked later
// on the game-loop thread.
func validateInput(req *inputRequest) *apiError {
	if s := req.CoordinateSpace; s != "" && s != spaceFramebuffer && s != spaceWindow {
		return reqErr("coordinate_space must be 'framebuffer' or 'window'")
	}
	if len(req.Actions) == 0 {
		return reqErr("actions must not be empty")
	}
	for i := range req.Actions {
		a := &req.Actions[i]
		switch a.Type {
		case actionMouseMove:
			if a.X == nil || a.Y == nil {
				return reqErr("mouse_move requires x and y")
			}
		case actionMouseDown, actionMouseUp, actionMouseClick:
			if _, ok := lookupButton(a.Button); !ok {
				return &apiError{Code: codeInvalidButton, Message: "unknown button: " + a.Button, Status: 400}
			}
			if (a.X == nil) != (a.Y == nil) {
				return reqErr(a.Type + " requires both x and y, or neither")
			}
		case actionScroll:
			if (a.X == nil) != (a.Y == nil) {
				return reqErr("scroll requires both x and y, or neither")
			}
		case actionKeyDown, actionKeyUp, actionKeyPress:
			if _, ok := lookupKey(a.Key); !ok {
				return &apiError{Code: codeInvalidKey, Message: "unknown key: " + a.Key, Status: 400}
			}
		case actionTypeText:
			if a.Text == "" {
				return reqErr("type_text requires text")
			}
		case actionWaitFrames:
			// frames < 1 is normalized to 1 at apply time.
		default:
			return reqErr("unknown action type: " + a.Type)
		}
	}
	return nil
}

// inputResult is delivered from the game-loop thread back to the HTTP handler.
type inputResult struct {
	png        []byte
	fbW, fbH   int
	winW, winH int
	frame      uint64
	actionsRun int
	warnings   []string
	reqErr     *apiError // request-level failure (e.g. out-of-bounds coordinate)
	shotErr    error     // screenshot capture failure
}

// inputCommand executes a batched input request on the game-loop thread,
// spreading multi-frame effects (clicks, typing, settle) across frames via
// host.RunAfterFrames, then replies (optionally with a screenshot).
type inputCommand struct {
	req   *inputRequest
	reply chan inputResult
}

func (c *inputCommand) apply(d *driver) {
	host := d.host
	winW := host.Window.Width()
	winH := host.Window.Height()
	fwinW, fwinH := float32(winW), float32(winH)
	fbW, fbH := d.framebufferSize(winW, winH)
	space := c.req.space()

	// convert maps a request coordinate to logical window points, validating
	// bounds against the active coordinate space.
	convert := func(x, y float32) (float32, float32, *apiError) {
		return toLogical(space, x, y, winW, winH, fbW, fbH)
	}

	// Pre-validate every coordinate so a bad request applies nothing.
	for i := range c.req.Actions {
		a := &c.req.Actions[i]
		if a.X != nil && a.Y != nil {
			if _, _, err := convert(*a.X, *a.Y); err != nil {
				c.reply <- inputResult{reqErr: err}
				return
			}
		}
	}

	warnings := []string{}
	cursor := 1 // frame offsets start at 1 (offset 0 would collide with 1)
	schedule := host.RunAfterFrames

	for i := range c.req.Actions {
		a := &c.req.Actions[i]
		switch a.Type {
		case actionMouseMove:
			lx, ly, _ := convert(*a.X, *a.Y)
			f := cursor
			schedule(f, func() { host.Window.Mouse.SetPosition(lx, ly, fwinW, fwinH) })
			cursor = f + 1
		case actionMouseDown, actionMouseUp:
			btn, _ := lookupButton(a.Button)
			lx, ly, hasPos := resolvePos(a, convert)
			down := a.Type == actionMouseDown
			f := cursor
			schedule(f, func() {
				if hasPos {
					host.Window.Mouse.SetPosition(lx, ly, fwinW, fwinH)
				}
				if down {
					host.Window.Mouse.SetDown(btn)
				} else {
					host.Window.Mouse.SetUp(btn)
				}
			})
			cursor = f + 1
		case actionMouseClick:
			btn, _ := lookupButton(a.Button)
			hold := max(a.HoldFrames, 1)
			lx, ly, hasPos := resolvePos(a, convert)
			f := cursor
			schedule(f, func() {
				if hasPos {
					host.Window.Mouse.SetPosition(lx, ly, fwinW, fwinH)
				}
				host.Window.Mouse.SetDown(btn)
			})
			schedule(f+hold, func() { host.Window.Mouse.SetUp(btn) })
			cursor = f + hold + 1
		case actionScroll:
			dx, dy := a.DX, a.DY
			lx, ly, hasPos := resolvePos(a, convert)
			f := cursor
			schedule(f, func() {
				if hasPos {
					host.Window.Mouse.SetPosition(lx, ly, fwinW, fwinH)
				}
				host.Window.Mouse.SetScroll(dx, dy)
			})
			cursor = f + 1
		case actionKeyDown:
			key, _ := lookupKey(a.Key)
			f := cursor
			schedule(f, func() { host.Window.Keyboard.SetKeyDown(key) })
			cursor = f + 1
		case actionKeyUp:
			key, _ := lookupKey(a.Key)
			f := cursor
			schedule(f, func() { host.Window.Keyboard.SetKeyUp(key) })
			cursor = f + 1
		case actionKeyPress:
			key, _ := lookupKey(a.Key)
			f := cursor
			if a.HoldFrames >= 1 {
				hold := a.HoldFrames
				schedule(f, func() { host.Window.Keyboard.SetKeyDown(key) })
				schedule(f+hold, func() { host.Window.Keyboard.SetKeyUp(key) })
				cursor = f + hold + 1
			} else {
				schedule(f, func() { host.Window.Keyboard.SetKeyDownUp(key) })
				cursor = f + 1
			}
		case actionTypeText:
			f := cursor
			for _, r := range a.Text {
				key, shift, ok := runeToKey(r)
				if !ok {
					warnings = append(warnings, "type_text: "+strconv.QuoteRune(r)+" has no key mapping, skipped")
					continue
				}
				k := key
				if shift {
					schedule(f, func() {
						host.Window.Keyboard.SetKeyDown(hid.KeyboardKeyLeftShift)
						host.Window.Keyboard.SetKeyDownUp(k)
					})
					schedule(f+1, func() { host.Window.Keyboard.SetKeyUp(hid.KeyboardKeyLeftShift) })
					f += 2
				} else {
					schedule(f, func() { host.Window.Keyboard.SetKeyDownUp(k) })
					f++
				}
			}
			cursor = f
		case actionWaitFrames:
			cursor += max(a.Frames, 1)
		}
	}

	settle := max(c.req.SettleFrames, 0)
	actionsRun := len(c.req.Actions)
	wantShot := c.req.ReturnScreenshot
	schedule(cursor+settle, func() {
		res := inputResult{
			frame:      host.Frame(),
			actionsRun: actionsRun,
			warnings:   warnings,
			winW:       winW,
			winH:       winH,
			fbW:        fbW,
			fbH:        fbH,
		}
		if wantShot {
			png, w, h, err := capturePNG(host)
			res.png, res.shotErr = png, err
			if err == nil {
				res.fbW, res.fbH = w, h
			}
		}
		c.reply <- res
	})
}

// resolvePos converts an action's optional coordinate. The coordinate has
// already passed pre-validation, so the error is discarded here.
func resolvePos(a *inputAction, convert func(x, y float32) (float32, float32, *apiError)) (float32, float32, bool) {
	if a.X == nil || a.Y == nil {
		return 0, 0, false
	}
	lx, ly, _ := convert(*a.X, *a.Y)
	return lx, ly, true
}

// shotResult is delivered from the game-loop thread to the GET /v1/screenshot
// handler.
type shotResult struct {
	png        []byte
	fbW, fbH   int
	winW, winH int
	frame      uint64
	err        error
}

// shotCommand captures a screenshot, optionally after settling N frames so the
// image reflects recently injected input.
type shotCommand struct {
	settle int
	reply  chan shotResult
}

func (c *shotCommand) apply(d *driver) {
	capture := func() {
		png, w, h, err := capturePNG(d.host)
		c.reply <- shotResult{
			png:   png,
			fbW:   w,
			fbH:   h,
			winW:  d.host.Window.Width(),
			winH:  d.host.Window.Height(),
			frame: d.host.Frame(),
			err:   err,
		}
	}
	if c.settle > 0 {
		d.host.RunAfterFrames(c.settle, capture)
	} else {
		capture()
	}
}

// quitCommand gracefully shuts the game down. It replies first so the HTTP
// response can flush, then triggers host.Close a couple of frames later; the
// game loop then exits and tears the server down via host.OnClose.
type quitCommand struct {
	reply chan struct{}
}

func (c *quitCommand) apply(d *driver) {
	close(c.reply)
	d.host.RunAfterFrames(2, d.host.Close)
}

// resizeRequest is the decoded body of POST /v1/resize. Width/Height are the
// requested logical window size in points.
type resizeRequest struct {
	Width            int  `json:"width"`
	Height           int  `json:"height"`
	SettleFrames     int  `json:"settle_frames"`
	ReturnScreenshot bool `json:"return_screenshot"`
}

const maxWindowDim = 16384

func validateResize(req *resizeRequest) *apiError {
	if req.Width < 1 || req.Height < 1 {
		return reqErr("width and height must be at least 1")
	}
	if req.Width > maxWindowDim || req.Height > maxWindowDim {
		return reqErr(fmt.Sprintf("width and height must be at most %d", maxWindowDim))
	}
	return nil
}

type resizeResult struct {
	png        []byte
	fbW, fbH   int
	winW, winH int
	frame      uint64
	shotErr    error
}

// resizeCommand resizes the window on the game-loop thread. The OS resize is
// asynchronous (on macOS it dispatches to the main thread), and the swap chain
// is recreated via the normal OnResize path, so the result is read after a
// settle delay — that is when the new framebuffer size and scale are stable.
type resizeCommand struct {
	width, height int
	settle        int
	wantShot      bool
	reply         chan resizeResult
}

const maxResizeAttempts = 12

func (c *resizeCommand) apply(d *driver) {
	beforeW, beforeH := d.framebufferSize(d.host.Window.Width(), d.host.Window.Height())
	d.host.Window.SetSize(c.width, c.height)
	settle := c.settle
	if settle <= 0 {
		settle = 1
	}
	requestedChange := c.width != beforeW || c.height != beforeH
	attempts := 0
	var finish func()
	finish = func() {
		winW := d.host.Window.Width()
		winH := d.host.Window.Height()
		fbW, fbH := d.framebufferSize(winW, winH)
		// The resize lands on the main thread asynchronously and the swap chain
		// rebuilds a frame or two later. While the framebuffer is still the
		// pre-resize size the rebuild hasn't happened yet, so wait and retry —
		// otherwise the reported geometry (and any screenshot) would be stale.
		if requestedChange && fbW == beforeW && fbH == beforeH && attempts < maxResizeAttempts {
			attempts++
			d.host.RunAfterFrames(2, finish)
			return
		}
		res := resizeResult{winW: winW, winH: winH, fbW: fbW, fbH: fbH, frame: d.host.Frame()}
		if c.wantShot {
			png, w, h, err := capturePNG(d.host)
			res.png, res.shotErr = png, err
			if err == nil {
				res.fbW, res.fbH = w, h
			}
		}
		c.reply <- res
	}
	d.host.RunAfterFrames(settle, finish)
}
