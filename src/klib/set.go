/******************************************************************************/
/* set.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import "encoding/json"

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(map[T]struct{})
}

func (s Set[T]) Add(val T) {
	s[val] = struct{}{}
}

func (s Set[T]) Remove(val T) {
	delete(s, val)
}

func (s Set[T]) Contains(val T) bool {
	_, ok := s[val]
	return ok
}

func (s Set[T]) ToSlice() []T {
	res := make([]T, len(s))
	idx := 0
	for val := range s {
		res[idx] = val
		idx++
	}
	return res
}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

// UnmarshalJSON decodes a JSON array into the set
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}
	var keys []T
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	newSet := make(Set[T], len(keys))
	for _, k := range keys {
		newSet[k] = struct{}{}
	}
	*s = newSet
	return nil
}
