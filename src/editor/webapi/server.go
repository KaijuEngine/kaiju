/******************************************************************************/
/* server.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package webapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultPort = 1337
	bindHost    = "127.0.0.1"
)

var ErrMissingAPIKey = errors.New("webapi: missing API key")

type Config struct {
	Enabled bool
	Port    int32
	APIKey  string
}

type Server[T any] struct {
	editor T
	mu     sync.RWMutex
	server *http.Server
	port   int32
	apiKey string
}

type HelpResponse struct {
	Version   string  `json:"version"`
	Endpoints []Route `json:"endpoints"`
}

func New[T any](editor T) *Server[T] {
	return &Server[T]{editor: editor}
}

func (s *Server[T]) Apply(config Config) error {
	config = normalizeConfig(config)
	if !config.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return s.Close(ctx)
	}
	if config.APIKey == "" {
		return ErrMissingAPIKey
	}

	addr := net.JoinHostPort(bindHost, strconv.Itoa(int(config.Port)))
	s.mu.Lock()
	if s.server != nil && s.port == config.Port {
		s.apiKey = config.APIKey
		s.mu.Unlock()
		return nil
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	oldServer := s.server
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.server = httpServer
	s.port = config.Port
	s.apiKey = config.APIKey
	s.mu.Unlock()

	if oldServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := oldServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to stop previous editor Web API server", "error", err)
		}
		cancel()
	}

	go func() {
		slog.Info("editor Web API server started", "address", addr)
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("editor Web API server stopped unexpectedly", "error", err)
		}
	}()
	return nil
}

func (s *Server[T]) Close(ctx context.Context) error {
	s.mu.Lock()
	httpServer := s.server
	s.server = nil
	s.port = 0
	s.apiKey = ""
	s.mu.Unlock()
	if httpServer == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	slog.Info("stopping editor Web API server")
	if err := httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="kaiju-editor-webapi"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if r.URL.Path == HelpPath {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, []string{http.MethodGet})
			return
		}
		_, _, routes := routesFor[T]()
		writeJSON(w, http.StatusOK, HelpResponse{
			Version:   strings.TrimPrefix(VersionPrefix, "/"),
			Endpoints: routes,
		})
		return
	}

	routes, methodsByPath, _ := routesFor[T]()
	key := routeKey{method: r.Method, path: r.URL.Path}
	if handler, ok := routes[key]; ok {
		handler.ServeEditorWebAPI(s.editor, w, r)
		return
	}
	if methods, ok := methodsByPath[r.URL.Path]; ok {
		methodNotAllowed(w, methods)
		return
	}
	http.NotFound(w, r)
}

func normalizeConfig(config Config) Config {
	config.APIKey = strings.TrimSpace(config.APIKey)
	if config.Port <= 0 || config.Port > 65535 {
		config.Port = DefaultPort
	}
	return config
}

func (s *Server[T]) authorized(r *http.Request) bool {
	fields := strings.Fields(r.Header.Get("Authorization"))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return false
	}
	s.mu.RLock()
	apiKey := s.apiKey
	s.mu.RUnlock()
	if apiKey == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(fields[1]), []byte(apiKey)) == 1
}

func methodNotAllowed(w http.ResponseWriter, methods []string) {
	methods = append([]string(nil), methods...)
	sort.Strings(methods)
	w.Header().Set("Allow", strings.Join(methods, ", "))
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode editor Web API response", "error", err)
	}
}

func Address(port int32) string {
	if port <= 0 || port > 65535 {
		port = DefaultPort
	}
	return fmt.Sprintf("http://%s:%d", bindHost, port)
}
