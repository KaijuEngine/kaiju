package content_archive

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"unsafe"
)

// Asset holds metadata.
type Asset struct {
	Name   string
	Offset uint64
	Size   uint32
	CRC    uint32 // CRC32 of original (deobf) data.
	Data   []byte // Loaded/cached deobf data.
}

// Archive manages the packed assets.
type Archive struct {
	mmap   []byte
	assets map[string]Asset
	obfKey []byte
}

func OpenArchiveFile(path string, key []byte) (*Archive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return OpenArchiveFromBytes(data, key)
}

func OpenArchiveFromBytes(data []byte, key []byte) (*Archive, error) {
	if len(data) < 18 || !bytes.Equal(data[:4], title()) {
		return nil, errors.New("invalid content archive")
	}
	arc := &Archive{
		mmap:   data,
		assets: make(map[string]Asset),
		obfKey: key,
	}
	numFiles := binary.LittleEndian.Uint32(arc.mmap[6:10])
	indexEnd := binary.LittleEndian.Uint64(arc.mmap[10:18])
	readMapArea := arc.mmap[18:]
	pos := uint64(0)
	for i := uint32(0); i < numFiles; i++ {
		nameEnd := bytes.IndexByte(readMapArea[pos:], 0)
		if nameEnd == -1 {
			return nil, fmt.Errorf("bad name at %d", pos)
		}
		a := Asset{}
		a.Name = string(readMapArea[pos : pos+uint64(nameEnd)])
		pos += uint64(nameEnd + 1)
		a.Offset = binary.LittleEndian.Uint64(readMapArea[pos : pos+8])
		pos += uint64(unsafe.Sizeof(a.Offset))
		a.Size = binary.LittleEndian.Uint32(readMapArea[pos : pos+4])
		pos += uint64(unsafe.Sizeof(a.Size))
		a.CRC = binary.LittleEndian.Uint32(readMapArea[pos : pos+4])
		pos += uint64(unsafe.Sizeof(a.CRC))
		arc.assets[a.Name] = a
		if pos > indexEnd {
			return nil, fmt.Errorf("index overflow at file %d", i)
		}
	}
	return arc, nil
}

func (a *Archive) Exists(name string) bool {
	_, ok := a.assets[name]
	return ok
}

func (a *Archive) Read(name string) ([]byte, error) {
	asset, ok := a.assets[name]
	if !ok {
		return nil, fmt.Errorf("asset %q not found", name)
	}
	if asset.Data != nil {
		return asset.Data, nil
	}
	start := int(asset.Offset)
	end := start + int(asset.Size)
	obfData := a.mmap[start:end]
	deobfData := make([]byte, asset.Size)
	copy(deobfData, obfData)
	keyLen := len(a.obfKey)
	if keyLen > 0 {
		for i, b := range deobfData {
			deobfData[i] = b ^ a.obfKey[i%keyLen]
		}
	}
	computedCRC := crc32.ChecksumIEEE(deobfData)
	if computedCRC != asset.CRC {
		return nil, fmt.Errorf("CRC mismatch for %s (expected %08x, got %08x)", name, asset.CRC, computedCRC)
	}
	asset.Data = deobfData
	return deobfData, nil
}
