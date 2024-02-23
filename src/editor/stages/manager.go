/******************************************************************************/
/* manager.go                                                                 */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package stages

import (
	"bytes"
	"kaiju/assets/asset_importer"
	"kaiju/assets/asset_info"
	"kaiju/editor/alert"
	"kaiju/editor/editor_config"
	"kaiju/engine"
	"kaiju/filesystem"
	"kaiju/klib"
	"os"
	"path/filepath"
)

type Manager struct {
	host     *engine.Host
	registry *asset_importer.ImportRegistry
	stage    string
}

func NewManager(host *engine.Host, registry *asset_importer.ImportRegistry) Manager {
	return Manager{
		host:     host,
		registry: registry,
	}
}

func (m *Manager) Save() error {
	if m.stage == "" {
		name := <-alert.NewInput("Stage Name", "Name of stage...", "", "Save", "Cancel")
		if name == "" {
			return nil
		}
		m.stage = filepath.Join("content/stages/", name+editor_config.FileExtensionStage)
	}
	stream := bytes.NewBuffer(make([]byte, 0))
	all := m.host.Entities()
	var err error = nil
	klib.BinaryWrite(stream, int32(len(all)))
	for i := 0; i < len(all) && err == nil; i++ {
		err = all[i].EditorSerialize(stream)
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

func (m *Manager) Load(adi asset_info.AssetDatabaseInfo, host *engine.Host) error {
	ok := <-alert.New("Save Changes", "You are changing stages, any unsaved changes will be lost. Are you sure you wish to continue?", "Yes", "No")
	if !ok {
		return nil
	}
	m.stage = adi.Path
	data, err := filesystem.ReadFile(m.stage)
	if err != nil {
		return err
	}
	stream := bytes.NewBuffer(data)
	eCount := int32(0)
	klib.BinaryRead(stream, &eCount)
	entities := make([]*engine.Entity, 0, eCount)
	for i := int32(0); i < eCount && err == nil; i++ {
		e := m.host.NewEntity()
		entities = append(entities, e)
		err = e.EditorDeserialize(stream, host)
	}
	if err != nil {
		for i := 0; i < len(entities); i++ {
			entities[i].Destroy()
		}
	}
	return err
}
