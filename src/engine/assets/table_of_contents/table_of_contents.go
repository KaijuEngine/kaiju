/******************************************************************************/
/* table_of_contents.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package table_of_contents

import (
	"bytes"
	"encoding/json"

	"kaijuengine.com/build"
	"kaijuengine.com/engine/encoding/pod"
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
