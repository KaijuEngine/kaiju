/******************************************************************************/
/* common_workspace.go                                                        */
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

package common_workspace

import (
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
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

func (w *CommonWorkspace) Update(float64) {}
