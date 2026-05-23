/******************************************************************************/
/* map_test.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"reflect"
	"testing"
)

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := MapKeys(m)

	if len(keys) != 3 {
		t.Errorf("len(keys) = %d, expected 3", len(keys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}
	for k := range m {
		if !keySet[k] {
			t.Errorf("key %q not found in returned keys", k)
		}
	}
}

func TestMapKeysEmpty(t *testing.T) {
	m := map[string]int{}
	keys := MapKeys(m)
	if len(keys) != 0 {
		t.Errorf("len(keys) = %d, expected 0", len(keys))
	}
}

func TestMapKeysSorted(t *testing.T) {
	m := map[int]string{3: "c", 1: "a", 2: "b"}
	keys := MapKeysSorted(m)

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("keys = %v, expected %v", keys, expected)
	}
}

func TestMapKeysSortedStrings(t *testing.T) {
	m := map[string]int{"banana": 2, "apple": 1, "cherry": 3}
	keys := MapKeysSorted(m)

	expected := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("keys = %v, expected %v", keys, expected)
	}
}

func TestMapKeysSortedEmpty(t *testing.T) {
	m := map[int]string{}
	keys := MapKeysSorted(m)
	if len(keys) != 0 {
		t.Errorf("len(keys) = %d, expected 0", len(keys))
	}
}

func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	values := MapValues(m)

	if len(values) != 3 {
		t.Errorf("len(values) = %d, expected 3", len(values))
	}

	valSet := make(map[int]bool)
	for _, v := range values {
		valSet[v] = true
	}
	for _, v := range []int{1, 2, 3} {
		if !valSet[v] {
			t.Errorf("value %d not found in returned values", v)
		}
	}
}

func TestMapValuesEmpty(t *testing.T) {
	m := map[string]int{}
	values := MapValues(m)
	if len(values) != 0 {
		t.Errorf("len(values) = %d, expected 0", len(values))
	}
}

func TestMapJoin(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"c": 3, "d": 4}
	result := MapJoin(a, b)

	expected := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, expected %v", result, expected)
	}
}

func TestMapJoinOverlap(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 20, "c": 3}
	result := MapJoin(a, b)

	// b values should overwrite a values
	if result["b"] != 20 {
		t.Errorf("result[\"b\"] = %d, expected 20", result["b"])
	}
	if result["a"] != 1 {
		t.Errorf("result[\"a\"] = %d, expected 1", result["a"])
	}
	if result["c"] != 3 {
		t.Errorf("result[\"c\"] = %d, expected 3", result["c"])
	}
}

func TestMapJoinEmptyMaps(t *testing.T) {
	a := map[string]int{}
	b := map[string]int{}
	result := MapJoin(a, b)
	if len(result) != 0 {
		t.Errorf("len(result) = %d, expected 0", len(result))
	}
}

func TestMapJoinOneEmpty(t *testing.T) {
	a := map[string]int{"a": 1}
	b := map[string]int{}
	result := MapJoin(a, b)

	expected := map[string]int{"a": 1}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, expected %v", result, expected)
	}
}

func TestMapJoinSmallerSecond(t *testing.T) {
	// Tests the swap logic: when len(b) < len(a), they swap
	a := map[string]int{"a": 1, "b": 2, "c": 3}
	b := map[string]int{"d": 4}
	result := MapJoin(a, b)

	expected := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("result = %v, expected %v", result, expected)
	}
}
