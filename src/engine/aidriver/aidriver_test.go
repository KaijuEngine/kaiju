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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.True(t, ok, name)
		assert.Equal(t, want, got, name)
	}
	_, ok := lookupKey("definitely-not-a-key")
	assert.False(t, ok)
}

func TestLookupButton(t *testing.T) {
	b, ok := lookupButton("LEFT")
	require.True(t, ok)
	assert.Equal(t, hid.MouseButtonLeft, b)
	b, ok = lookupButton("right")
	require.True(t, ok)
	assert.Equal(t, hid.MouseButtonRight, b)
	_, ok = lookupButton("scrollwheel")
	assert.False(t, ok)
}

func TestRuneToKey(t *testing.T) {
	k, shift, ok := runeToKey('a')
	require.True(t, ok)
	assert.Equal(t, hid.KeyboardKeyA, k)
	assert.False(t, shift)

	k, shift, ok = runeToKey('A')
	require.True(t, ok)
	assert.Equal(t, hid.KeyboardKeyA, k)
	assert.True(t, shift)

	k, shift, ok = runeToKey('!')
	require.True(t, ok)
	assert.Equal(t, hid.KeyboardKey1, k)
	assert.True(t, shift)

	k, shift, ok = runeToKey(' ')
	require.True(t, ok)
	assert.Equal(t, hid.KeyboardKeySpace, k)
	assert.False(t, shift)

	_, _, ok = runeToKey('€')
	assert.False(t, ok)
}

func TestComputeScale(t *testing.T) {
	assert.Equal(t, scaleXY{X: 2, Y: 2}, computeScale(1280, 720, 2560, 1440))
	assert.Equal(t, scaleXY{X: 1, Y: 1}, computeScale(800, 600, 800, 600))
	// Guards against division by zero when no renderer is present.
	assert.Equal(t, scaleXY{X: 1, Y: 1}, computeScale(0, 0, 100, 100))
}

func TestToLogicalFramebufferScale2(t *testing.T) {
	// Retina: a pixel read off a 2560x1440 screenshot maps to half its value in
	// the 1280x720 logical space the engine injects into.
	lx, ly, err := toLogical(spaceFramebuffer, 1000, 800, 1280, 720, 2560, 1440)
	require.Nil(t, err)
	assert.InDelta(t, 500.0, lx, 0.001)
	assert.InDelta(t, 400.0, ly, 0.001)
}

func TestToLogicalScale1Identity(t *testing.T) {
	lx, ly, err := toLogical(spaceFramebuffer, 100, 50, 1280, 720, 1280, 720)
	require.Nil(t, err)
	assert.Equal(t, float32(100), lx)
	assert.Equal(t, float32(50), ly)
}

func TestToLogicalWindowPassthrough(t *testing.T) {
	lx, ly, err := toLogical(spaceWindow, 100, 50, 1280, 720, 2560, 1440)
	require.Nil(t, err)
	assert.Equal(t, float32(100), lx)
	assert.Equal(t, float32(50), ly)
}

func TestToLogicalOutOfBounds(t *testing.T) {
	_, _, err := toLogical(spaceFramebuffer, 5000, 10, 1280, 720, 2560, 1440)
	require.NotNil(t, err)
	assert.Equal(t, codeInvalidCoordinate, err.Code)
	assert.Equal(t, 400, err.Status)

	_, _, err = toLogical(spaceFramebuffer, -1, 10, 1280, 720, 2560, 1440)
	require.NotNil(t, err)

	// x == window width is out of bounds (valid range is [0, width)).
	_, _, err = toLogical(spaceWindow, 1280, 10, 1280, 720, 2560, 1440)
	require.NotNil(t, err)
}

func TestValidateInput(t *testing.T) {
	valid := &inputRequest{Actions: []inputAction{
		{Type: actionKeyPress, Key: "Return"},
		{Type: actionTypeText, Text: "hi"},
		{Type: actionWaitFrames, Frames: 2},
		{Type: actionScroll, DX: 0, DY: -3},
	}}
	assert.Nil(t, validateInput(valid))

	x := float32(10)
	withCoords := &inputRequest{Actions: []inputAction{
		{Type: actionMouseMove, X: &x, Y: &x},
		{Type: actionMouseClick, Button: "left", X: &x, Y: &x},
	}}
	assert.Nil(t, validateInput(withCoords))
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
		require.NotNil(t, err, c.name)
		assert.Equal(t, c.code, err.Code, c.name)
		assert.Equal(t, 400, err.Status, c.name)
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
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var h healthResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&h))
	assert.True(t, h.OK)
	assert.Equal(t, serviceName, h.Service)
	assert.Equal(t, apiVersion, h.Version)
}

func TestHelpEndpoint(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/help")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var h helpResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&h))
	assert.NotEmpty(t, h.Endpoints)
}

func TestGuardRejectsBrowserOrigin(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/v1/health", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://evil.example")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestQuitRejectsGet(t *testing.T) {
	// The happy path drives host.Close and needs a live game loop, so it is
	// covered by the runtime smoke test. Here we just confirm the route is
	// registered as POST-only (a GET yields 405, not 404).
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/quit")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestResizeRejectsGet(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/resize")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
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
		require.NoError(t, err, body)
		var env errorEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
		resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, body)
		require.NotNil(t, env.Error, body)
		assert.Equal(t, codeInvalidRequest, env.Error.Code, body)
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
		require.NoError(t, err, c.body)
		var env errorEnvelope
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
		resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, c.body)
		require.NotNil(t, env.Error, c.body)
		assert.Equal(t, c.code, env.Error.Code, c.body)
	}
}
