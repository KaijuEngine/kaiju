/******************************************************************************/
/* cache_database_errors.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_database

import (
	"errors"
	"fmt"
)

var CacheContentNameEqual = errors.New("name already matches new name, nothing to do")

type ReadDuringBuildError struct{}
type NotInCacheError struct {
	Id string
}
type DuplicateIdError struct {
	Id string
}

func (e ReadDuringBuildError) Error() string {
	return "the database is currently building, it can't be used until it's done"
}

func (e NotInCacheError) Error() string {
	return fmt.Sprintf("the id '%s' was not found in the cache", e.Id)
}

func (e DuplicateIdError) Error() string {
	return fmt.Sprintf("the id '%s' already exists in the cache", e.Id)
}
