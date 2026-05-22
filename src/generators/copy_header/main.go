package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"

	"kaijuengine.com/platform/filesystem"
)

const containsCheck = "Copyright (c) 2015-present Brent Farris."

const header = `/******************************************************************************/
/* [NAME] */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/`

func findRootFolder() (string, error) {
	wd, err := os.Getwd()
	if _, goMain, _, ok := runtime.Caller(0); ok {
		if newWd, pathErr := filepath.Abs(filepath.Dir(goMain)); pathErr == nil {
			wd = filepath.Dir(newWd + "/../../")
		}
	} else if err != nil {
		return "", err
	}
	return wd, nil
}

func main() {
	root, err := findRootFolder()
	if err != nil {
		panic(err)
	}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" || filepath.Ext(path) == ".s" {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			src, err := filesystem.ReadTextFile(path)
			if err != nil {
				return err
			}
			if !strings.Contains(src, containsCheck) {
				nameInsert := filepath.Base(path)
				nameInsert = nameInsert + strings.Repeat(" ", 80-6-utf8.RuneCountInString(nameInsert))
				namedHeader := strings.Replace(header, "[NAME]", nameInsert, 1)
				newSrc := namedHeader + "\n\n" + src
				if err = filesystem.WriteTextFile(path, newSrc); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
