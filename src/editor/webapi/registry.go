/******************************************************************************/
/* registry.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package webapi

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

const (
	VersionPrefix = "/v1"
	HelpPath      = "/help"
)

var (
	ErrDuplicateRoute = errors.New("webapi: duplicate route")
	ErrInvalidRoute   = errors.New("webapi: invalid route")
)

type Route struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	Example     string `json:"example,omitempty"`
}

type Endpoint[T any] interface {
	Routes() []Route
	ServeEditorWebAPI(editor T, w http.ResponseWriter, r *http.Request)
}

type routeKey struct {
	method string
	path   string
}

type routeEntry struct {
	route   Route
	handler any
}

var endpointRegistry = struct {
	mu      sync.RWMutex
	entries []routeEntry
	routes  map[routeKey]struct{}
}{
	routes: map[routeKey]struct{}{},
}

func Register[T any](handler Endpoint[T]) error {
	if any(handler) == nil {
		return fmt.Errorf("%w: handler is nil", ErrInvalidRoute)
	}
	routes := handler.Routes()
	if len(routes) == 0 {
		return fmt.Errorf("%w: handler has no routes", ErrInvalidRoute)
	}
	entries := make([]routeEntry, 0, len(routes))
	localRoutes := map[routeKey]struct{}{}
	for _, route := range routes {
		normalized, err := normalizeRoute(route)
		if err != nil {
			return err
		}
		key := routeKey{method: normalized.Method, path: normalized.Path}
		if _, exists := localRoutes[key]; exists {
			return fmt.Errorf("%w: %s %s", ErrDuplicateRoute, key.method, key.path)
		}
		localRoutes[key] = struct{}{}
		entries = append(entries, routeEntry{route: normalized, handler: handler})
	}

	endpointRegistry.mu.Lock()
	defer endpointRegistry.mu.Unlock()
	for _, entry := range entries {
		key := routeKey{method: entry.route.Method, path: entry.route.Path}
		if _, exists := endpointRegistry.routes[key]; exists {
			return fmt.Errorf("%w: %s %s", ErrDuplicateRoute, key.method, key.path)
		}
	}
	for _, entry := range entries {
		key := routeKey{method: entry.route.Method, path: entry.route.Path}
		endpointRegistry.routes[key] = struct{}{}
		endpointRegistry.entries = append(endpointRegistry.entries, entry)
	}
	return nil
}

func MustRegister[T any](handler Endpoint[T]) {
	if err := Register(handler); err != nil {
		panic(err)
	}
}

func normalizeRoute(route Route) (Route, error) {
	route.Method = strings.ToUpper(strings.TrimSpace(route.Method))
	route.Path = strings.TrimSpace(route.Path)
	route.Description = strings.TrimSpace(route.Description)
	route.Example = strings.TrimSpace(route.Example)
	if !validMethod(route.Method) {
		return Route{}, fmt.Errorf("%w: invalid method %q", ErrInvalidRoute, route.Method)
	}
	path, err := versionedPath(route.Path)
	if err != nil {
		return Route{}, err
	}
	route.Path = path
	return route, nil
}

func validMethod(method string) bool {
	if method == "" {
		return false
	}
	for _, r := range method {
		if r <= ' ' || r == '/' {
			return false
		}
	}
	return true
}

func versionedPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path is empty", ErrInvalidRoute)
	}
	if strings.ContainsAny(path, " \t\r\n?#") {
		return "", fmt.Errorf("%w: path %q contains invalid characters", ErrInvalidRoute, path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if path == HelpPath {
		return "", fmt.Errorf("%w: %s is reserved", ErrInvalidRoute, HelpPath)
	}
	if path == VersionPrefix {
		return "", fmt.Errorf("%w: route path must be under %s/", ErrInvalidRoute, VersionPrefix)
	}
	if strings.HasPrefix(path, VersionPrefix+"/") {
		return path, nil
	}
	return VersionPrefix + path, nil
}

func routesFor[T any]() (map[routeKey]Endpoint[T], map[string][]string, []Route) {
	endpointRegistry.mu.RLock()
	defer endpointRegistry.mu.RUnlock()
	routes := map[routeKey]Endpoint[T]{}
	methodsByPath := map[string][]string{}
	helpRoutes := make([]Route, 0, len(endpointRegistry.entries)+1)
	helpRoutes = append(helpRoutes, Route{
		Method:      http.MethodGet,
		Path:        HelpPath,
		Description: "Lists available editor Web API endpoints.",
		Example:     `curl -H "Authorization: Bearer <api-key>" http://127.0.0.1:1337/help`,
	})
	for _, entry := range endpointRegistry.entries {
		handler, ok := entry.handler.(Endpoint[T])
		if !ok {
			continue
		}
		key := routeKey{method: entry.route.Method, path: entry.route.Path}
		routes[key] = handler
		methodsByPath[entry.route.Path] = append(methodsByPath[entry.route.Path], entry.route.Method)
		helpRoutes = append(helpRoutes, entry.route)
	}
	for path := range methodsByPath {
		sort.Strings(methodsByPath[path])
	}
	sort.SliceStable(helpRoutes, func(i, j int) bool {
		if helpRoutes[i].Path == helpRoutes[j].Path {
			return helpRoutes[i].Method < helpRoutes[j].Method
		}
		return helpRoutes[i].Path < helpRoutes[j].Path
	})
	return routes, methodsByPath, helpRoutes
}
