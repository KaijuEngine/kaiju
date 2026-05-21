/******************************************************************************/
/* cursor.go                                                                  */
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

package ui

import (
	"log/slog"
	"path"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/rendering"
)

type Cursor UI

type CursorTheme struct {
	Name     string
	BasePath string
	Size     matrix.Vec2
	Textures map[windowing.CursorKind]string
	Hotspots map[windowing.CursorKind]matrix.Vec2
}

type cursorData struct {
	panelData
	theme    CursorTheme
	lastKind windowing.CursorKind
}

var KenneyCursorTheme = CursorTheme{
	Name:     "kenney",
	BasePath: "cursors/kenney",
	Size:     matrix.NewVec2(32, 32),
	Textures: map[windowing.CursorKind]string{
		windowing.CursorKindAuto:         "default.png",
		windowing.CursorKindDefault:      "default.png",
		windowing.CursorKindNone:         "none.png",
		windowing.CursorKindContextMenu:  "context_menu.png",
		windowing.CursorKindText:         "text.png",
		windowing.CursorKindVerticalText: "vertical_text.png",
		windowing.CursorKindPointer:      "pointer.png",
		windowing.CursorKindHelp:         "help.png",
		windowing.CursorKindWait:         "wait.png",
		windowing.CursorKindProgress:     "progress.png",
		windowing.CursorKindCrosshair:    "crosshair.png",
		windowing.CursorKindCell:         "cell.png",
		windowing.CursorKindAlias:        "alias.png",
		windowing.CursorKindCopy:         "copy.png",
		windowing.CursorKindMove:         "move.png",
		windowing.CursorKindNoDrop:       "no_drop.png",
		windowing.CursorKindNotAllowed:   "not_allowed.png",
		windowing.CursorKindGrab:         "grab.png",
		windowing.CursorKindGrabbing:     "grabbing.png",
		windowing.CursorKindResizeN:      "resize_n.png",
		windowing.CursorKindResizeE:      "resize_e.png",
		windowing.CursorKindResizeS:      "resize_s.png",
		windowing.CursorKindResizeW:      "resize_w.png",
		windowing.CursorKindResizeNE:     "resize_ne.png",
		windowing.CursorKindResizeNW:     "resize_nw.png",
		windowing.CursorKindResizeSE:     "resize_se.png",
		windowing.CursorKindResizeSW:     "resize_sw.png",
		windowing.CursorKindResizeNS:     "resize_ns.png",
		windowing.CursorKindResizeEW:     "resize_ew.png",
		windowing.CursorKindResizeNWSE:   "resize_nwse.png",
		windowing.CursorKindResizeNESW:   "resize_nesw.png",
		windowing.CursorKindResizeCol:    "resize_col.png",
		windowing.CursorKindResizeRow:    "resize_row.png",
		windowing.CursorKindResizeAll:    "resize_all.png",
		windowing.CursorKindZoomIn:       "zoom_in.png",
		windowing.CursorKindZoomOut:      "zoom_out.png",
	},
	Hotspots: map[windowing.CursorKind]matrix.Vec2{
		windowing.CursorKindText:         matrix.NewVec2(16, 16),
		windowing.CursorKindVerticalText: matrix.NewVec2(16, 16),
		windowing.CursorKindWait:         matrix.NewVec2(16, 16),
		windowing.CursorKindProgress:     matrix.NewVec2(16, 16),
		windowing.CursorKindCrosshair:    matrix.NewVec2(16, 16),
		windowing.CursorKindCell:         matrix.NewVec2(16, 16),
		windowing.CursorKindMove:         matrix.NewVec2(16, 16),
		windowing.CursorKindGrab:         matrix.NewVec2(16, 16),
		windowing.CursorKindGrabbing:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeN:      matrix.NewVec2(16, 16),
		windowing.CursorKindResizeE:      matrix.NewVec2(16, 16),
		windowing.CursorKindResizeS:      matrix.NewVec2(16, 16),
		windowing.CursorKindResizeW:      matrix.NewVec2(16, 16),
		windowing.CursorKindResizeNE:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeNW:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeSE:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeSW:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeNS:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeEW:     matrix.NewVec2(16, 16),
		windowing.CursorKindResizeNWSE:   matrix.NewVec2(16, 16),
		windowing.CursorKindResizeNESW:   matrix.NewVec2(16, 16),
		windowing.CursorKindResizeCol:    matrix.NewVec2(16, 16),
		windowing.CursorKindResizeRow:    matrix.NewVec2(16, 16),
		windowing.CursorKindResizeAll:    matrix.NewVec2(16, 16),
		windowing.CursorKindZoomIn:       matrix.NewVec2(16, 16),
		windowing.CursorKindZoomOut:      matrix.NewVec2(16, 16),
	},
}

func (u *UI) ToCursor() *Cursor                  { return (*Cursor)(u) }
func (c *Cursor) Base() *UI                      { return (*UI)(c) }
func (d *cursorData) innerPanelData() *panelData { return &d.panelData }

func (c *Cursor) CursorData() *cursorData {
	return c.Base().elmData.(*cursorData)
}

func (c *Cursor) Init(theme CursorTheme) {
	c.Base().elmData = &cursorData{
		theme:    theme,
		lastKind: windowing.CursorKind(-1),
	}

	p := c.Base().ToPanel()
	p.Init(nil, ElementTypeCursor)
	p.AllowClickThrough()

	if p.shaderData != nil {
		p.shaderData.BorderLen = matrix.Vec2Zero()
	}

	layout := c.Base().Layout()
	layout.SetPositioning(PositioningFixed)
	layout.SetZ(100)
	layout.Scale(theme.Size.X(), theme.Size.Y())
	c.setKind(windowing.CursorKindDefault)
}

func (c *Cursor) SetTheme(theme CursorTheme) {
	data := c.CursorData()
	data.theme = theme
	data.lastKind = windowing.CursorKind(-1)
	c.Base().Layout().Scale(theme.Size.X(), theme.Size.Y())
	c.setKind(c.Base().Host().Window.CursorKind())
}

func (c *Cursor) SyncWithWindow() {
	base := c.Base()
	host := base.Host()
	if host == nil {
		return
	}
	if !host.Window.UsesVirtualCursor() ||
		!host.Window.CursorVisible() ||
		host.Window.CursorKind() == windowing.CursorKindNone {
		base.Hide()
		return
	}
	base.Show()
}

func (theme CursorTheme) TexturePath(kind windowing.CursorKind) string {
	if kind == windowing.CursorKindAuto {
		kind = windowing.CursorKindDefault
	}
	if texture, ok := theme.Textures[kind]; ok {
		return path.Join(theme.BasePath, texture)
	}
	return path.Join(theme.BasePath, theme.Textures[windowing.CursorKindDefault])
}

func (theme CursorTheme) Hotspot(kind windowing.CursorKind) matrix.Vec2 {
	if hotspot, ok := theme.Hotspots[kind]; ok {
		return hotspot
	}
	return matrix.Vec2Zero()
}

func CursorThemeByName(name string) CursorTheme {
	switch name {
	case KenneyCursorTheme.Name:
		return KenneyCursorTheme
	default:
		return KenneyCursorTheme
	}
}

func (c *Cursor) setKind(kind windowing.CursorKind) {
	data := c.CursorData()
	texturePath := data.theme.TexturePath(kind)
	tex, err := c.Base().Host().TextureCache().Texture(texturePath, rendering.TextureFilterNearest)
	if err != nil {
		slog.Error("failed to load virtual cursor texture",
			"theme", data.theme.Name, "kind", kind, "texture", texturePath, "error", err)
		return
	}
	c.Base().ToPanel().SetBackground(tex)
	data.lastKind = kind
}

func (c *Cursor) update(deltaTime float64) {
	defer tracing.NewRegion("Cursor.update").End()

	base := c.Base()
	host := base.Host()
	if host == nil {
		return
	}

	if !host.Window.UsesVirtualCursor() ||
		!host.Window.CursorVisible() ||
		host.Window.CursorKind() == windowing.CursorKindNone {
		base.Hide()
		return
	}
	base.Show()

	kind := host.Window.CursorKind()
	data := c.CursorData()
	if data.lastKind != kind {
		c.setKind(kind)
	}

	pos := host.Window.Cursor.ScreenPosition()
	hotspot := data.theme.Hotspot(kind)
	base.Layout().SetOffset(pos.X()-hotspot.X(), pos.Y()-hotspot.Y())

	base.ToPanel().update(deltaTime)
}
