/******************************************************************************/
/* table_of_contents.go                                                       */
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

package table_of_contents

import (
	"bytes"
	"encoding/json"
	"kaiju/build"
	"kaiju/engine/encoding/pod"
)

type TableOfContents struct {
	Entries map[string]TableEntry
}

type TableEntry struct {
	Id   string
	Name string
}

func init() {
	pod.Register(TableOfContents{})
	pod.Register(TableEntry{})
}

func New() TableOfContents {
	return TableOfContents{
		Entries: make(map[string]TableEntry),
	}
}

func Deserialize(data []byte) (TableOfContents, error) {
	var toc TableOfContents
	var err error
	if build.Editor || build.Debug {
		err = json.Unmarshal(data, &toc)
	} else {
		r := bytes.NewReader(data)
		err = pod.NewDecoder(r).Decode(&toc)
	}
	return toc, err
}

func (t TableOfContents) Serialize() ([]byte, error) { return json.Marshal(t) }

func (t TableEntry) IsValid() bool { return t.Id != "" && t.Name != "" }

func (t *TableOfContents) Add(entry TableEntry) bool {
	if _, ok := t.Entries[entry.Name]; ok {
		return false
	}
	t.Entries[entry.Name] = entry
	return true
}

func (t *TableOfContents) Remove(key string) {
	delete(t.Entries, key)
}

func (t TableOfContents) SelectByName(name string) (TableEntry, bool) {
	e, ok := t.Entries[name]
	return e, ok
}

func (t TableOfContents) SelectById(id string) (TableEntry, bool) {
	for _, v := range t.Entries {
		if v.Id == id {
			return v, true
		}
	}
	return TableEntry{}, false
}
