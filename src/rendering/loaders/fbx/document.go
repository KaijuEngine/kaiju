/******************************************************************************/
/* document.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fbx

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	BinaryHeader = "Kaydara FBX Binary  \x00\x1a\x00"
)

var (
	ErrASCIINotSupported = errors.New("ASCII FBX is not supported yet")
	ErrInvalidHeader     = errors.New("invalid FBX file")
)

type Document struct {
	Version uint32
}

func Parse(data []byte) (Document, error) {
	if isASCII(data) {
		return Document{}, ErrASCIINotSupported
	}
	if !bytes.HasPrefix(data, []byte(BinaryHeader)) {
		return Document{}, ErrInvalidHeader
	}
	versionOffset := len(BinaryHeader)
	if len(data) < versionOffset+4 {
		return Document{}, ErrInvalidHeader
	}
	return Document{
		Version: binary.LittleEndian.Uint32(data[versionOffset : versionOffset+4]),
	}, nil
}

func isASCII(data []byte) bool {
	const sniffBytes = 256
	data = bytes.TrimLeft(data, " \t\r\n")
	if len(data) == 0 {
		return false
	}
	check := data
	if len(check) > sniffBytes {
		check = check[:sniffBytes]
	}
	return bytes.HasPrefix(check, []byte("Kaydara FBX ASCII")) ||
		(bytes.HasPrefix(check, []byte(";")) && bytes.Contains(check, []byte("FBX")))
}
