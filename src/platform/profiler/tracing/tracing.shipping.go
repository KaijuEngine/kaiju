//go:build !debug

package tracing

type TraceRegion struct{}

func NewRegion(name string) TraceRegion { return TraceRegion{} }
func (t TraceRegion) End()              {}
