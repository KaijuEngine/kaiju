/******************************************************************************/
/* mock_database.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package assets

import (
	"errors"
)

// MockDatabase implements the assets.Database interface for testing.
type MockDatabase struct {
	files map[string][]byte
}

func (m *MockDatabase) PostWindowCreate(windowHandle PostWindowCreateHandle) error { return nil }

func (m *MockDatabase) Cache(key string, data []byte) {}
func (m *MockDatabase) CacheRemove(key string)        {}
func (m *MockDatabase) CacheClear()                   {}
func (m *MockDatabase) Close()                        {}

func (m *MockDatabase) Exists(key string) bool { _, ok := m.files[key]; return ok }

func (m *MockDatabase) Read(key string) ([]byte, error) {
	if v, ok := m.files[key]; ok {
		return v, nil
	}
	return []byte{}, errors.New("file not found")
}

func (m *MockDatabase) ReadText(key string) (string, error) {
	if v, ok := m.files[key]; ok {
		return string(v), nil
	}
	return "", errors.New("file not found")
}

func (m *MockDatabase) AddFile(key string, data []byte) {
	m.files[key] = data
}

func (m *MockDatabase) RemoveFile(key string) {
	delete(m.files, key)
}

func NewMockDB(files map[string][]byte) *MockDatabase {
	return &MockDatabase{files: files}
}
