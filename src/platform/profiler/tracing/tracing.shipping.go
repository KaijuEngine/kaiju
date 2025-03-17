//go:build shipping

package tracing

type TraceRegion struct{}

func NewTraceRegion(name string) TraceRegion { return TraceRegion{} }
func (t TraceRegion) End()                   {}
