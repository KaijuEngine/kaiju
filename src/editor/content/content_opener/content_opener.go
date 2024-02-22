/******************************************************************************/
/* content_opener.go                                                          */
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

package content_opener

import (
	"errors"
	"kaiju/assets/asset_importer"
	"kaiju/assets/asset_info"
	"kaiju/editor/memento"
	"kaiju/host_container"
)

var (
	ErrNoOpener = errors.New("no opener found")
)

type ContentOpener interface {
	Handles(adi asset_info.AssetDatabaseInfo) bool
	Open(adi asset_info.AssetDatabaseInfo,
		container *host_container.Container, history *memento.History) error
}

type Opener struct {
	openers   []ContentOpener
	container *host_container.Container
	importer  *asset_importer.ImportRegistry
	history   *memento.History
}

func New(importer *asset_importer.ImportRegistry,
	container *host_container.Container, history *memento.History) Opener {
	return Opener{
		importer:  importer,
		container: container,
		history:   history,
	}
}

func (o *Opener) Register(opener ContentOpener) {
	o.openers = append(o.openers, opener)
}

func (o *Opener) Open(adi asset_info.AssetDatabaseInfo) error {
	for i := range o.openers {
		if o.openers[i].Handles(adi) {
			return o.openers[i].Open(adi, o.container, o.history)
		}
	}
	return ErrNoOpener
}

func (o *Opener) OpenPath(path string) error {
	if !asset_info.Exists(path) {
		if err := o.importer.Import(path); err != nil {
			return err
		}
	}
	if adi, err := asset_info.Read(path); err != nil {
		return err
	} else {
		return o.Open(adi)
	}
}
