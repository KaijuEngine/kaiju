/******************************************************************************/
/* state.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package aidriver

import (
	"net/http"
	"time"
)

type sizeXY struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type scaleXY struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// stateResponse is the body of GET /v1/state. It is the contract anchor that
// lets a client translate screenshot pixels into the logical points the engine
// expects: window is logical size, framebuffer is the screenshot's pixel size,
// and scale is framebuffer/window per axis (2.0 on a Retina display).
type stateResponse struct {
	Frame           uint64  `json:"frame"`
	Focused         bool    `json:"focused"`
	Window          sizeXY  `json:"window"`
	Framebuffer     sizeXY  `json:"framebuffer"`
	Scale           scaleXY `json:"scale"`
	CoordinateSpace string  `json:"coordinate_space"`
}

type healthResponse struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type routeHelp struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type helpResponse struct {
	Version   string      `json:"version"`
	Endpoints []routeHelp `json:"endpoints"`
}

var helpRoutes = []routeHelp{
	{http.MethodGet, "/v1/health", "Liveness check."},
	{http.MethodGet, "/v1/help", "This endpoint list."},
	{http.MethodGet, "/v1/state", "Window/framebuffer size, retina scale, focus, frame number."},
	{http.MethodGet, "/v1/screenshot", "PNG of the last presented frame. Query: settle=N waits N frames first."},
	{http.MethodPost, "/v1/input", "Batched input actions. See coordinate_space, settle_frames, return_screenshot, actions[]."},
	{http.MethodPost, "/v1/resize", "Resize the window: {width, height, settle_frames?, return_screenshot?}."},
	{http.MethodPost, "/v1/quit", "Gracefully close the game."},
}

func computeScale(winW, winH, fbW, fbH int) scaleXY {
	sx, sy := float32(1), float32(1)
	if winW > 0 {
		sx = float32(fbW) / float32(winW)
	}
	if winH > 0 {
		sy = float32(fbH) / float32(winH)
	}
	return scaleXY{X: sx, Y: sy}
}

// stateCommand reads window/framebuffer/focus state on the game-loop thread.
type stateCommand struct {
	reply chan stateResponse
}

func (c *stateCommand) apply(d *driver) {
	winW := d.host.Window.Width()
	winH := d.host.Window.Height()
	fbW, fbH := d.framebufferSize(winW, winH)
	c.reply <- stateResponse{
		Frame:           d.host.Frame(),
		Focused:         d.focused,
		Window:          sizeXY{Width: winW, Height: winH},
		Framebuffer:     sizeXY{Width: fbW, Height: fbH},
		Scale:           computeScale(winW, winH, fbW, fbH),
		CoordinateSpace: spaceFramebuffer,
	}
}

func (d *driver) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{OK: true, Service: serviceName, Version: apiVersion})
}

func (d *driver) handleHelp(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, helpResponse{Version: apiVersion, Endpoints: helpRoutes})
}

func (d *driver) handleState(w http.ResponseWriter, _ *http.Request) {
	reply := make(chan stateResponse, 1)
	if !d.enqueue(&stateCommand{reply: reply}) {
		writeAPIError(w, &apiError{Code: codeBusy, Message: "driver command queue is full", Status: 503})
		return
	}
	select {
	case s := <-reply:
		writeJSON(w, http.StatusOK, s)
	case <-time.After(commandTimeout):
		writeAPIError(w, &apiError{Code: codeTimeout, Message: "state request timed out", Status: 504})
	}
}
