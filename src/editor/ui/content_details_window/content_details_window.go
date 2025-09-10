/******************************************************************************/
/* content_details_window.go                                                  */
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

package content_details_window

import (
	"kaiju/editor/editor_interface"
	"kaiju/editor/ui/details_common"
	"kaiju/engine/assets/asset_info"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"log/slog"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

const sizeConfig = "detailsWindowSize"

type ContentDetails struct {
	editor   editor_interface.Editor
	doc      *document.Document
	adis     []asset_info.AssetDatabaseInfo
	viewData contentDetailsData
}

type contentDetailsData struct {
	Name   string
	Count  int
	Fields []contentDataField
}

type contentDataField struct {
	Name    string
	Type    string
	Field   reflect.Value
	Options []string
}

func New(editor editor_interface.Editor) *ContentDetails {
	d := &ContentDetails{editor: editor}
	d.editor.Events().OnContentSelect.Event.Add(d.contentSelected)
	return d
}

func (d *ContentDetails) TabTitle() string             { return "Content Details" }
func (d *ContentDetails) Document() *document.Document { return d.doc }

func (d *ContentDetails) Destroy() {
	if d.doc != nil {
		d.doc.Destroy()
		d.doc = nil
	}
}

func (d *ContentDetails) SetADIs(adis []asset_info.AssetDatabaseInfo) {
	if len(d.adis) > 0 {
		d.saveAdis()
	}
	d.adis = adis
	d.viewData.Count = len(adis)
	if d.viewData.Count == 1 {
		a := &adis[0]
		d.viewData.Name = strings.TrimSuffix(filepath.Base(a.Path), filepath.Ext(a.Path))
		d.viewData.Fields = pullADIFields(a)
	}
	d.editor.ReloadTabs(d.TabTitle())
}

func pullADIFields(adi *asset_info.AssetDatabaseInfo) []contentDataField {
	structure := adi.MetadataStructure()
	if structure == nil {
		return []contentDataField{}
	}
	v := reflect.ValueOf(structure).Elem()
	t := v.Type()
	fields := make([]contentDataField, 0, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		vf := v.Field(i)
		field := contentDataField{
			Name:  f.Name,
			Type:  f.Type.Name(),
			Field: vf,
		}
		if op, ok := f.Tag.Lookup("options"); ok && op != "" {
			if v, ok := asset_info.ImageMetaOptions[op]; ok {
				keys := reflect.ValueOf(v).MapKeys()
				field.Options = make([]string, len(keys))
				for i := range keys {
					field.Options[i] = keys[i].String()
				}
				slices.Sort(field.Options)
			} else {
				slog.Error("failed to load the content metadata options for key", "key", op)
			}
		}
		fields = append(fields, field)
	}
	return fields
}

func (d *ContentDetails) Reload(uiMan *ui.Manager, root *document.Element) {
	if d.doc != nil {
		d.doc.Destroy()
	}
	host := d.editor.Host()
	host.CreatingEditorEntities()
	d.doc = klib.MustReturn(markup.DocumentFromHTMLAssetRooted(
		uiMan, "editor/ui/content_details_window/content_details_window.html", d.viewData,
		map[string]func(*document.Element){
			"changeData": d.changeData,
			"save":       d.save,
		}, root))
	host.DoneCreatingEditorEntities()
	d.doc.Clean()
}

func (d *ContentDetails) contentSelected() {
	paths := d.editor.Events().OnContentSelect.Content
	adis := []asset_info.AssetDatabaseInfo{}
	for i := range paths {
		a, err := asset_info.Read(paths[i])
		if err != nil {
			slog.Warn("failed to open the asset database info for file", "path", paths[i], "error", err)
			continue
		}
		adis = append(adis, a)
	}
	d.SetADIs(adis)
}

func (f *contentDataField) setReflectValue(strVal string) {
	v := f.Field
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(details_common.ToInt(strVal))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(details_common.ToUint(strVal))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(details_common.ToFloat(strVal))
	case reflect.String:
		v.SetString(strVal)
	case reflect.Bool:
		if strings.ToLower(strVal) == "false" || strVal == "0" {
			v.SetBool(false)
		} else {
			v.SetBool(true)
		}
	}
}

func (d *ContentDetails) changeData(elm *document.Element) {
	id := elm.Attribute("id")
	var field *contentDataField
	var v reflect.Value
	for i := range d.viewData.Fields {
		if d.viewData.Fields[i].Name == id {
			field = &d.viewData.Fields[i]
			v = field.Field
			break
		}
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(details_common.ToInt(elm.UI.ToInput().Text()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(details_common.ToUint(elm.UI.ToInput().Text()))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(details_common.ToFloat(elm.UI.ToInput().Text()))
	case reflect.String:
		if len(field.Options) > 0 {
			v.SetString(details_common.SelectString(elm))
		} else {
			v.SetString(details_common.InputString(elm))
		}
	case reflect.Bool:
		v.SetBool(elm.UI.ToCheckbox().IsChecked())
	}
}

func (d *ContentDetails) save(*document.Element) {
	d.saveAdis()
}

func (d *ContentDetails) saveAdis() {
	for i := range d.adis {
		if err := asset_info.Write(d.adis[i]); err != nil {
			slog.Error("failed to update the asset database info",
				"asset", d.adis[i].Path, "error", err)
		}
	}
}

// TODO:  This was copied from data_input_reflections.go, should turn both into common function
func (d contentDetailsData) PascalToTitle(str string) string {
	re := regexp.MustCompile("([A-Z])")
	result := re.ReplaceAllString(str, " $1")
	return strings.TrimSpace(result)
}

func (f *contentDataField) IsInput() bool    { return details_common.IsInput(f.Type) }
func (f *contentDataField) IsCheckbox() bool { return details_common.IsCheckbox(f.Type) }
