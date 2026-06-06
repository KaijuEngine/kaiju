/******************************************************************************/
/* common_workspace.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package common_workspace

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

type CommonWorkspace struct {
	Host      *engine.Host
	Doc       *document.Document
	UiMan     ui.Manager
	IsBlurred bool
}

func (w *CommonWorkspace) InitializeWithUI(host *engine.Host, htmlPath string, withData any, funcMap map[string]func(*document.Element)) error {
	defer tracing.NewRegion("CommonWorkspace.InitializeWithUI").End()
	w.Host = host
	w.UiMan.Init(host)
	return w.ReloadUI(htmlPath, withData, funcMap)
}

func (w *CommonWorkspace) ReloadUI(htmlPath string, withData any, funcMap map[string]func(*document.Element)) error {
	if w.Doc != nil {
		w.Doc.Destroy()
		w.Doc = nil
	}
	var err error
	w.Doc, err = markup.DocumentFromHTMLAsset(&w.UiMan, htmlPath, withData, funcMap)
	if err == nil {
		w.Doc.Deactivate()
	}
	return err
}

func (w *CommonWorkspace) CommonOpen() {
	defer tracing.NewRegion("CommonWorkspace.CommonOpen").End()
	w.Doc.Activate()
	w.UiMan.EnableUpdate()
}

func (w *CommonWorkspace) CommonClose() {
	defer tracing.NewRegion("CommonWorkspace.CommonClose").End()
	w.UiMan.DisableUpdate()
	w.Doc.Deactivate()
}

func (w *CommonWorkspace) Focus() {
	defer tracing.NewRegion("CommonWorkspace.Focus").End()
	w.UiMan.EnableUpdate()
	w.IsBlurred = false
}

func (w *CommonWorkspace) Blur() {
	defer tracing.NewRegion("CommonWorkspace.Blur").End()
	w.UiMan.DisableUpdate()
	w.IsBlurred = true
}

func (w *CommonWorkspace) IsFocusedOnInput() bool {
	return w.UiMan.Group.IsFocusedOnInput()
}

// CommonShutdown tears down the workspace's UI document AND its UI manager.
// Called by a workspace's Shutdown() implementation when the editor disables
// the workspace at runtime. Embedding workspaces should drop any event
// subscriptions before calling this.
//
// The UiMan.Shutdown call is critical: without it, a subsequent re-init
// (when the workspace is re-enabled) would call UiMan.Init a second time,
// adding a second host.UIUpdater callback for the same Manager. Two
// concurrent updates would then race on the same Manager's iteration
// slices and panic with index-out-of-range.
func (w *CommonWorkspace) CommonShutdown() {
	defer tracing.NewRegion("CommonWorkspace.CommonShutdown").End()
	if w.Doc != nil {
		w.Doc.Destroy()
		w.Doc = nil
	}
	w.UiMan.Shutdown()
}

func (w *CommonWorkspace) Update(float64) {}
