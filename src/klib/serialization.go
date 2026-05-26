/******************************************************************************/
/* serialization.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"bytes"
	"encoding/json"
	"io"
)

func JsonDecode[T any](decoder *json.Decoder, container *T) error {
	if err := decoder.Decode(container); err == io.EOF {
		return err
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

func ByteArrayToString(byteArray []byte) string {
	return string(string(bytes.TrimRight(byteArray[:], "\x00")))
}
