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
