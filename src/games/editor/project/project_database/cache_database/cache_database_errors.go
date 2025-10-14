package cache_database

import "fmt"

type ReadDuringBuildError struct{}
type NotInCacheError struct {
	Id string
}

func (e ReadDuringBuildError) Error() string {
	return "the database is currently building, it can't be used until it's done"
}

func (e NotInCacheError) Error() string {
	return fmt.Sprintf("the id '%s' was not found in the cache", e.Id)
}
