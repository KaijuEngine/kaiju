//go:build !shipping

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
