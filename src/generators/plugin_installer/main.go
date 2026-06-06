/******************************************************************************/
/* main.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"os"
	"os/exec"
	"time"

	"kaijuengine.com/platform/filesystem"
)

// This program is used for installing a modified version of the editor.
// Typically this is done when a plugin has been enabled. When a plugin is
// enabled within the editor, the editor will export it's code, compile the code
// with the plugin installed, and then launch this program. After launching this
// program, the editor will close, this program will then copy the newly
// compiled editor executable in it's place, then launch it.

func main() {
	if len(os.Args) < 3 {
		panic("expected 3 args, <exe> <to> <from>")
	}
	to := os.Args[1]
	from := os.Args[2]
	retries := 10
	for retries > 0 {
		if err := filesystem.CopyFileOverwrite(from, to); err == nil {
			if err := exec.Command(to).Run(); err != nil {
				panic(err)
			}
			break
		}
		retries--
		time.Sleep(time.Second)
	}
}
