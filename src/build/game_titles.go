/******************************************************************************/
/* game_titles.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package build

import "strings"

const (
	CompanyName    = "Kaiju Engine"
	CompanyDirName = "KaijuEngine"
)

type GameTitle string

const (
	GameTitleEditor    = GameTitle("Editor")
	GameTitleRawGame   = GameTitle("Kaiju Game")
	GameTitleGenerator = GameTitle("Generator")
)

func (t GameTitle) String() string           { return string(t) }
func (t GameTitle) AsFilePathString() string { return strings.ReplaceAll(string(t), " ", "") }
