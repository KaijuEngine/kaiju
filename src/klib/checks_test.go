/******************************************************************************/
/* checks_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"errors"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------

func TestMust_NoError(t *testing.T) {
	// Must with nil error should not panic.
	Must(nil)
}

func TestMust_WithError(t *testing.T) {
	testErr := errors.New("test error")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Must should have panicked on non-nil error")
		}
	}()

	Must(testErr)
}

// ---------------------------------------------------------------------------

func TestMustReturn_NoError(t *testing.T) {
	result := MustReturn(42, nil)
	if result != 42 {
		t.Errorf("MustReturn should return the value when err is nil, got %d", result)
	}
}

func TestMustReturn_WithString(t *testing.T) {
	result := MustReturn("hello", nil)
	if result != "hello" {
		t.Errorf("MustReturn should return string value, got %s", result)
	}
}

func TestMustReturn_WithError(t *testing.T) {
	testErr := errors.New("test error")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustReturn should have panicked on non-nil error")
		}
	}()

	MustReturn(0, testErr)
}

// ---------------------------------------------------------------------------

func TestMustReturn2_NoError(t *testing.T) {
	a, b := MustReturn2(10, 20, nil)
	if a != 10 || b != 20 {
		t.Errorf("MustReturn2 should return both values, got (%d, %d)", a, b)
	}
}

func TestMustReturn2_DifferentTypes(t *testing.T) {
	a, b := MustReturn2("hello", 3.14, nil)
	if a != "hello" || b != 3.14 {
		t.Errorf("MustReturn2 should return different types, got (%s, %f)", a, b)
	}
}

func TestMustReturn2_WithError(t *testing.T) {
	testErr := errors.New("test error")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustReturn2 should have panicked on non-nil error")
		}
	}()

	MustReturn2(0, 0, testErr)
}

// ---------------------------------------------------------------------------

func TestShould_NoError(t *testing.T) {
	result := Should(nil)
	if result != false {
		t.Errorf("Should should return false for nil error, got %t", result)
	}
}

func TestShould_WithError(t *testing.T) {
	testErr := errors.New("test error")
	result := Should(testErr)
	if result != true {
		t.Errorf("Should should return true for non-nil error, got %t", result)
	}
}

// ---------------------------------------------------------------------------

func TestShouldReturn_NoError(t *testing.T) {
	result := ShouldReturn(99, nil)
	if result != 99 {
		t.Errorf("ShouldReturn should return the value when err is nil, got %d", result)
	}
}

func TestShouldReturn_WithError(t *testing.T) {
	testErr := errors.New("test error")
	result := ShouldReturn(42, testErr)
	if result != 42 {
		t.Errorf("ShouldReturn should still return the value even with error, got %d", result)
	}
}

// ---------------------------------------------------------------------------

func TestFloatEquals_Float64_Equal(t *testing.T) {
	a, b := 3.141592653589793, 3.141592653589793
	if !FloatEquals(a, b) {
		t.Error("FloatEquals should return true for equal float64 values")
	}
}

func TestFloatEquals_Float32_Equal(t *testing.T) {
	a, b := float32(3.14159), float32(3.14159)
	if !FloatEquals(a, b) {
		t.Error("FloatEquals should return true for equal float32 values")
	}
}

func TestFloatEquals_NearEqual(t *testing.T) {
	a := 1.0
	b := 1.0 + math.SmallestNonzeroFloat64/2 // Within tolerance
	if !FloatEquals(a, b) {
		t.Error("FloatEquals should return true for nearly equal values within tolerance")
	}
}

func TestFloatEquals_NotEqual(t *testing.T) {
	a, b := 1.0, 2.0
	if FloatEquals(a, b) {
		t.Error("FloatEquals should return false for significantly different values")
	}
}

func TestFloatEquals_Zero(t *testing.T) {
	a, b := 0.0, 0.0
	if !FloatEquals(a, b) {
		t.Error("FloatEquals should return true for zero values")
	}
}

func TestFloatEquals_Negative(t *testing.T) {
	a, b := -1.234, -1.234
	if !FloatEquals(a, b) {
		t.Error("FloatEquals should return true for equal negative values")
	}
}

func TestFloatEquals_LargeValues(t *testing.T) {
	a, b := 1e20, 1e20
	if !FloatEquals(a, b) {
		t.Error("FloatEquals should return true for equal large values")
	}
}

func TestFloatEquals_SmallTolerance(t *testing.T) {
	// Use values with a meaningful gap — any difference > SmallestNonzeroFloat64
	// is outside tolerance.
	a := 1.0
	b := 2.0
	if FloatEquals(a, b) {
		t.Error("FloatEquals should return false for values outside tolerance")
	}
}
