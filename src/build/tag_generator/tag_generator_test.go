package main

import (
	"os"
	"testing"
)

func TestCreateTagFiles(t *testing.T) {
	createTagFiles()
	for i := range availableTags {
		if _, err := os.Stat(setFile(availableTags[i])); err != nil {
			t.Fail()
		}
		if _, err := os.Stat(notFile(availableTags[i])); err != nil {
			t.Fail()
		}
	}
}
