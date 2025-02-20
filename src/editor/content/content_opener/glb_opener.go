package content_opener

import (
	"kaiju/assets/asset_info"
	"kaiju/collision"
	"kaiju/editor/editor_config"
	"kaiju/editor/interfaces"
	"kaiju/engine"
)

type GlbOpener struct{}

func (o GlbOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeGlb
}

func (o GlbOpener) Open(adi asset_info.AssetDatabaseInfo, ed interfaces.Editor) error {
	host := ed.Host()
	e := engine.NewEntity(ed.Host().WorkGroup())
	e.GenerateId()
	host.AddEntity(e)
	e.SetName(adi.MetaValue("name"))
	bvh := collision.NewBVH()
	bvh.Transform = &e.Transform
	for i := range adi.Children {
		if err := load(host, adi.Children[i], e, bvh); err != nil {
			return err
		}
	}
	if !bvh.IsLeaf() {
		e.EditorBindings.Set("bvh", bvh)
		ed.BVH().Insert(bvh)
		e.OnDestroy.Add(func() { bvh.RemoveNode() })
	}
	ed.History().Add(&modelOpenHistory{
		host:   host,
		entity: e,
	})
	ed.Hierarchy().Reload()
	host.Window.Focus()
	return nil
}
