/******************************************************************************/
/* editor.go                                                                  */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package interfaces

import (
	"kaiju/assets/asset_importer"
	"kaiju/collision"
	"kaiju/editor/codegen"
	"kaiju/editor/memento"
	"kaiju/editor/selection"
	"kaiju/editor/stages"
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/ui/editor_window"
	"kaiju/editor/ui/hierarchy"
	"kaiju/editor/ui/status_bar"
	"kaiju/editor/viewport/controls"
	"kaiju/engine"
	"kaiju/host_container"
)

type Editor interface {
	Container() *host_container.Container
	Host() *engine.Host
	Camera() *controls.EditorCamera
	StageManager() *stages.Manager
	Selection() *selection.Selection
	History() *memento.History
	WindowListing() *editor_window.Listing
	StatusBar() *status_bar.StatusBar
	Hierarchy() *hierarchy.Hierarchy
	ContextMenu() *context_menu.ContextMenu
	ImportRegistry() *asset_importer.ImportRegistry
	OpenProject()
	AvailableDataBindings() []codegen.GeneratedType
	ReloadEntityDataListing()
	CreateEntity(name string) *engine.Entity
	// TODO:  BVH stuff can be encapsulated into another structure
	BVH() *collision.BVH
	BVHEntityUpdates(entities ...*engine.Entity)
	IsMouseOverViewport() bool
}
