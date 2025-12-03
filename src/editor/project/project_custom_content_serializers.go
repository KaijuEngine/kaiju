/******************************************************************************/
/* project_custom_content_serializers.go                                      */
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

package project

import (
	"bytes"
	"encoding/json"
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/runtime/encoding/gob"
	"kaiju/stages"
	"log/slog"
)

func (p *Project) initializeCustomSerializers() {
	p.contentSerializers = make(map[string]func([]byte) ([]byte, error))
	p.contentSerializers[(content_database.Stage{}).TypeName()] = p.stageArchiveSerializer
	toc := content_database.TableOfContents{}
	p.contentSerializers[toc.TypeName()] = toc.ArchiveSerializer
}

func (p *Project) stageArchiveSerializer(rawData []byte) ([]byte, error) {
	var ss stages.StageJson
	if err := json.Unmarshal(rawData, &ss); err != nil {
		return rawData, err
	}
	s := stages.Stage{}
	s.FromMinimized(ss)
	var removeUnpackedDataBindings func(desc *stages.EntityDescription)
	removeUnpackedDataBindings = func(desc *stages.EntityDescription) {
		for i := range desc.DataBinding {
			g, ok := p.EntityDataBinding(desc.DataBinding[i].RegistraionKey)
			if ok {
				de := entity_data_binding.EntityDataEntry{}
				de.ReadEntityDataBindingType(g)
				for k, v := range desc.DataBinding[i].Fields {
					de.SetFieldByName(k, v)
				}
				desc.RawDataBinding = append(desc.RawDataBinding, de.BoundData)
			} else {
				slog.Warn("failed to locate the data binding for registration key",
					"key", desc.DataBinding[i].RegistraionKey)
			}
		}
		desc.DataBinding = desc.DataBinding[0:]
		for i := range desc.Children {
			removeUnpackedDataBindings(&desc.Children[i])
		}
	}
	for i := range s.Entities {
		removeUnpackedDataBindings(&s.Entities[i])
	}
	stream := bytes.NewBuffer(rawData)
	stream.Reset()
	err := gob.NewEncoder(stream).Encode(s)
	return stream.Bytes(), err
}
