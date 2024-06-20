/******************************************************************************/
/* stage.go                                                                   */
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
	"io"
	"kaiju/assets/asset_info"
	"kaiju/engine"
	"kaiju/filesystem"
	"kaiju/klib"
)

func SerializeEntity(stream io.Writer, entity *engine.Entity) error {
	err := entity.Serialize(stream)
	klib.BinaryWrite(stream, int32(len(entity.Children)))
	for i := 0; i < len(entity.Children) && err == nil; i++ {
		err = SerializeEntity(stream, entity.Children[i])
	}
	return err
}

func deserializeEntity(stream io.Reader, to *engine.Entity, host *engine.Host) error {
	err := to.Deserialize(stream, host)
	host.AddEntity(to)
	childCount := int32(0)
	klib.BinaryRead(stream, &childCount)
	for i := int32(0); i < childCount && err == nil; i++ {
		c := engine.NewEntity()
		c.SetParent(to)
		err = deserializeEntity(stream, c, host)
	}
	return err
}

func Load(adi asset_info.AssetDatabaseInfo, host *engine.Host) error {
	data, err := filesystem.ReadFile(adi.Path)
	if err != nil {
		return err
	}
	stream := bytes.NewBuffer(data)
	eCount := int32(0)
	klib.BinaryRead(stream, &eCount)
	entities := make([]*engine.Entity, 0, eCount)
	for i := int32(0); i < eCount && err == nil; i++ {
		e := engine.NewEntity()
		err = deserializeEntity(stream, e, host)
		entities = append(entities, e)
	}
	if err != nil {
		for i := 0; i < len(entities); i++ {
			entities[i].Destroy()
		}
	}
	return err
}
