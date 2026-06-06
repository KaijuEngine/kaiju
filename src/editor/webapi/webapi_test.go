/******************************************************************************/
/* webapi_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package webapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testEditor struct {
	called bool
}

type testEndpoint struct {
	routes []Route
}

func (e testEndpoint) Routes() []Route { return e.routes }

func (e testEndpoint) ServeEditorWebAPI(editor *testEditor, w http.ResponseWriter, r *http.Request) {
	editor.called = true
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func resetRegistryForTest(t *testing.T) {
	t.Helper()
	endpointRegistry.mu.Lock()
	defer endpointRegistry.mu.Unlock()
	endpointRegistry.entries = nil
	endpointRegistry.routes = map[routeKey]struct{}{}
}

func TestRegisterPrefixesRoutesWithVersion(t *testing.T) {
	resetRegistryForTest(t)
	err := Register[*testEditor](testEndpoint{routes: []Route{{
		Method:      http.MethodGet,
		Path:        "/ping",
		Description: "Ping test",
		Example:     "curl http://127.0.0.1:1337/v1/ping",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	_, _, routes := routesFor[*testEditor]()
	found := false
	for _, route := range routes {
		if route.Method == http.MethodGet && route.Path == "/v1/ping" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected /v1/ping to be registered, got %#v", routes)
	}
}

func TestRegisterRejectsDuplicateRoute(t *testing.T) {
	resetRegistryForTest(t)
	endpoint := testEndpoint{routes: []Route{{Method: http.MethodGet, Path: "/ping"}}}
	if err := Register[*testEditor](endpoint); err != nil {
		t.Fatal(err)
	}
	err := Register[*testEditor](endpoint)
	if !errors.Is(err, ErrDuplicateRoute) {
		t.Fatalf("expected duplicate route error, got %v", err)
	}
}

func TestRegisterRejectsReservedHelpPath(t *testing.T) {
	resetRegistryForTest(t)
	err := Register[*testEditor](testEndpoint{routes: []Route{{Method: http.MethodGet, Path: HelpPath}}})
	if !errors.Is(err, ErrInvalidRoute) {
		t.Fatalf("expected invalid route error, got %v", err)
	}
}

func TestServeHTTPRequiresBearerToken(t *testing.T) {
	resetRegistryForTest(t)
	if err := Register[*testEditor](testEndpoint{routes: []Route{{Method: http.MethodGet, Path: "/ping"}}}); err != nil {
		t.Fatal(err)
	}
	editor := &testEditor{}
	server := New(editor)
	server.apiKey = "secret"

	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("missing auth status = %d, want %d", res.Code, http.StatusUnauthorized)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("wrong auth status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
	if editor.called {
		t.Fatal("handler should not have been called")
	}
}

func TestServeHTTPRoutesAuthorizedRequest(t *testing.T) {
	resetRegistryForTest(t)
	if err := Register[*testEditor](testEndpoint{routes: []Route{{Method: http.MethodPost, Path: "/ping"}}}); err != nil {
		t.Fatal(err)
	}
	editor := &testEditor{}
	server := New(editor)
	server.apiKey = "secret"

	req := httptest.NewRequest(http.MethodPost, "/v1/ping", nil)
	req.Header.Set("Authorization", "Bearer secret")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if !editor.called {
		t.Fatal("handler should have been called")
	}
}

func TestServeHTTPReturnsMethodNotAllowedForKnownPath(t *testing.T) {
	resetRegistryForTest(t)
	if err := Register[*testEditor](testEndpoint{routes: []Route{{Method: http.MethodPost, Path: "/ping"}}}); err != nil {
		t.Fatal(err)
	}
	server := New(&testEditor{})
	server.apiKey = "secret"

	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	req.Header.Set("Authorization", "Bearer secret")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
	if res.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("Allow = %q, want %q", res.Header().Get("Allow"), http.MethodPost)
	}
}

func TestHelpIncludesEndpointMetadata(t *testing.T) {
	resetRegistryForTest(t)
	if err := Register[*testEditor](testEndpoint{routes: []Route{{
		Method:      http.MethodPost,
		Path:        "/ping",
		Description: "Ping test",
		Example:     `{"ok":true}`,
	}}}); err != nil {
		t.Fatal(err)
	}
	server := New(&testEditor{})
	server.apiKey = "secret"

	req := httptest.NewRequest(http.MethodGet, HelpPath, nil)
	req.Header.Set("Authorization", "Bearer secret")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var help HelpResponse
	if err := json.NewDecoder(res.Body).Decode(&help); err != nil {
		t.Fatal(err)
	}
	if help.Version != "v1" {
		t.Fatalf("version = %q, want v1", help.Version)
	}
	found := false
	for _, route := range help.Endpoints {
		if route.Method == http.MethodPost &&
			route.Path == "/v1/ping" &&
			route.Description == "Ping test" &&
			route.Example == `{"ok":true}` {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("help response missing registered route metadata: %#v", help.Endpoints)
	}
}
