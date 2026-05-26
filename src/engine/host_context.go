/******************************************************************************/
/* host_context.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"context"
	"time"
)

/******************************************************************************/
/* Functions to fulfil the context.Context interface                          */
/******************************************************************************/

// Deadline is here to fulfil context.Context and will return zero and false
func (h *Host) Deadline() (time.Time, bool) { return time.Time{}, false }

// Done is here to fulfil context.Context and will cose the CloseSignal channel
func (h *Host) Done() <-chan struct{} { return h.CloseSignal }

// Err is here to fulfil context.Context and will return nil or context.Canceled
func (h *Host) Err() error {
	if h.Closing {
		return context.Canceled
	}
	return nil
}

// Value is here to fulfil context.Context and will always return nil
func (h *Host) Value(key any) any { return nil }
