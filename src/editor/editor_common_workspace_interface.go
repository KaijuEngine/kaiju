/******************************************************************************/
/* editor_common_workspace_interface.go                                       */
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

package editor

import (
	"kaiju/editor/editor_events"
	"kaiju/editor/editor_overlay/reference_viewer"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"log/slog"
)

func (ed *Editor) Events() *editor_events.EditorEvents {
	return &ed.events
}

func (ed *Editor) History() *memento.History {
	return &ed.history
}

func (ed *Editor) Project() *project.Project {
	return &ed.project
}

func (ed *Editor) ProjectFileSystem() *project_file_system.FileSystem {
	return ed.project.FileSystem()
}

func (ed *Editor) Cache() *content_database.Cache {
	return ed.project.CacheDatabase()
}

func (ed *Editor) Settings() *editor_settings.Settings {
	return &ed.settings
}

func (ed *Editor) StageView() *editor_stage_view.StageView {
	return &ed.stageView
}

func (ed *Editor) ShowReferences(id string) {
	refs, err := ed.project.FindReferences(id)
	if err != nil {
		slog.Error("failed to read the references for the content", "id", id, "error", err)
		return
	}
	ed.BlurInterface()
	o, _ := reference_viewer.Show(ed.host, refs)
	o.OnClose.Add(ed.FocusInterface)
}
