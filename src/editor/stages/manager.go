/******************************************************************************/
/* manager.go                                                                 */
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

package stages

import (
	"bytes"
	"errors"
	"kaiju/engine/assets/asset_importer"
	"kaiju/engine/assets/asset_info"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/editor/memento"
	"kaiju/editor/ui/status_bar"
	"kaiju/engine"
	"kaiju/platform/filesystem"
	"kaiju/klib"
	"kaiju/engine/systems/stages"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	ErrorSaveCancelled = errors.New("save was requested then cancelled")
)

type Manager struct {
	host     *engine.Host
	registry *asset_importer.ImportRegistry
	history  *memento.History
	stage    string
}

func NewManager(host *engine.Host, registry *asset_importer.ImportRegistry,
	history *memento.History) Manager {
	return Manager{
		host:     host,
		registry: registry,
		history:  history,
	}
}

func (m *Manager) StageName() string { return m.stage }

func (m *Manager) confirmCheck() bool {
	return <-alert.New("Save Changes",
		"You are changing stages, any unsaved changes will be lost. Are you sure you wish to continue?",
		"Yes", "No", m.host)
}

func (m *Manager) New() {
	if !m.confirmCheck() {
		return
	}
	m.stage = ""
	m.history.Clear()
	m.host.ClearEntities()
}

func (m *Manager) saveInternal() error {
	if m.stage == "" {
		name := <-alert.NewInput("Stage Name", "Name of stage...",
			"", "Save", "Cancel", m.host)
		if name == "" {
			return ErrorSaveCancelled
		}
		path := filepath.Join("content/stages/", name+editor_config.FileExtensionStage)
		if _, err := os.Stat(path); err == nil {
			ok := <-alert.New("Overwrite stage?",
				"The stage "+path+" already exists. Would you like to overwrite it?",
				"Yes", "No", m.host)
			if !ok {
				return ErrorSaveCancelled
			}
		}
		m.stage = path
	}
	stream := bytes.NewBuffer(make([]byte, 0))
	all := m.host.Entities()
	roots := make([]*engine.Entity, 0, len(all))
	for i := 0; i < len(all); i++ {
		if all[i].IsRoot() {
			roots = append(roots, all[i])
		}
	}
	var err error = nil
	klib.BinaryWrite(stream, int32(len(roots)))
	for i := 0; i < len(roots) && err == nil; i++ {
		err = stages.SerializeEntity(stream, roots[i])
	}
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(m.stage), os.ModePerm)
	if err = filesystem.WriteFile(m.stage, stream.Bytes()); err != nil {
		return err
	}
	m.registry.ImportIfNew(m.stage)
	return nil
}

func (m *Manager) Save(statusBar *status_bar.StatusBar) error {
	err := m.saveInternal()
	if err == nil {
		if statusBar != nil {
			statusBar.SetMessage("Stage saved")
		}
	} else if errors.Is(err, ErrorSaveCancelled) && statusBar != nil {
		statusBar.SetMessage("Stage save cancelled")
	} else {
		slog.Error("Save stage failed", slog.String("error", err.Error()))
	}
	return err
}

func (m *Manager) Load(adi asset_info.AssetDatabaseInfo, host *engine.Host) error {
	if !m.confirmCheck() {
		return nil
	}
	m.history.Clear()
	m.host.ClearEntities()
	m.stage = adi.Path
	return stages.Load(adi, host)
}
