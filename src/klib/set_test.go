/******************************************************************************/
/* set_test.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestNewSet(t *testing.T) {
	s := NewSet[int]()
	if len(s) != 0 {
		t.Errorf("NewSet expected empty set, got %d elements", len(s))
	}
}

func TestAdd(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Add(1) // duplicate

	if len(s) != 2 {
		t.Errorf("Add expected 2 elements, got %d", len(s))
	}

	for _, want := range []int{1, 2} {
		if !s.Contains(want) {
			t.Errorf("Add: expected %d in set", want)
		}
	}
}

func TestAddString(t *testing.T) {
	s := NewSet[string]()
	s.Add("hello")
	s.Add("world")

	if len(s) != 2 {
		t.Errorf("Add (string) expected 2 elements, got %d", len(s))
	}

	if !s.Contains("hello") || !s.Contains("world") {
		t.Error("Add (string): expected both strings in set")
	}
}

func TestRemove(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Remove(1)

	if len(s) != 1 {
		t.Errorf("Remove expected 1 element, got %d", len(s))
	}

	if s.Contains(1) {
		t.Error("Remove: element 1 should not be in set")
	}

	if !s.Contains(2) {
		t.Error("Remove: element 2 should still be in set")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Remove(99) // should not panic

	if len(s) != 1 {
		t.Errorf("Remove non-existent expected 1 element, got %d", len(s))
	}
}

func TestContains(t *testing.T) {
	s := NewSet[string]()
	s.Add("present")

	if !s.Contains("present") {
		t.Error("Contains: expected 'present' in set")
	}

	if s.Contains("absent") {
		t.Error("Contains: expected 'absent' not in set")
	}
}

func TestContainsEmptySet(t *testing.T) {
	s := NewSet[int]()
	if s.Contains(1) {
		t.Error("Contains on empty set should return false")
	}
}

func TestToSlice(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)

	result := s.ToSlice()
	if len(result) != 3 {
		t.Fatalf("ToSlice expected 3 elements, got %d", len(result))
	}

	for _, want := range []int{1, 2, 3} {
		found := slices.Contains(result, want)
		if !found {
			t.Errorf("ToSlice: expected %d in result slice", want)
		}
	}
}

func TestToSliceEmpty(t *testing.T) {
	s := NewSet[int]()
	result := s.ToSlice()
	if result == nil {
		t.Error("ToSlice on empty set should return empty slice, not nil")
	}
	if len(result) != 0 {
		t.Errorf("ToSlice empty set expected 0 elements, got %d", len(result))
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		set      Set[int]
		contains []int
	}{
		{
			name:     "empty set",
			set:      NewSet[int](),
			contains: []int{},
		},
		{
			name: "single element",
			set: func() Set[int] {
				s := NewSet[int]()
				s.Add(42)
				return s
			}(),
			contains: []int{42},
		},
		{
			name: "multiple elements",
			set: func() Set[int] {
				s := NewSet[int]()
				s.Add(1)
				s.Add(2)
				s.Add(3)
				return s
			}(),
			contains: []int{1, 2, 3},
		},
		{
			name:     "nil set",
			set:      nil,
			contains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.set.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON error: %v", err)
			}

			if tt.contains == nil {
				if string(data) != "null" {
					t.Errorf("MarshalJSON nil set expected null, got %s", string(data))
				}
				return
			}

			var result []int
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal of MarshalJSON output failed: %v", err)
			}

			if len(result) != len(tt.contains) {
				t.Fatalf("MarshalJSON expected %d elements, got %d", len(tt.contains), len(result))
			}

			for _, want := range tt.contains {
				if !slices.Contains(result, want) {
					t.Errorf("MarshalJSON: expected %d in output", want)
				}
			}
		})
	}
}

func TestMarshalJSONStrings(t *testing.T) {
	s := NewSet[string]()
	s.Add("a")
	s.Add("b")

	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(result))
	}

	if !slices.Contains(result, "a") || !slices.Contains(result, "b") {
		t.Error("MarshalJSON: expected 'a' and 'b' in output")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []int
		wantNil  bool
	}{
		{
			name:     "empty array",
			json:     "[]",
			expected: nil,
		},
		{
			name:     "single element",
			json:     "[42]",
			expected: []int{42},
		},
		{
			name:     "multiple elements",
			json:     "[1, 2, 3]",
			expected: []int{1, 2, 3},
		},
		{
			name:    "null",
			json:    "null",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Set[int]
			err := s.UnmarshalJSON([]byte(tt.json))
			if err != nil {
				t.Fatalf("UnmarshalJSON error: %v", err)
			}

			if tt.wantNil {
				if s != nil {
					t.Error("UnmarshalJSON null expected nil set")
				}
				return
			}

			if s == nil {
				t.Fatal("UnmarshalJSON returned nil set")
			}

			if len(s) != len(tt.expected) {
				t.Fatalf("UnmarshalJSON expected %d elements, got %d", len(tt.expected), len(s))
			}

			for _, want := range tt.expected {
				if !s.Contains(want) {
					t.Errorf("UnmarshalJSON: expected %d in set", want)
				}
			}
		})
	}
}

func TestUnmarshalJSONStrings(t *testing.T) {
	var s Set[string]
	err := s.UnmarshalJSON([]byte(`["hello", "world"]`))
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}

	if len(s) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(s))
	}

	if !s.Contains("hello") || !s.Contains("world") {
		t.Error("Expected 'hello' and 'world' in set")
	}
}

func TestUnmarshalJSONInvalid(t *testing.T) {
	var s Set[int]
	err := s.UnmarshalJSON([]byte("not valid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	original := NewSet[int]()
	original.Add(10)
	original.Add(20)
	original.Add(30)

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Set[int]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(decoded) != len(original) {
		t.Fatalf("Round trip: expected %d elements, got %d", len(original), len(decoded))
	}

	for v := range original {
		if !decoded.Contains(v) {
			t.Errorf("Round trip: expected %d in decoded set", v)
		}
	}
}

func TestSetWithStructs(t *testing.T) {
	type Point struct {
		X, Y int
	}

	s := NewSet[Point]()
	s.Add(Point{1, 2})
	s.Add(Point{3, 4})

	if len(s) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(s))
	}

	if !s.Contains(Point{1, 2}) {
		t.Error("Expected Point{1, 2} in set")
	}

	if s.Contains(Point{5, 6}) {
		t.Error("Did not expect Point{5, 6} in set")
	}
}

func TestSetLen(t *testing.T) {
	s := NewSet[int]()
	if len(s) != 0 {
		t.Errorf("Empty set length should be 0, got %d", len(s))
	}

	s.Add(1)
	if len(s) != 1 {
		t.Errorf("After adding 1, length should be 1, got %d", len(s))
	}

	s.Add(1) // duplicate
	if len(s) != 1 {
		t.Errorf("After adding duplicate, length should still be 1, got %d", len(s))
	}

	s.Remove(1)
	if len(s) != 0 {
		t.Errorf("After removing, length should be 0, got %d", len(s))
	}
}
