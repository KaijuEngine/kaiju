package klib

import (
	"os"
	"sort"
)

type dirEntries []os.DirEntry

func (d dirEntries) Len() int {
	return len(d)
}

func (d dirEntries) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d dirEntries) Less(i, j int) bool {
	if d[i].IsDir() != d[j].IsDir() {
		return d[i].IsDir()
	}
	return d[i].Name() < d[j].Name()
}

func SortDirEntries(entries []os.DirEntry) []os.DirEntry {
	sorted := dirEntries(entries)
	sort.Sort(sorted)
	return sorted
}
