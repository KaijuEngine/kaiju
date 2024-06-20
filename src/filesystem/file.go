/******************************************************************************/
/* file.go                                                                    */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package filesystem

import (
	"io"
	"os"
	"strings"
)

// WriteFile writes the binary data to the file at the specified path. This will
// use permissions 0644 for the file. If the file already exists, it will be
// overwritten.
func WriteFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// WriteTextFile writes the text data to the file at the specified path. This
// will use permissions 0644 for the file. If the file already exists, it will
// be overwritten.
func WriteTextFile(path string, data string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// FileExists returns true if the file exists at the specified path.
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// ReadFile reads the binary data from the file at the specified path. If the
// file does not exist, an error will be returned.
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()
	buff, err := io.ReadAll(file)
	return buff, err
}

// ReadTextFile reads the text data from the file at the specified path. If the
// file does not exist, an error will be returned.
func ReadTextFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var text strings.Builder
	_, err = io.Copy(&text, file)
	return text.String(), err
}

// CopyFile copies the file from the source path to the destination path. If the
// destination file already exists, an error will be returned.
func CopyFile(src, dst string) error {
	if strings.HasSuffix(src, ".go") {
		return CopyGoSourceFile(src, dst)
	} else {
		sf, err := os.Open(src)
		if err != nil {
			return err
		}
		defer sf.Close()
		_, err = os.Stat(dst)
		if err == nil {
			return os.ErrExist
		}
		df, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer df.Close()
		_, err = io.Copy(df, sf)
		return err
	}
}

// CopyGoSourceFile copies go specific source code from the source path to the
// destination path. If the destination file already exists, an error will be
// returned. This function is used to ensure that the generated files have the
// correct header.
func CopyGoSourceFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	_, err = os.Stat(dst)
	if err == nil {
		return os.ErrExist
	}
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = df.WriteString(`// Code generated by "Kaiju build system"; DO NOT EDIT.
`)
	if err != nil {
		return err
	}
	_, err = io.Copy(df, sf)
	return err
}
