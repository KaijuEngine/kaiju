/******************************************************************************/
/* sprite_group.go                                                            */
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

package sprite

import (
	"kaiju/debug"
	"kaiju/engine"
	"kaiju/klib"
	"slices"
	"weak"
)

type SpriteGroupId = int
type IndexedSprite struct {
	id      SpriteGroupId
	sprite  Sprite
	updates bool
}

type SpriteGroup struct {
	host     weak.Pointer[engine.Host]
	nextId   int
	index    []IndexedSprite
	updateId engine.UpdateId
}

func (g *SpriteGroup) Init(host *engine.Host) {
	g.host = weak.Make(host)
	// TODO:  Need to remove the update somewhere
	g.updateId = host.Updater.AddUpdate(g.update)
}

func (g *SpriteGroup) Reserve(count int) {
	g.index = klib.SliceSetCap(g.index, count)
}

func (g *SpriteGroup) Find(id SpriteGroupId) *Sprite {
	for i := range g.index {
		if g.index[i].id == id {
			return &g.index[i].sprite
		}
	}
	return nil
}

func (g *SpriteGroup) Add(sprite Sprite) SpriteGroupId {
	g.nextId++
	if g.updateId > 0 {
		host := g.host.Value()
		debug.EnsureNotNil(host)
		host.Updater.RemoveUpdate(&sprite.updateId)
	}
	sprite.updateId = 0
	entry := IndexedSprite{
		id:      g.nextId,
		sprite:  sprite,
		updates: sprite.isSpriteSheet() || sprite.isFlipBook() || sprite.isUVAnimated(),
	}
	g.index = append(g.index, entry)
	return entry.id
}

func (g *SpriteGroup) AddBlank() *IndexedSprite {
	g.Add(Sprite{})
	return &g.index[len(g.index)-1]
}

func (g *SpriteGroup) Remove(id SpriteGroupId) {
	for i := range g.index {
		if g.index[i].id == id {
			g.index = slices.Delete(g.index, i, i+1)
			break
		}
	}
}

func (g *SpriteGroup) update(deltaTime float64) {
	for i := range g.index {
		if !g.index[i].updates {
			continue
		}
		g.index[i].sprite.update(deltaTime)
	}
}
