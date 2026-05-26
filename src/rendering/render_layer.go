/******************************************************************************/
/* render_layer.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

type RenderLayer uint8
type RenderLayerMask uint64

const (
	RenderLayerIndexWorld RenderLayer = iota
	RenderLayerIndexUI
	RenderLayerIndexEditor
	RenderLayerIndexEditorPicking
	RenderLayerIndexEditorGizmoPicking
)

const (
	RenderLayerWorld              RenderLayerMask = 1 << RenderLayerIndexWorld
	RenderLayerUI                 RenderLayerMask = 1 << RenderLayerIndexUI
	RenderLayerEditor             RenderLayerMask = 1 << RenderLayerIndexEditor
	RenderLayerEditorPicking      RenderLayerMask = 1 << RenderLayerIndexEditorPicking
	RenderLayerEditorGizmoPicking RenderLayerMask = 1 << RenderLayerIndexEditorGizmoPicking
	RenderLayerAll                RenderLayerMask = RenderLayerWorld | RenderLayerUI | RenderLayerEditor
)

func (l RenderLayer) Mask() RenderLayerMask {
	return RenderLayerMask(1) << l
}

func normalizeRenderLayerMask(mask RenderLayerMask) RenderLayerMask {
	if mask == 0 {
		return RenderLayerWorld
	}
	return mask
}
