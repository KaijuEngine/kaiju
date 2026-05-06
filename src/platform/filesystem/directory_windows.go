//go:build windows && filedialog

/******************************************************************************/
/* directory_windows.go                                                       */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package filesystem

/*
#cgo windows LDFLAGS: -lshell32 -lole32 -luuid
#include <stdlib.h>
#include "directory.win32.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"kaijuengine.com/build"

	"golang.org/x/sys/windows"
)

type queuedDialogCallback struct {
	callback func(NativeDialogResult)
	result   NativeDialogResult
}

var windowsDialogRuntime = struct {
	mu        sync.Mutex
	nextID    atomic.Uint64
	active    map[uint64]struct{}
	callbacks []queuedDialogCallback
	closing   bool
	wg        sync.WaitGroup
}{
	active: make(map[uint64]struct{}),
}

func knownPaths() map[string]string {
	folders := map[string]*windows.KNOWNFOLDERID{
		"Desktop":   windows.FOLDERID_Desktop,
		"Documents": windows.FOLDERID_Documents,
		"Downloads": windows.FOLDERID_Downloads,
		"Music":     windows.FOLDERID_Music,
		"Pictures":  windows.FOLDERID_Pictures,
		"Videos":    windows.FOLDERID_Videos,
	}
	paths := make(map[string]string, len(folders))
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		str := fmt.Sprintf("%c:\\", r)
		s, err := os.Stat(str)
		if err != nil || !s.IsDir() {
			continue
		}
		// TODO:  Get the drive name
		paths[str] = str
	}
	for name, id := range folders {
		path, err := windows.KnownFolderPath(id, 0)
		if err != nil {
			fmt.Printf("%s: %v\n", name, err)
			continue
		}
		paths[name] = path
	}
	return paths
}

func imageDirectory() (string, error) {
	userFolder, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userFolder, "Pictures"), nil
}

func gameDirectory() (string, error) {
	appdata, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appdata, "../Local", build.CompanyDirName, build.Title.AsFilePathString()), nil
}

func openFileInTextEditor(path string) *exec.Cmd {
	return exec.Command("cmd.exe", "/C", "notepad", path)
}

func openFileBrowserCommand(path string) *exec.Cmd {
	return exec.Command("cmd.exe", "/C", "start", path)
}

func openFileBrowserSelectCommand(path string) *exec.Cmd {
	return exec.Command("explorer.exe", "/select,", path)
}

func openFileDialogWindow(startPath string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	if ok == nil {
		return errors.New("open file dialog ok callback cannot be nil")
	}
	request := NativeDialogRequest{
		Mode:             NativeDialogModeOpenFile,
		CurrentDirectory: startPath,
		Filters:          extensionsToDialogFilters(extensions),
		WindowHandle:     windowHandle,
	}
	return openNativeDialogWindow(request, makeSimpleDialogCallback(ok, cancel))
}

func openSaveFileDialogWindow(startPath string, fileName string, extensions []DialogExtension, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	if ok == nil {
		return errors.New("save file dialog ok callback cannot be nil")
	}
	request := NativeDialogRequest{
		Mode:             NativeDialogModeSaveFile,
		CurrentDirectory: startPath,
		FileName:         fileName,
		Filters:          extensionsToDialogFilters(extensions),
		WindowHandle:     windowHandle,
	}
	return openNativeDialogWindow(request, makeSimpleDialogCallback(ok, cancel))
}

func openFolderDialogWindow(startPath string, ok func(path string), cancel func(), windowHandle unsafe.Pointer) error {
	if ok == nil {
		return errors.New("open folder dialog ok callback cannot be nil")
	}
	request := NativeDialogRequest{
		Mode:             NativeDialogModeOpenFolder,
		CurrentDirectory: startPath,
		WindowHandle:     windowHandle,
	}
	return openNativeDialogWindow(request, makeSimpleDialogCallback(ok, cancel))
}

func openNativeDialogWindow(request NativeDialogRequest, callback func(NativeDialogResult)) error {
	if callback == nil {
		return errors.New("native dialog callback cannot be nil")
	}

	if err := validateNativeDialogRequest(request); err != nil {
		return err
	}

	windowsDialogRuntime.mu.Lock()
	if windowsDialogRuntime.closing {
		windowsDialogRuntime.mu.Unlock()
		return errors.New("native dialog runtime is shutting down")
	}
	id := windowsDialogRuntime.nextID.Add(1)
	windowsDialogRuntime.active[id] = struct{}{}
	windowsDialogRuntime.wg.Add(1)
	windowsDialogRuntime.mu.Unlock()

	go runNativeDialogRequest(id, request, callback)
	return nil
}

func validateNativeDialogRequest(request NativeDialogRequest) error {
	if request.Mode > NativeDialogModeOpenFolder {
		return fmt.Errorf("invalid native dialog mode: %d", request.Mode)
	}

	// The C bridge encodes filters/options with raw |, ;, and newline delimiters
	// Reject ambiguous fields here
	for i := range request.Filters {
		filter := request.Filters[i]
		if strings.ContainsAny(filter.Name, "|\r\n") {
			return fmt.Errorf("native dialog filter %d name contains an unsupported delimiter", i)
		}
		for j := range filter.Patterns {
			if strings.ContainsAny(filter.Patterns[j], "|;\r\n") {
				return fmt.Errorf("native dialog filter %d pattern %d contains an unsupported delimiter", i, j)
			}
		}
	}
	for i := range request.Options {
		option := request.Options[i]
		if strings.ContainsAny(option.Name, "|\r\n") {
			return fmt.Errorf("native dialog option %d name contains an unsupported delimiter", i)
		}
		for j := range option.Values {
			if strings.ContainsAny(option.Values[j], "|;\r\n") {
				return fmt.Errorf("native dialog option %d value %d contains an unsupported delimiter", i, j)
			}
		}
	}
	return nil
}

func runNativeDialogRequest(id uint64, request NativeDialogRequest, callback func(NativeDialogResult)) {
	defer windowsDialogRuntime.wg.Done()
	defer func() {
		windowsDialogRuntime.mu.Lock()
		delete(windowsDialogRuntime.active, id)
		windowsDialogRuntime.mu.Unlock()
	}()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	result := executeNativeDialog(request)
	if len(result.SelectedOptions) == 0 {
		result.SelectedOptions = defaultSelectedOptions(request.Options)
	}

	windowsDialogRuntime.mu.Lock()
	if !windowsDialogRuntime.closing {
		windowsDialogRuntime.callbacks = append(windowsDialogRuntime.callbacks, queuedDialogCallback{
			callback: callback,
			result:   result,
		})
	}
	windowsDialogRuntime.mu.Unlock()
}

func executeNativeDialog(request NativeDialogRequest) NativeDialogResult {
	var cReq C.dialog_request_t

	cReq.mode = C.int(request.Mode)
	cReq.show_hidden = 0
	if request.ShowHidden {
		cReq.show_hidden = 1
	}
	cReq.hwnd = request.WindowHandle

	cTitle := cStringOrNil(request.Title)
	cCurrentDir := cStringOrNil(request.CurrentDirectory)
	cFileName := cStringOrNil(request.FileName)
	cRoot := cStringOrNil(request.Root)
	cFilters := cStringOrNil(encodeDialogFilters(request.Filters))
	cOptions := cStringOrNil(encodeDialogOptions(request.Options))

	defer freeCString(cTitle)
	defer freeCString(cCurrentDir)
	defer freeCString(cFileName)
	defer freeCString(cRoot)
	defer freeCString(cFilters)
	defer freeCString(cOptions)

	cReq.title_utf8 = cTitle
	cReq.current_directory_utf8 = cCurrentDir
	cReq.filename_utf8 = cFileName
	cReq.root_utf8 = cRoot
	cReq.filters_utf8 = cFilters
	cReq.options_utf8 = cOptions

	cRes := C.run_native_file_dialog(&cReq)
	defer C.free_native_file_dialog_result(&cRes)

	result := NativeDialogResult{
		SelectedFilterIndex: int(cRes.selected_filter_index),
		HResult:             int32(cRes.hresult),
		Paths:               make([]string, 0, int(cRes.path_count)),
	}

	switch int(cRes.status) {
	case int(C.DIALOG_STATUS_OK):
		result.Status = NativeDialogStatusAccepted
	case int(C.DIALOG_STATUS_CANCEL):
		result.Status = NativeDialogStatusCancel
	default:
		result.Status = NativeDialogStatusFailed
	}

	if cRes.error_utf8 != nil {
		result.Err = errors.New(C.GoString(cRes.error_utf8))
	}
	if cRes.selected_options_utf8 != nil {
		result.SelectedOptions = parseDialogSelectedOptions(C.GoString(cRes.selected_options_utf8))
	}

	if cRes.path_count > 0 && cRes.paths_utf8 != nil {
		paths := unsafe.Slice((**C.char)(unsafe.Pointer(cRes.paths_utf8)), int(cRes.path_count))
		for i := range paths {
			if paths[i] != nil {
				result.Paths = append(result.Paths, C.GoString(paths[i]))
			}
		}
	}

	applyNativeDialogResultGuards(&result, request.Root)
	return result
}

func applyNativeDialogResultGuards(result *NativeDialogResult, root string) {
	if result == nil {
		return
	}

	if result.Status == NativeDialogStatusAccepted && strings.TrimSpace(root) != "" {
		for i := range result.Paths {
			if !isPathWithinDialogRoot(root, result.Paths[i]) {
				result.Status = NativeDialogStatusFailed
				result.Err = errors.New("selected path is outside configured dialog root")
				result.Paths = nil
				break
			}
		}
	}

	if result.Status == NativeDialogStatusAccepted && len(result.Paths) == 0 {
		result.Status = NativeDialogStatusCancel
	}

	if result.Status == NativeDialogStatusFailed && len(result.Paths) > 0 {
		// Error results should never expose partial selections to callers.
		result.Paths = nil
	}

	if result.Status == NativeDialogStatusFailed && result.Err == nil {
		result.Err = fmt.Errorf("native file dialog failed (HRESULT=0x%08X)", uint32(result.HResult))
	}
}

func isPathWithinDialogRoot(root, path string) bool {
	rootNorm, ok := normalizeDialogPathForContainment(root)
	if !ok {
		return false
	}
	pathNorm, ok := normalizeDialogPathForContainment(path)
	if !ok {
		return false
	}
	if !strings.HasPrefix(pathNorm, rootNorm) {
		return false
	}
	// Drive roots include the separator (e.g. C:\), so a prefix match already implies containment.
	if isDriveRootPath(rootNorm) {
		return true
	}
	if len(pathNorm) == len(rootNorm) {
		return true
	}
	next := pathNorm[len(rootNorm)]
	return next == '\\' || next == '/'
}

func normalizeDialogPathForContainment(path string) (string, bool) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", false
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", false
	}
	norm := filepath.Clean(abs)
	if !isDriveRootPath(norm) {
		norm = strings.TrimRight(norm, `\/`)
	}
	return strings.ToLower(norm), true
}

func isDriveRootPath(path string) bool {
	return len(path) == 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/')
}

func processDialogCallbacks() {
	windowsDialogRuntime.mu.Lock()
	pending := windowsDialogRuntime.callbacks
	windowsDialogRuntime.callbacks = nil
	windowsDialogRuntime.mu.Unlock()

	for i := range pending {
		func(item queuedDialogCallback) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("panic while executing native dialog callback", "panic", r)
				}
			}()
			item.callback(item.result)
		}(pending[i])
	}
}

func shutdownNativeDialogs() {
	windowsDialogRuntime.mu.Lock()
	if windowsDialogRuntime.closing {
		windowsDialogRuntime.mu.Unlock()
		return
	}
	windowsDialogRuntime.closing = true
	windowsDialogRuntime.callbacks = nil
	windowsDialogRuntime.mu.Unlock()

	done := make(chan struct{})
	go func() {
		windowsDialogRuntime.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		slog.Warn("timed out waiting for native dialog worker threads to finish")
	}
}

func makeSimpleDialogCallback(ok func(path string), cancel func()) func(NativeDialogResult) {
	return func(result NativeDialogResult) {
		if ok != nil && result.Status == NativeDialogStatusAccepted && len(result.Paths) > 0 {
			ok(result.Paths[0])
			return
		}
		if result.Status == NativeDialogStatusFailed {
			slog.Error("native dialog request failed", "error", result.Err, "hresult", fmt.Sprintf("0x%08X", uint32(result.HResult)))
		}
		if cancel != nil {
			cancel()
		}
	}
}

func extensionsToDialogFilters(extensions []DialogExtension) []DialogFilter {
	filters := make([]DialogFilter, 0, len(extensions)+1)
	for i := range extensions {
		e := extensions[i]
		name := strings.TrimSpace(e.Name)
		if name == "" {
			name = "Files"
		}
		pattern := normalizeDialogPattern(e.Extension)
		filters = append(filters, DialogFilter{Name: name, Patterns: []string{pattern}})
	}
	if len(filters) == 0 {
		filters = append(filters, DialogFilter{Name: "All Files (*.*)", Patterns: []string{"*.*"}})
	}
	return filters
}

func normalizeDialogPattern(extension string) string {
	ext := strings.TrimSpace(extension)
	if ext == "" || ext == ".*" || ext == "*" || ext == "*.*" {
		return "*.*"
	}
	if strings.Contains(ext, "*") {
		return ext
	}
	if strings.HasPrefix(ext, ".") {
		return "*" + ext
	}
	return "*." + ext
}

func encodeDialogFilters(filters []DialogFilter) string {
	if len(filters) == 0 {
		return ""
	}
	var b strings.Builder
	for i := range filters {
		f := filters[i]
		name := strings.TrimSpace(f.Name)
		if name == "" {
			name = "Files"
		}
		patterns := make([]string, 0, len(f.Patterns))
		for j := range f.Patterns {
			p := strings.TrimSpace(f.Patterns[j])
			if p != "" {
				patterns = append(patterns, p)
			}
		}
		if len(patterns) == 0 {
			patterns = []string{"*.*"}
		}
		b.WriteString(name)
		b.WriteString("|")
		b.WriteString(strings.Join(patterns, ";"))
		if i != len(filters)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func encodeDialogOptions(options []DialogCustomOption) string {
	if len(options) == 0 {
		return ""
	}
	var b strings.Builder
	first := true
	for i := range options {
		opt := options[i]
		name := strings.TrimSpace(opt.Name)
		if name == "" {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false
		b.WriteString(name)
		b.WriteString("|")
		b.WriteString(fmt.Sprintf("%d", opt.Default))
		b.WriteString("|")
		if len(opt.Values) > 0 {
			values := make([]string, 0, len(opt.Values))
			for j := range opt.Values {
				v := strings.TrimSpace(opt.Values[j])
				if v != "" {
					values = append(values, v)
				}
			}
			b.WriteString(strings.Join(values, ";"))
		}
	}
	return b.String()
}

func parseDialogSelectedOptions(encoded string) map[string]any {
	selected := map[string]any{}
	if strings.TrimSpace(encoded) == "" {
		return selected
	}
	lines := strings.Split(encoded, "\n")
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		kind := strings.TrimSpace(parts[1])
		value := strings.TrimSpace(parts[2])
		if name == "" {
			continue
		}
		switch kind {
		case "b":
			selected[name] = value == "1"
		case "i":
			idx, err := strconv.Atoi(value)
			if err == nil {
				selected[name] = idx
			}
		}
	}
	return selected
}

func defaultSelectedOptions(options []DialogCustomOption) map[string]any {
	selected := make(map[string]any, len(options))
	for i := range options {
		opt := options[i]
		name := strings.TrimSpace(opt.Name)
		if name == "" {
			continue
		}
		if len(opt.Values) == 0 {
			selected[name] = opt.Default != 0
			continue
		}
		idx := opt.Default
		if idx < 0 {
			idx = 0
		}
		if idx >= len(opt.Values) {
			idx = len(opt.Values) - 1
		}
		selected[name] = idx
	}
	return selected
}

func cStringOrNil(v string) *C.char {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return C.CString(v)
}

func freeCString(s *C.char) {
	if s != nil {
		C.free(unsafe.Pointer(s))
	}
}
