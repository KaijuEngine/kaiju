/******************************************************************************/
/* aidriver_test.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

// These are white-box tests (package aidriver, not aidriver_test): the only
// exported symbol is Start, which requires a live GPU-backed host, so the
// meaningful unit-testable logic (key mapping, retina coordinate scaling,
// request validation, and host-independent HTTP paths) lives unexported and is
// exercised here directly.
package aidriver

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/platform/hid"
)

func TestLookupKey(t *testing.T) {
	cases := map[string]hid.KeyboardKey{
		"Return": hid.KeyboardKeyReturn,
		"enter":  hid.KeyboardKeyReturn,
		"  ESC ": hid.KeyboardKeyEscape,
		"f5":     hid.KeyboardKeyF5,
		"a":      hid.KeyboardKeyA,
		"Z":      hid.KeyboardKeyZ,
		"space":  hid.KeyboardKeySpace,
		"left":   hid.KeyboardKeyLeft,
		"cmd":    hid.KeyboardKeyLeftMeta,
	}
	for name, want := range cases {
		got, ok := lookupKey(name)
		if !ok {
			t.Fatalf("lookupKey(%q): expected found", name)
		}
		if got != want {
			t.Errorf("lookupKey(%q) = %v, want %v", name, got, want)
		}
	}
	if _, ok := lookupKey("definitely-not-a-key"); ok {
		t.Errorf("lookupKey(unknown): expected not found")
	}
}

func TestLookupButton(t *testing.T) {
	b, ok := lookupButton("LEFT")
	if !ok {
		t.Fatalf("lookupButton(LEFT): expected found")
	}
	if b != hid.MouseButtonLeft {
		t.Errorf("lookupButton(LEFT) = %v, want %v", b, hid.MouseButtonLeft)
	}
	b, ok = lookupButton("right")
	if !ok {
		t.Fatalf("lookupButton(right): expected found")
	}
	if b != hid.MouseButtonRight {
		t.Errorf("lookupButton(right) = %v, want %v", b, hid.MouseButtonRight)
	}
	if _, ok = lookupButton("scrollwheel"); ok {
		t.Errorf("lookupButton(scrollwheel): expected not found")
	}
}

func TestRuneToKey(t *testing.T) {
	k, shift, ok := runeToKey('a')
	if !ok {
		t.Fatalf("runeToKey('a'): expected found")
	}
	if k != hid.KeyboardKeyA {
		t.Errorf("runeToKey('a') key = %v, want %v", k, hid.KeyboardKeyA)
	}
	if shift {
		t.Errorf("runeToKey('a') shift = true, want false")
	}

	k, shift, ok = runeToKey('A')
	if !ok {
		t.Fatalf("runeToKey('A'): expected found")
	}
	if k != hid.KeyboardKeyA {
		t.Errorf("runeToKey('A') key = %v, want %v", k, hid.KeyboardKeyA)
	}
	if !shift {
		t.Errorf("runeToKey('A') shift = false, want true")
	}

	k, shift, ok = runeToKey('!')
	if !ok {
		t.Fatalf("runeToKey('!'): expected found")
	}
	if k != hid.KeyboardKey1 {
		t.Errorf("runeToKey('!') key = %v, want %v", k, hid.KeyboardKey1)
	}
	if !shift {
		t.Errorf("runeToKey('!') shift = false, want true")
	}

	k, shift, ok = runeToKey(' ')
	if !ok {
		t.Fatalf("runeToKey(' '): expected found")
	}
	if k != hid.KeyboardKeySpace {
		t.Errorf("runeToKey(' ') key = %v, want %v", k, hid.KeyboardKeySpace)
	}
	if shift {
		t.Errorf("runeToKey(' ') shift = true, want false")
	}

	if _, _, ok = runeToKey('€'); ok {
		t.Errorf("runeToKey('€'): expected not found")
	}
}

func TestComputeScale(t *testing.T) {
	if got := computeScale(1280, 720, 2560, 1440); got != (scaleXY{X: 2, Y: 2}) {
		t.Errorf("computeScale retina = %+v, want {2 2}", got)
	}
	if got := computeScale(800, 600, 800, 600); got != (scaleXY{X: 1, Y: 1}) {
		t.Errorf("computeScale identity = %+v, want {1 1}", got)
	}
	// Guards against division by zero when no renderer is present.
	if got := computeScale(0, 0, 100, 100); got != (scaleXY{X: 1, Y: 1}) {
		t.Errorf("computeScale zero = %+v, want {1 1}", got)
	}
}

func TestToLogicalFramebufferScale2(t *testing.T) {
	// Retina: a pixel read off a 2560x1440 screenshot maps to half its value in
	// the 1280x720 logical space the engine injects into.
	lx, ly, err := toLogical(spaceFramebuffer, 1000, 800, 1280, 720, 2560, 1440)
	if err != nil {
		t.Fatalf("toLogical returned error: %+v", err)
	}
	if math.Abs(float64(lx)-500.0) > 0.001 {
		t.Errorf("lx = %v, want 500", lx)
	}
	if math.Abs(float64(ly)-400.0) > 0.001 {
		t.Errorf("ly = %v, want 400", ly)
	}
}

func TestToLogicalScale1Identity(t *testing.T) {
	lx, ly, err := toLogical(spaceFramebuffer, 100, 50, 1280, 720, 1280, 720)
	if err != nil {
		t.Fatalf("toLogical returned error: %+v", err)
	}
	if lx != float32(100) {
		t.Errorf("lx = %v, want 100", lx)
	}
	if ly != float32(50) {
		t.Errorf("ly = %v, want 50", ly)
	}
}

func TestToLogicalWindowPassthrough(t *testing.T) {
	lx, ly, err := toLogical(spaceWindow, 100, 50, 1280, 720, 2560, 1440)
	if err != nil {
		t.Fatalf("toLogical returned error: %+v", err)
	}
	if lx != float32(100) {
		t.Errorf("lx = %v, want 100", lx)
	}
	if ly != float32(50) {
		t.Errorf("ly = %v, want 50", ly)
	}
}

func TestToLogicalOutOfBounds(t *testing.T) {
	_, _, err := toLogical(spaceFramebuffer, 5000, 10, 1280, 720, 2560, 1440)
	if err == nil {
		t.Fatalf("toLogical(x out of bounds): expected error")
	}
	if err.Code != codeInvalidCoordinate {
		t.Errorf("err.Code = %v, want %v", err.Code, codeInvalidCoordinate)
	}
	if err.Status != 400 {
		t.Errorf("err.Status = %v, want 400", err.Status)
	}

	if _, _, err = toLogical(spaceFramebuffer, -1, 10, 1280, 720, 2560, 1440); err == nil {
		t.Errorf("toLogical(negative x): expected error")
	}

	// x == window width is out of bounds (valid range is [0, width)).
	if _, _, err = toLogical(spaceWindow, 1280, 10, 1280, 720, 2560, 1440); err == nil {
		t.Errorf("toLogical(x == width): expected error")
	}
}

func TestValidateInput(t *testing.T) {
	valid := &inputRequest{Actions: []inputAction{
		{Type: actionKeyPress, Key: "Return"},
		{Type: actionTypeText, Text: "hi"},
		{Type: actionWaitFrames, Frames: 2},
		{Type: actionScroll, DX: 0, DY: -3},
	}}
	if err := validateInput(valid); err != nil {
		t.Errorf("validateInput(valid) = %+v, want nil", err)
	}

	x := float32(10)
	withCoords := &inputRequest{Actions: []inputAction{
		{Type: actionMouseMove, X: &x, Y: &x},
		{Type: actionMouseClick, Button: "left", X: &x, Y: &x},
	}}
	if err := validateInput(withCoords); err != nil {
		t.Errorf("validateInput(withCoords) = %+v, want nil", err)
	}
}

func TestValidateInputErrors(t *testing.T) {
	x := float32(10)
	cases := []struct {
		name string
		req  *inputRequest
		code string
	}{
		{"empty", &inputRequest{}, codeInvalidRequest},
		{"unknown-type", &inputRequest{Actions: []inputAction{{Type: "bogus"}}}, codeInvalidRequest},
		{"move-missing-xy", &inputRequest{Actions: []inputAction{{Type: actionMouseMove}}}, codeInvalidRequest},
		{"unknown-key", &inputRequest{Actions: []inputAction{{Type: actionKeyPress, Key: "nope"}}}, codeInvalidKey},
		{"unknown-button", &inputRequest{Actions: []inputAction{{Type: actionMouseClick, Button: "purple", X: &x, Y: &x}}}, codeInvalidButton},
		{"half-coords", &inputRequest{Actions: []inputAction{{Type: actionMouseClick, Button: "left", X: &x}}}, codeInvalidRequest},
		{"empty-text", &inputRequest{Actions: []inputAction{{Type: actionTypeText}}}, codeInvalidRequest},
		{"bad-space", &inputRequest{CoordinateSpace: "weird", Actions: []inputAction{{Type: actionKeyPress, Key: "a"}}}, codeInvalidRequest},
	}
	for _, c := range cases {
		err := validateInput(c.req)
		if err == nil {
			t.Fatalf("%s: validateInput expected error", c.name)
		}
		if err.Code != c.code {
			t.Errorf("%s: err.Code = %v, want %v", c.name, err.Code, c.code)
		}
		if err.Status != 400 {
			t.Errorf("%s: err.Status = %v, want 400", c.name, err.Status)
		}
	}
}

// newTestServer wires the routes against a host-less driver. Only the paths that
// never dereference the (nil) window are exercised here: health, help, the CSRF
// guard, and request-validation rejections (which return before enqueueing).
func newTestServer() *httptest.Server {
	d := newDriver(&engine.Host{})
	mux := http.NewServeMux()
	d.routes(mux)
	return httptest.NewServer(withGuards(mux))
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/health")
	if err != nil {
		t.Fatalf("GET /v1/health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var h healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if !h.OK {
		t.Errorf("h.OK = false, want true")
	}
	if h.Service != serviceName {
		t.Errorf("h.Service = %q, want %q", h.Service, serviceName)
	}
	if h.Version != apiVersion {
		t.Errorf("h.Version = %q, want %q", h.Version, apiVersion)
	}
}

func TestHelpEndpoint(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/help")
	if err != nil {
		t.Fatalf("GET /v1/help: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var h helpResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		t.Fatalf("decode help: %v", err)
	}
	if len(h.Endpoints) == 0 {
		t.Errorf("h.Endpoints is empty, want non-empty")
	}
}

func TestGuardRejectsBrowserOrigin(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/v1/health", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Origin", "http://evil.example")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestQuitRejectsGet(t *testing.T) {
	// The happy path drives host.Close and needs a live game loop, so it is
	// covered by the runtime smoke test. Here we just confirm the route is
	// registered as POST-only (a GET yields 405, not 404).
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/quit")
	if err != nil {
		t.Fatalf("GET /v1/quit: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestResizeRejectsGet(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/resize")
	if err != nil {
		t.Fatalf("GET /v1/resize: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestResizeValidation(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	// Invalid dimensions are rejected before reaching the (host-less) game loop.
	for _, body := range []string{
		`{"width":0,"height":600}`,
		`{"width":800,"height":-1}`,
		`{"width":99999,"height":600}`,
		`not json`,
	} {
		resp, err := http.Post(srv.URL+"/v1/resize", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("%s: POST /v1/resize: %v", body, err)
		}
		var env errorEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			resp.Body.Close()
			t.Fatalf("%s: decode envelope: %v", body, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want %d", body, resp.StatusCode, http.StatusBadRequest)
		}
		if env.Error == nil {
			t.Fatalf("%s: env.Error is nil", body)
		}
		if env.Error.Code != codeInvalidRequest {
			t.Errorf("%s: env.Error.Code = %v, want %v", body, env.Error.Code, codeInvalidRequest)
		}
	}
}

func TestInputValidationOverHTTP(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	cases := []struct {
		body string
		code string
	}{
		{`{"actions":[]}`, codeInvalidRequest},
		{`{"actions":[{"type":"bogus"}]}`, codeInvalidRequest},
		{`{"actions":[{"type":"mouse_move"}]}`, codeInvalidRequest},
		{`{"actions":[{"type":"key_press","key":"NoSuchKey"}]}`, codeInvalidKey},
		{`{"actions":[{"type":"mouse_click","button":"purple","x":1,"y":1}]}`, codeInvalidButton},
		{`not json`, codeInvalidRequest},
	}
	for _, c := range cases {
		resp, err := http.Post(srv.URL+"/v1/input", "application/json", strings.NewReader(c.body))
		if err != nil {
			t.Fatalf("%s: POST /v1/input: %v", c.body, err)
		}
		var env errorEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			resp.Body.Close()
			t.Fatalf("%s: decode envelope: %v", c.body, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want %d", c.body, resp.StatusCode, http.StatusBadRequest)
		}
		if env.Error == nil {
			t.Fatalf("%s: env.Error is nil", c.body)
		}
		if env.Error.Code != c.code {
			t.Errorf("%s: env.Error.Code = %v, want %v", c.body, env.Error.Code, c.code)
		}
	}
}
