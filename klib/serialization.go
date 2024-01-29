package klib

import (
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
