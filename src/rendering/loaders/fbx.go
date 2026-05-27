/******************************************************************************/
/* fbx.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package loaders

import (
	"errors"
	"path/filepath"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
	fbxdoc "kaijuengine.com/rendering/loaders/fbx"
	"kaijuengine.com/rendering/loaders/load_result"
)

func FBX(path string, assetDB assets.Database) (load_result.Result, error) {
	defer tracing.NewRegion("loaders.FBX").End()
	if !assetDB.Exists(path) {
		return load_result.Result{}, errors.New("file does not exist")
	} else if filepath.Ext(path) == ".fbx" {
		data, err := assetDB.Read(path)
		if err != nil {
			return load_result.Result{}, err
		}
		doc, err := fbxdoc.Parse(data)
		if err != nil {
			return load_result.Result{}, err
		}
		return fbxdoc.ToLoadResultWithPath(doc, path)
	} else {
		return load_result.Result{}, errors.New("invalid file extension")
	}
}
