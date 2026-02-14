/******************************************************************************/
/* set.go                                                                     */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
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
