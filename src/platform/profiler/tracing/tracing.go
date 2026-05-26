//go:build debug

/******************************************************************************/
/* tracing.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package tracing

import (
	"context"
	"runtime/trace"
)

var ctx = context.Background()

type TraceRegion struct {
	r *trace.Region
}

func NewRegion(name string) TraceRegion {
	return TraceRegion{r: trace.StartRegion(ctx, name)}
}

func (t TraceRegion) End() {
	t.r.End()
}
