/******************************************************************************/
/* server.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package aidriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"kaijuengine.com/engine"
)

const (
	bindHost       = "127.0.0.1"
	defaultPort    = 7777
	serviceName    = "kaiju-aidriver"
	apiVersion     = "1"
	commandTimeout = 15 * time.Second
	maxBodyBytes   = 1 << 20 // 1 MiB
	maxFrames      = 300     // upper bound for settle/hold/wait frame counts
)

// Start brings up the localhost AI-driver control server for the given host and
// registers the per-frame command drain. It is invoked from the bootstrap layer
// only under the `ai_driver` build tag, on the game-loop thread. Failures
// (e.g. the port already being in use) are logged and leave the game running
// normally, just without a control surface.
func Start(host *engine.Host) {
	if host == nil || host.Window == nil {
		slog.Error("AI Driver cannot start without a host window")
		return
	}
	d := newDriver(host)
	host.Window.OnActivate.Add(func() { d.focused = true })
	host.Window.OnDeactivate.Add(func() { d.focused = false })

	port := defaultPort
	if v := os.Getenv("AI_DRIVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 65535 {
			port = p
		} else {
			slog.Warn("AI Driver ignoring invalid AI_DRIVER_PORT", "value", v)
		}
	}
	addr := net.JoinHostPort(bindHost, strconv.Itoa(port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("AI Driver failed to bind; control surface disabled", "addr", addr, "error", err)
		return
	}

	mux := http.NewServeMux()
	d.routes(mux)
	srv := &http.Server{
		Handler:           withGuards(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	host.RunNextFrame(d.drainOnce)
	host.OnClose.Add(func() {
		close(d.closing)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("AI Driver shutdown error", "error", err)
		}
	})

	go func() {
		slog.Info("AI Driver started", "addr", addr,
			"endpoints", "/v1/health /v1/help /v1/state /v1/screenshot /v1/input /v1/resize /v1/quit")
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("AI Driver server stopped unexpectedly", "error", err)
		}
	}()
}

func (d *driver) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/health", d.handleHealth)
	mux.HandleFunc("GET /v1/help", d.handleHelp)
	mux.HandleFunc("GET /v1/state", d.handleState)
	mux.HandleFunc("GET /v1/screenshot", d.handleScreenshot)
	mux.HandleFunc("POST /v1/input", d.handleInput)
	mux.HandleFunc("POST /v1/resize", d.handleResize)
	mux.HandleFunc("POST /v1/quit", d.handleQuit)
}

// withGuards rejects browser-originated requests as a cheap anti-CSRF measure.
// curl and other CLI tools send neither header, so they pass through; a web
// page cannot drive the game even though the port is bound on localhost.
func withGuards(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (d *driver) handleScreenshot(w http.ResponseWriter, r *http.Request) {
	settle := 0
	if v := r.URL.Query().Get("settle"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			settle = min(n, maxFrames)
		}
	}
	reply := make(chan shotResult, 1)
	if !d.enqueue(&shotCommand{settle: settle, reply: reply}) {
		writeAPIError(w, &apiError{Code: codeBusy, Message: "driver command queue is full", Status: 503})
		return
	}
	select {
	case res := <-reply:
		if res.err != nil {
			writeAPIError(w, &apiError{Code: codeCaptureFailed, Message: res.err.Error(), Status: 503})
			return
		}
		writePNG(w, res.png, res.frame, res.fbW, res.fbH, res.winW, res.winH)
	case <-time.After(commandTimeout):
		writeAPIError(w, &apiError{Code: codeTimeout, Message: "screenshot timed out", Status: 504})
	}
}

func (d *driver) handleInput(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req inputRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, reqErr("invalid JSON body: "+err.Error()))
		return
	}
	if err := validateInput(&req); err != nil {
		writeAPIError(w, err)
		return
	}
	if req.SettleFrames > maxFrames {
		req.SettleFrames = maxFrames
	}
	reply := make(chan inputResult, 1)
	if !d.enqueue(&inputCommand{req: &req, reply: reply}) {
		writeAPIError(w, &apiError{Code: codeBusy, Message: "driver command queue is full", Status: 503})
		return
	}
	select {
	case res := <-reply:
		if res.reqErr != nil {
			writeAPIError(w, res.reqErr)
			return
		}
		if req.ReturnScreenshot {
			if res.shotErr != nil {
				writeAPIError(w, &apiError{Code: codeCaptureFailed, Message: res.shotErr.Error(), Status: 503})
				return
			}
			writePNG(w, res.png, res.frame, res.fbW, res.fbH, res.winW, res.winH)
			return
		}
		writeJSON(w, http.StatusOK, inputResponse{
			OK:         true,
			FrameAfter: res.frame,
			ActionsRun: res.actionsRun,
			Warnings:   res.warnings,
		})
	case <-time.After(commandTimeout):
		writeAPIError(w, &apiError{Code: codeTimeout, Message: "input request timed out", Status: 504})
	}
}

// defaultResizeSettle is the frame delay before reading back a resize. The OS
// resize is asynchronous: rendering pauses while the layer resizes and the swap
// chain rebuilds on the first frame after, so a few frames are needed before the
// new size and a screenshot are stable.
const defaultResizeSettle = 5

func (d *driver) handleResize(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req resizeRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes)).Decode(&req); err != nil {
		writeAPIError(w, reqErr("invalid JSON body: "+err.Error()))
		return
	}
	if err := validateResize(&req); err != nil {
		writeAPIError(w, err)
		return
	}
	settle := req.SettleFrames
	if settle <= 0 {
		settle = defaultResizeSettle
	}
	settle = min(settle, maxFrames)

	reply := make(chan resizeResult, 1)
	if !d.enqueue(&resizeCommand{width: req.Width, height: req.Height, settle: settle, wantShot: req.ReturnScreenshot, reply: reply}) {
		writeAPIError(w, &apiError{Code: codeBusy, Message: "driver command queue is full", Status: 503})
		return
	}
	select {
	case res := <-reply:
		if req.ReturnScreenshot {
			if res.shotErr != nil {
				writeAPIError(w, &apiError{Code: codeCaptureFailed, Message: res.shotErr.Error(), Status: 503})
				return
			}
			writePNG(w, res.png, res.frame, res.fbW, res.fbH, res.winW, res.winH)
			return
		}
		writeJSON(w, http.StatusOK, resizeResponse{
			OK:          true,
			Frame:       res.frame,
			Window:      sizeXY{Width: res.winW, Height: res.winH},
			Framebuffer: sizeXY{Width: res.fbW, Height: res.fbH},
			Scale:       computeScale(res.winW, res.winH, res.fbW, res.fbH),
		})
	case <-time.After(commandTimeout):
		writeAPIError(w, &apiError{Code: codeTimeout, Message: "resize request timed out", Status: 504})
	}
}

func (d *driver) handleQuit(w http.ResponseWriter, _ *http.Request) {
	reply := make(chan struct{}, 1)
	if !d.enqueue(&quitCommand{reply: reply}) {
		writeAPIError(w, &apiError{Code: codeBusy, Message: "driver command queue is full", Status: 503})
		return
	}
	select {
	case <-reply:
		writeJSON(w, http.StatusOK, quitResponse{OK: true, Message: "host is shutting down"})
	case <-time.After(commandTimeout):
		writeAPIError(w, &apiError{Code: codeTimeout, Message: "quit request timed out", Status: 504})
	}
}

type inputResponse struct {
	OK         bool     `json:"ok"`
	FrameAfter uint64   `json:"frame_after"`
	ActionsRun int      `json:"actions_run"`
	Warnings   []string `json:"warnings,omitempty"`
}

type quitResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type resizeResponse struct {
	OK          bool    `json:"ok"`
	Frame       uint64  `json:"frame"`
	Window      sizeXY  `json:"window"`
	Framebuffer sizeXY  `json:"framebuffer"`
	Scale       scaleXY `json:"scale"`
}

type errorEnvelope struct {
	Error *apiError `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("AI Driver failed to encode response", "error", err)
	}
}

func writeAPIError(w http.ResponseWriter, e *apiError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	if err := json.NewEncoder(w).Encode(errorEnvelope{Error: e}); err != nil {
		slog.Error("AI Driver failed to encode error", "error", err)
	}
}

func writePNG(w http.ResponseWriter, data []byte, frame uint64, fbW, fbH, winW, winH int) {
	sc := computeScale(winW, winH, fbW, fbH)
	h := w.Header()
	h.Set("Content-Type", "image/png")
	h.Set("Content-Length", strconv.Itoa(len(data)))
	h.Set("X-Aidriver-Frame", strconv.FormatUint(frame, 10))
	h.Set("X-Aidriver-Framebuffer", fmt.Sprintf("%dx%d", fbW, fbH))
	h.Set("X-Aidriver-Window", fmt.Sprintf("%dx%d", winW, winH))
	h.Set("X-Aidriver-Scale", fmt.Sprintf("%gx%g", sc.X, sc.Y))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
