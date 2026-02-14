package vector_graphics

import (
	"kaiju/matrix"
	"time"
)

// AnimationProperty represents a typed property that can be animated
type AnimationProperty uint8

const (
	AnimPropertyNone AnimationProperty = iota

	// Transform properties
	AnimPropertyPositionX
	AnimPropertyPositionY
	AnimPropertyScaleX
	AnimPropertyScaleY
	AnimPropertyRotation
	AnimPropertySkewX
	AnimPropertySkewY

	// Shape properties
	AnimPropertyStrokeWidth
	AnimPropertyStrokeColorR
	AnimPropertyStrokeColorG
	AnimPropertyStrokeColorB
	AnimPropertyStrokeColorA
	AnimPropertyFillColorR
	AnimPropertyFillColorG
	AnimPropertyFillColorB
	AnimPropertyFillColorA

	// Visibility
	AnimPropertyOpacity
)

// AnimationKeyframe represents a point in time with interpolated values
type AnimationKeyframe struct {
	Time   time.Duration
	Value  matrix.Float
	Easing EasingFunc
}

// EasingFunc defines an easing function for animations
type EasingFunc func(t matrix.Float) matrix.Float

// Animation represents an animated property of a vector graphic element
type Animation struct {
	Property  AnimationProperty
	Keyframes []AnimationKeyframe
	Duration  time.Duration
	Repeat    int // Number of repeats (0 = infinite)
}

func (a *Animation) IsValid() bool { return a.Property != AnimPropertyNone }
