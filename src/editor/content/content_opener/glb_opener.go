package content_opener

import (
	"kaiju/editor/content/content_history"
	"kaiju/editor/editor_config"
	"kaiju/editor/editor_interface"
	"kaiju/engine"
	"kaiju/engine/assets/asset_importer"
	"kaiju/engine/assets/asset_info"
	"kaiju/engine/collision"
)

type GlbOpener struct{}

func (o GlbOpener) Handles(adi asset_info.AssetDatabaseInfo) bool {
	return adi.Type == editor_config.AssetTypeGlb
}

func (o GlbOpener) Open(adi asset_info.AssetDatabaseInfo, ed editor_interface.Editor) error {
	host := ed.Host()
	e := engine.NewEntity(ed.Host().WorkGroup())
	e.GenerateId()
	e.Transform.SetPosition(ed.Camera().LookAtPoint())
	host.AddEntity(e)
	meta := adi.Metadata.(*asset_importer.MeshMetadata)
	e.SetName(meta.Name)
	bvh := collision.NewBVH()
	bvh.Transform = &e.Transform
	for i := range adi.Children {
		if err := loadMesh(host, adi.Children[i], e, bvh); err != nil {
			return err
		}
	}
	if !bvh.IsLeaf() {
		e.EditorBindings.Set("bvh", bvh)
		ed.BVH().Insert(bvh)
		e.OnDestroy.Add(func() { bvh.RemoveNode() })
	}
	ed.History().Add(&content_history.ModelOpen{
		Host:   host,
		Entity: e,
		Editor: ed,
	})
	ed.ReloadTabs("Hierarchy")
	host.Window.Focus()
	return nil
}
