/******************************************************************************/
/* dir_entry_sorter.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"os"
	"sort"
	"strings"
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
	return strings.ToLower(d[i].Name()) < strings.ToLower(d[j].Name())
}

func SortDirEntries(entries []os.DirEntry) []os.DirEntry {
	sorted := dirEntries(entries)
	sort.Sort(sorted)
	return sorted
}
