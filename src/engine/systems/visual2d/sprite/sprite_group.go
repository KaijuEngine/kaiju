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
	updateId int
}

func (g *SpriteGroup) Init(host *engine.Host) {
	g.host = weak.Make(host)
	host.Updater.AddUpdate(g.update)
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
		host.Updater.RemoveUpdate(sprite.updateId)
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
