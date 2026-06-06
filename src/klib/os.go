/******************************************************************************/
/* os.go                                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"io"
	"os"
	"runtime"
)

func ExeExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func ReadRootFile(fs *os.Root, filePath string) ([]byte, error) {
	if f, err := fs.OpenFile(filePath, os.O_RDONLY, os.ModePerm); err != nil {
		return []byte{}, err
	} else {
		b, e := io.ReadAll(f)
		f.Close()
		return b, e
	}
}

func WriteRootFile(fs *os.Root, filePath string, data []byte) error {
	if f, err := fs.OpenFile(filePath, os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		_, err = f.Write(data)
		f.Close()
		return err
	}
}

func IsMobile() bool {
	switch runtime.GOOS {
	case "android":
		fallthrough
	case "ios":
		return true
	}
	return false
}
