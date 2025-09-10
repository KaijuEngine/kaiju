/******************************************************************************/
/* importer.go                                                                */
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

package asset_importer

import (
	"errors"
	"kaiju/engine/assets/asset_info"
	"path/filepath"

	"github.com/KaijuEngine/uuid"
)

// Importer is an interface for importing assets into the asset database.
// It is used to allow for custom importers to be added to the system
// to handle new asset types.
// Handles should return true if the importer can handle the given path.
// Import should import the asset into the asset database.
type Importer interface {
	Handles(path string) bool
	Import(path string) error
}

type ImportRegistry struct {
	importers []Importer
}

func NewImportRegistry() ImportRegistry {
	return ImportRegistry{
		importers: make([]Importer, 0),
	}
}

func (r *ImportRegistry) Register(importer Importer) {
	r.importers = append(r.importers, importer)
}

func (r *ImportRegistry) ImportIfNew(path string) error {
	if filepath.Ext(path) == asset_info.InfoExtension {
		return nil
	}
	if !asset_info.Exists(path) {
		return r.Import(path)
	}
	return nil
}

func (r *ImportRegistry) Import(path string) error {
	if filepath.Ext(path) == asset_info.InfoExtension {
		return nil
	}
	// We go back to front so devs can override default importers
	for i := len(r.importers) - 1; i >= 0; i-- {
		if r.importers[i].Handles(path) {
			return r.importers[i].Import(path)
		}
	}
	return ErrNoImporter
}

func (r *ImportRegistry) ImportUsingDefault(path string) error {
	for i := range r.importers {
		if r.importers[i].Handles(path) {
			return r.importers[i].Import(path)
		}
	}
	return ErrNoImporter
}

func createADI(importer Importer, path string, cleanup func(adi asset_info.AssetDatabaseInfo)) (asset_info.AssetDatabaseInfo, error) {
	adi, err := asset_info.Read(path)
	if errors.Is(err, asset_info.ErrNoInfo) {
		adi = asset_info.New(path, uuid.New().String())
		err = nil
	} else if err == nil && cleanup != nil {
		cleanup(adi)
	}
	return adi, err
}
