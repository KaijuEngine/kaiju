/******************************************************************************/
/* content_archive_packer.go                                                  */
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

package content_archive

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"kaiju/debug"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unsafe"
)

func title() []byte { return []byte{0x50, 0x45, 0x43, 0x4B} } // "PECK"

type SourceContent struct {
	Key              string
	FullPath         string
	RawData          []byte
	CustomSerializer func(rawData []byte) ([]byte, error)
}

func CreateArchiveFromFolder(inPath, outPath string, key []byte) error {
	defer tracing.NewRegion("content_archive.CreateArchiveFromFolder").End()
	files := []SourceContent{}
	err := filepath.Walk(inPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		relPath := strings.TrimPrefix(filepath.ToSlash(path), inPath+"/")
		if relPath == "" {
			return nil
		}
		files = append(files, SourceContent{
			Key:      relPath,
			FullPath: path,
		})
		return nil
	})
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no assets found in the supplied folder: %s", inPath)
	}
	return CreateArchiveFromFiles(outPath, files, key)
}

func CreateArchiveFromFiles(outPath string, files []SourceContent, key []byte) error {
	defer tracing.NewRegion("content_archive.CreateArchiveFromFiles").End()
	if len(files) == 0 {
		return fmt.Errorf("no assets were provided to archive")
	}
	entries := make([]Asset, 0, len(files))
	buff := bytes.NewBuffer([]byte{})
	obfData := make([]byte, 0)
	for i := range files {
		var err error
		buff.Reset()
		if len(files[i].RawData) > 0 {
			_, err = buff.ReadFrom(bytes.NewReader(files[i].RawData))
		} else {
			var f *os.File
			if f, err = os.Open(files[i].FullPath); err == nil {
				_, err = buff.ReadFrom(f)
			}
		}
		if err != nil {
			return err
		}
		srcData := buff.Bytes()
		if files[i].CustomSerializer != nil {
			srcData, err = files[i].CustomSerializer(srcData)
		}
		if err != nil {
			return err
		}
		crc := crc32.ChecksumIEEE(srcData)
		obfData = obfData[:0]
		obfData = slices.Grow(obfData, len(srcData))
		copy(obfData, srcData)
		keyLen := len(key)
		if keyLen > 0 {
			for i, b := range obfData {
				obfData[i] = b ^ key[i%keyLen]
			}
		}
		entries = append(entries, Asset{
			Name: files[i].Key,
			Data: obfData,
			Size: uint32(len(srcData)),
			CRC:  crc,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	indexSize := uint64(0)
	for i := range entries {
		indexSize += uint64(len([]byte(entries[i].Name))+1) +
			uint64(unsafe.Sizeof(entries[i].Offset)+
				unsafe.Sizeof(entries[i].Size)+unsafe.Sizeof(entries[i].CRC))
	}
	indexBuf := make([]byte, 0, indexSize)
	for i := range entries {
		indexBuf = append(indexBuf, []byte(entries[i].Name)...)
		indexBuf = append(indexBuf, byte(0))                     // Name null terminator
		indexBuf = binary.LittleEndian.AppendUint64(indexBuf, 0) // Offset placeholder
		indexBuf = binary.LittleEndian.AppendUint32(indexBuf, entries[i].Size)
		indexBuf = binary.LittleEndian.AppendUint32(indexBuf, entries[i].CRC)
	}
	debug.Ensure(indexSize == uint64(len(indexBuf)))
	dataStart := uint64(18) + indexSize
	offset := dataStart
	for i := range entries {
		entries[i].Offset = offset
		pad := (4 - int(entries[i].Size)%4) % 4
		offset += uint64(entries[i].Size) + uint64(pad)
	}
	indexWritePos := uint64(0)
	for i := range entries {
		indexWritePos += uint64(len([]byte(entries[i].Name)) + 1)
		binary.LittleEndian.PutUint64(indexBuf[indexWritePos:], entries[i].Offset)
		indexWritePos += uint64(unsafe.Sizeof(entries[i].Offset) +
			unsafe.Sizeof(entries[i].Size) + unsafe.Sizeof(entries[i].CRC))
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(title())
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint32(len(entries)))
	binary.Write(f, binary.LittleEndian, indexSize)
	f.Write(indexBuf)
	totalSize := uint64(0)
	for i := range entries {
		f.Write(entries[i].Data)
		pad := (4 - int(entries[i].Size)%4) % 4
		if pad > 0 {
			f.Write(make([]byte, pad))
		}
		totalSize += uint64(len(entries[i].Data) + pad)
	}
	slog.Info("packaged content archive", "count", len(entries),
		"size", totalSize, "obfuscated", len(key) > 0)
	return nil
}
