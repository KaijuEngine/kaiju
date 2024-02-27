//go:build debug && !editor

/******************************************************************************/
/* main.rt.dbg.go                                                             */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package bootstrap

import (
	"flag"
	"kaiju/assets/asset_info"
	"kaiju/engine"
	"kaiju/systems/stages"
	"log/slog"
	"os"
	"strings"
)

type cla struct {
	stage string
}

func logOps() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
}

func setupDebug(host *engine.Host) error {
	cla := buildCLA()
	if cla.stage != "" {
		path := strings.ReplaceAll(cla.stage, "\\", "/")
		if !strings.HasPrefix(path, "content/") {
			path = "content/stages/" + cla.stage + ".stg"
		}
		adi, err := asset_info.Read(path)
		if err != nil {
			return err
		}
		return stages.Load(adi, host)
	}
	return nil
}

func buildCLA() cla {
	fs := flag.NewFlagSet("Kaiju Debug Args", flag.ContinueOnError)
	stage := fs.String("stage", "", "The stage to immediately load into")
	fs.Parse(os.Args[1:])
	return cla{
		stage: *stage,
	}
}
