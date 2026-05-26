/******************************************************************************/
/* history_stage_workspace_details_change.go                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"reflect"

	"kaijuengine.com/platform/profiler/tracing"
)

type detailsDataChangeHistory struct {
	ValueChangeProcedure func(newVal reflect.Value)
	From                 any
	To                   any
}

func (h *detailsDataChangeHistory) Redo() {
	defer tracing.NewRegion("DetailsChangeHistory.Redo").End()
	h.ValueChangeProcedure(reflect.ValueOf(h.To))
}

func (h *detailsDataChangeHistory) Undo() {
	defer tracing.NewRegion("DetailsChangeHistory.Undo").End()
	h.ValueChangeProcedure(reflect.ValueOf(h.From))
}

func (h *detailsDataChangeHistory) Delete() {}
func (h *detailsDataChangeHistory) Exit()   {}
