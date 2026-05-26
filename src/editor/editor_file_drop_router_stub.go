//go:build !editor && !filedrop

/******************************************************************************/
/* editor_file_drop_router_stub.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

/******************************************************************************/
/* editor_file_drop_router_stub.go                                            */
/******************************************************************************/

package editor

type FileDropRouter struct{}

func (ed *Editor) FileDropRouter() *FileDropRouter { return &ed.fileDropRouter }

func (ed *Editor) connectFileDropRouter() {}
