//go:build !debug

/******************************************************************************/
/* tracing.shipping.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package tracing

type TraceRegion struct{}

func NewRegion(name string) TraceRegion { return TraceRegion{} }
func (t TraceRegion) End()              {}
