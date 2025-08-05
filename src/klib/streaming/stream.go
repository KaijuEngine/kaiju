package streaming

import (
	"encoding/binary"
	"io"
)

func StreamWrite(w io.Writer, data ...any) error {
	var err error
	for i := 0; i < len(data) && err == nil; i++ {
		switch d := data[i].(type) {
		case string:
			err = binary.Write(w, binary.LittleEndian, int32(len(d)))
			if err == nil {
				err = binary.Write(w, binary.LittleEndian, []byte(d))
			}
		case int:
			err = binary.Write(w, binary.LittleEndian, int32(d))
		default:
			err = binary.Write(w, binary.LittleEndian, data[i])
		}
	}
	return err
}

func StreamRead(r io.Reader, outData ...any) error {
	var err error
	for i := 0; i < len(outData) && err == nil; i++ {
		switch d := outData[i].(type) {
		case *string:
			size := int32(0)
			err = binary.Read(r, binary.LittleEndian, &size)
			if err == nil {
				data := make([]byte, size)
				err = binary.Read(r, binary.LittleEndian, data)
				if err == nil {
					*d = string(data)
				}
			}
		case *int:
			out := int32(0)
			err = binary.Read(r, binary.LittleEndian, &out)
			if err == nil {
				*d = int(out)
			}
		default:
			err = binary.Read(r, binary.LittleEndian, outData[i])
		}
	}
	return err
}
