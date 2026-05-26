/******************************************************************************/
/* network_test_helpers.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package network

// NewClientMessageFromBytes creates a ClientMessage from raw bytes for testing.
func NewClientMessageFromBytes(data []byte) ClientMessage {
	cm := ClientMessage{messageLen: uint16(len(data))}
	copy(cm.message[:], data)
	return cm
}
