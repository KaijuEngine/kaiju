/******************************************************************************/
/* text_file_opener.go                                                        */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package content_opener

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func IsATextFile(file string) bool {
	return strings.HasSuffix(file, ".html") ||
		strings.HasSuffix(file, ".css") ||
		strings.HasSuffix(file, ".ini") ||
		strings.HasSuffix(file, ".json") ||
		strings.HasSuffix(file, ".txt") ||
		strings.HasSuffix(file, ".md")
}

func EditTextFile(file string) error {
	var cmd *exec.Cmd
	cmd = exec.Command("code", file)
	if err := cmd.Run(); err != nil {
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", "-a", "TextEdit", file)
		case "linux":
			cmd = exec.Command("gedit", file)
		case "windows":
			cmd = exec.Command("notepad", file)
		default:
			return fmt.Errorf("unsupported OS")
		}
	}
	return cmd.Run()
}
