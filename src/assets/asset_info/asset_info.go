package asset_info

import (
	"encoding/json"
	"errors"
	"kaiju/filesystem"
	"os"
)

const (
	infoExtension = ".adi"
)

type AssetDatabaseInfo struct {
	ID   string
	Path string
}

func toADI(path string) string {
	return path + infoExtension
}

func HasInfo(path string) bool {
	s, err := os.Stat(toADI(path))
	return err == nil && !s.IsDir()
}

func CreateInfo(path string, id string) (AssetDatabaseInfo, error) {
	if HasInfo(path) {
		return AssetDatabaseInfo{}, errors.New("asset database already has info for this file")
	}
	adi := AssetDatabaseInfo{
		ID:   id,
		Path: path,
	}
	adiFile := toADI(path)
	src, err := json.Marshal(adi)
	if err != nil {
		return adi, err
	}
	return adi, filesystem.WriteTextFile(adiFile, string(src))
}

func ReadInfo(path string) (AssetDatabaseInfo, error) {
	adi := AssetDatabaseInfo{}
	if HasInfo(path) {
		return adi, errors.New("asset database does not have info for this file")
	}
	adiFile := toADI(path)
	src, err := filesystem.ReadTextFile(adiFile)
	if err != nil {
		return adi, err
	}
	if err := json.Unmarshal([]byte(src), &adi); err != nil {
		return adi, err
	}
	return adi, nil
}

func WriteInfo(info AssetDatabaseInfo) error {
	adiFile := toADI(info.Path)
	src, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return filesystem.WriteTextFile(adiFile, string(src))
}

func ID(path string) (string, error) {
	aid, err := ReadInfo(path)
	if err != nil {
		return "", err
	}
	return aid.ID, nil
}
