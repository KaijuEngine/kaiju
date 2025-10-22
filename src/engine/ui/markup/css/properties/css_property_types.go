/******************************************************************************/
/* css_property_types.go                                                      */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package properties

// Specifies an accent color for user-interface controls
type AccentColor struct{}

func (p AccentColor) Key() string { return "accent-color" }
func (p AccentColor) Sort() int   { return 0 }

// Specifies the alignment between the lines inside a flexible container when the items do not use all available space
type AlignContent struct{}

func (p AlignContent) Key() string { return "align-content" }
func (p AlignContent) Sort() int   { return 0 }

// Specifies the alignment for items inside a flexible container
type AlignItems struct{}

func (p AlignItems) Key() string { return "align-items" }
func (p AlignItems) Sort() int   { return 0 }

// Specifies the alignment for selected items inside a flexible container
type AlignSelf struct{}

func (p AlignSelf) Key() string { return "align-self" }
func (p AlignSelf) Sort() int   { return 0 }

// Resets all properties (except unicode-bidi and direction)
type All struct{}

func (p All) Key() string { return "all" }
func (p All) Sort() int   { return 0 }

// A shorthand property for all the animation-* properties
type Animation struct{}

func (p Animation) Key() string { return "animation" }
func (p Animation) Sort() int   { return 0 }

// Specifies a delay for the start of an animation
type AnimationDelay struct{}

func (p AnimationDelay) Key() string { return "animation-delay" }
func (p AnimationDelay) Sort() int   { return 0 }

// Specifies whether an animation should be played forwards, backwards or in alternate cycles
type AnimationDirection struct{}

func (p AnimationDirection) Key() string { return "animation-direction" }
func (p AnimationDirection) Sort() int   { return 0 }

// Specifies how long an animation should take to complete one cycle
type AnimationDuration struct{}

func (p AnimationDuration) Key() string { return "animation-duration" }
func (p AnimationDuration) Sort() int   { return 0 }

// Specifies a style for the element when the animation is not playing (before it starts, after it ends, or both)
type AnimationFillMode struct{}

func (p AnimationFillMode) Key() string { return "animation-fill-mode" }
func (p AnimationFillMode) Sort() int   { return 0 }

// Specifies the number of times an animation should be played
type AnimationIterationCount struct{}

func (p AnimationIterationCount) Key() string { return "animation-iteration-count" }
func (p AnimationIterationCount) Sort() int   { return 0 }

// Specifies a name for the @keyframes animation
type AnimationName struct{}

func (p AnimationName) Key() string { return "animation-name" }
func (p AnimationName) Sort() int   { return 0 }

// Specifies whether the animation is running or paused
type AnimationPlayState struct{}

func (p AnimationPlayState) Key() string { return "animation-play-state" }
func (p AnimationPlayState) Sort() int   { return 0 }

// Specifies the speed curve of an animation
type AnimationTimingFunction struct{}

func (p AnimationTimingFunction) Key() string { return "animation-timing-function" }
func (p AnimationTimingFunction) Sort() int   { return 0 }

// Specifies preferred aspect ratio of an element
type AspectRatio struct{}

func (p AspectRatio) Key() string { return "aspect-ratio" }
func (p AspectRatio) Sort() int   { return 0 }

// Defines a graphical effect to the area behind an element
type BackdropFilter struct{}

func (p BackdropFilter) Key() string { return "backdrop-filter" }
func (p BackdropFilter) Sort() int   { return 0 }

// Defines whether or not the back face of an element should be visible when facing the user
type BackfaceVisibility struct{}

func (p BackfaceVisibility) Key() string { return "backface-visibility" }
func (p BackfaceVisibility) Sort() int   { return 0 }

// A shorthand property for all the background-* properties
type Background struct{}

func (p Background) Key() string { return "background" }
func (p Background) Sort() int   { return 0 }

// Sets whether a background image scrolls with the rest of the page, or is fixed
type BackgroundAttachment struct{}

func (p BackgroundAttachment) Key() string { return "background-attachment" }
func (p BackgroundAttachment) Sort() int   { return 0 }

// Specifies the blending mode of each background layer (color/image)
type BackgroundBlendMode struct{}

func (p BackgroundBlendMode) Key() string { return "background-blend-mode" }
func (p BackgroundBlendMode) Sort() int   { return 0 }

// Defines how far the background (color or image) should extend within an element
type BackgroundClip struct{}

func (p BackgroundClip) Key() string { return "background-clip" }
func (p BackgroundClip) Sort() int   { return 0 }

// Specifies the background color of an element
type BackgroundColor struct{}

func (p BackgroundColor) Key() string { return "background-color" }
func (p BackgroundColor) Sort() int   { return 0 }

// Specifies one or more background images for an element
type BackgroundImage struct{}

func (p BackgroundImage) Key() string { return "background-image" }
func (p BackgroundImage) Sort() int   { return 0 }

// Specifies the origin position of a background image
type BackgroundOrigin struct{}

func (p BackgroundOrigin) Key() string { return "background-origin" }
func (p BackgroundOrigin) Sort() int   { return 0 }

// Specifies the position of a background image
type BackgroundPosition struct{}

func (p BackgroundPosition) Key() string { return "background-position" }
func (p BackgroundPosition) Sort() int   { return 0 }

// Specifies the position of a background image on x-axis
type BackgroundPositionX struct{}

func (p BackgroundPositionX) Key() string { return "background-position-x" }
func (p BackgroundPositionX) Sort() int   { return 0 }

// Specifies the position of a background image on y-axis
type BackgroundPositionY struct{}

func (p BackgroundPositionY) Key() string { return "background-position-y" }
func (p BackgroundPositionY) Sort() int   { return 0 }

// Sets if/how a background image will be repeated
type BackgroundRepeat struct{}

func (p BackgroundRepeat) Key() string { return "background-repeat" }
func (p BackgroundRepeat) Sort() int   { return 0 }

// Specifies the size of the background images
type BackgroundSize struct{}

func (p BackgroundSize) Key() string { return "background-size" }
func (p BackgroundSize) Sort() int   { return 0 }

// Specifies the size of an element in block direction
type BlockSize struct{}

func (p BlockSize) Key() string { return "block-size" }
func (p BlockSize) Sort() int   { return 0 }

// A shorthand property for border-width, border-style and border-color
type Border struct{}

func (p Border) Key() string { return "border" }
func (p Border) Sort() int   { return 1 }

// A shorthand property for border-block-width, border-block-style and border-block-color
type BorderBlock struct{}

func (p BorderBlock) Key() string { return "border-block" }
func (p BorderBlock) Sort() int   { return 1 }

// Sets the color of the borders at start and end in the block direction
type BorderBlockColor struct{}

func (p BorderBlockColor) Key() string { return "border-block-color" }
func (p BorderBlockColor) Sort() int   { return 1 }

// Sets the color of the border at the end in the block direction
type BorderBlockEndColor struct{}

func (p BorderBlockEndColor) Key() string { return "border-block-end-color" }
func (p BorderBlockEndColor) Sort() int   { return 1 }

// Sets the style of the border at the end in the block direction
type BorderBlockEndStyle struct{}

func (p BorderBlockEndStyle) Key() string { return "border-block-end-style" }
func (p BorderBlockEndStyle) Sort() int   { return 1 }

// Sets the width of the border at the end in the block direction
type BorderBlockEndWidth struct{}

func (p BorderBlockEndWidth) Key() string { return "border-block-end-width" }
func (p BorderBlockEndWidth) Sort() int   { return 1 }

// Sets the color of the border at the start in the block direction
type BorderBlockStartColor struct{}

func (p BorderBlockStartColor) Key() string { return "border-block-start-color" }
func (p BorderBlockStartColor) Sort() int   { return 1 }

// Sets the style of the border at the start in the block direction
type BorderBlockStartStyle struct{}

func (p BorderBlockStartStyle) Key() string { return "border-block-start-style" }
func (p BorderBlockStartStyle) Sort() int   { return 1 }

// Sets the width of the border at the start in the block direction
type BorderBlockStartWidth struct{}

func (p BorderBlockStartWidth) Key() string { return "border-block-start-width" }
func (p BorderBlockStartWidth) Sort() int   { return 1 }

// Sets the style of the borders at start and end in the block direction
type BorderBlockStyle struct{}

func (p BorderBlockStyle) Key() string { return "border-block-style" }
func (p BorderBlockStyle) Sort() int   { return 1 }

// Sets the width of the borders at start and end in the block direction
type BorderBlockWidth struct{}

func (p BorderBlockWidth) Key() string { return "border-block-width" }
func (p BorderBlockWidth) Sort() int   { return 1 }

// A shorthand property for border-bottom-width, border-bottom-style and border-bottom-color
type BorderBottom struct{}

func (p BorderBottom) Key() string { return "border-bottom" }
func (p BorderBottom) Sort() int   { return 1 }

// Sets the color of the bottom border
type BorderBottomColor struct{}

func (p BorderBottomColor) Key() string { return "border-bottom-color" }
func (p BorderBottomColor) Sort() int   { return 1 }

// Defines the radius of the border of the bottom-left corner
type BorderBottomLeftRadius struct{}

func (p BorderBottomLeftRadius) Key() string { return "border-bottom-left-radius" }
func (p BorderBottomLeftRadius) Sort() int   { return 1 }

// Defines the radius of the border of the bottom-right corner
type BorderBottomRightRadius struct{}

func (p BorderBottomRightRadius) Key() string { return "border-bottom-right-radius" }
func (p BorderBottomRightRadius) Sort() int   { return 1 }

// Sets the style of the bottom border
type BorderBottomStyle struct{}

func (p BorderBottomStyle) Key() string { return "border-bottom-style" }
func (p BorderBottomStyle) Sort() int   { return 1 }

// Sets the width of the bottom border
type BorderBottomWidth struct{}

func (p BorderBottomWidth) Key() string { return "border-bottom-width" }
func (p BorderBottomWidth) Sort() int   { return 1 }

// Sets whether table borders should collapse into a single border or be separated
type BorderCollapse struct{}

func (p BorderCollapse) Key() string { return "border-collapse" }
func (p BorderCollapse) Sort() int   { return 1 }

// Sets the color of the four borders
type BorderColor struct{}

func (p BorderColor) Key() string { return "border-color" }
func (p BorderColor) Sort() int   { return 1 }

// A shorthand property for all the border-image-* properties
type BorderImage struct{}

func (p BorderImage) Key() string { return "border-image" }
func (p BorderImage) Sort() int   { return 1 }

// Specifies the amount by which the border image area extends beyond the border box
type BorderImageOutset struct{}

func (p BorderImageOutset) Key() string { return "border-image-outset" }
func (p BorderImageOutset) Sort() int   { return 1 }

// Specifies whether the border image should be repeated, rounded or stretched
type BorderImageRepeat struct{}

func (p BorderImageRepeat) Key() string { return "border-image-repeat" }
func (p BorderImageRepeat) Sort() int   { return 1 }

// Specifies how to slice the border image
type BorderImageSlice struct{}

func (p BorderImageSlice) Key() string { return "border-image-slice" }
func (p BorderImageSlice) Sort() int   { return 1 }

// Specifies the path to the image to be used as a border
type BorderImageSource struct{}

func (p BorderImageSource) Key() string { return "border-image-source" }
func (p BorderImageSource) Sort() int   { return 1 }

// Specifies the width of the border image
type BorderImageWidth struct{}

func (p BorderImageWidth) Key() string { return "border-image-width" }
func (p BorderImageWidth) Sort() int   { return 1 }

// A shorthand property for border-inline-width, border-inline-style and border-inline-color
type BorderInline struct{}

func (p BorderInline) Key() string { return "border-inline" }
func (p BorderInline) Sort() int   { return 1 }

// Sets the color of the borders at start and end in the inline direction
type BorderInlineColor struct{}

func (p BorderInlineColor) Key() string { return "border-inline-color" }
func (p BorderInlineColor) Sort() int   { return 1 }

// Sets the color of the border at the end in the inline direction
type BorderInlineEndColor struct{}

func (p BorderInlineEndColor) Key() string { return "border-inline-end-color" }
func (p BorderInlineEndColor) Sort() int   { return 1 }

// Sets the style of the border at the end in the inline direction
type BorderInlineEndStyle struct{}

func (p BorderInlineEndStyle) Key() string { return "border-inline-end-style" }
func (p BorderInlineEndStyle) Sort() int   { return 1 }

// Sets the width of the border at the end in the inline direction
type BorderInlineEndWidth struct{}

func (p BorderInlineEndWidth) Key() string { return "border-inline-end-width" }
func (p BorderInlineEndWidth) Sort() int   { return 1 }

// Sets the color of the border at the start in the inline direction
type BorderInlineStartColor struct{}

func (p BorderInlineStartColor) Key() string { return "border-inline-start-color" }
func (p BorderInlineStartColor) Sort() int   { return 1 }

// Sets the style of the border at the start in the inline direction
type BorderInlineStartStyle struct{}

func (p BorderInlineStartStyle) Key() string { return "border-inline-start-style" }
func (p BorderInlineStartStyle) Sort() int   { return 1 }

// Sets the width of the border at the start in the inline direction
type BorderInlineStartWidth struct{}

func (p BorderInlineStartWidth) Key() string { return "border-inline-start-width" }
func (p BorderInlineStartWidth) Sort() int   { return 1 }

// Sets the style of the borders at start and end in the inline direction
type BorderInlineStyle struct{}

func (p BorderInlineStyle) Key() string { return "border-inline-style" }
func (p BorderInlineStyle) Sort() int   { return 1 }

// Sets the width of the borders at start and end in the inline direction
type BorderInlineWidth struct{}

func (p BorderInlineWidth) Key() string { return "border-inline-width" }
func (p BorderInlineWidth) Sort() int   { return 1 }

// A shorthand property for all the border-left-* properties
type BorderLeft struct{}

func (p BorderLeft) Key() string { return "border-left" }
func (p BorderLeft) Sort() int   { return 1 }

// Sets the color of the left border
type BorderLeftColor struct{}

func (p BorderLeftColor) Key() string { return "border-left-color" }
func (p BorderLeftColor) Sort() int   { return 1 }

// Sets the style of the left border
type BorderLeftStyle struct{}

func (p BorderLeftStyle) Key() string { return "border-left-style" }
func (p BorderLeftStyle) Sort() int   { return 1 }

// Sets the width of the left border
type BorderLeftWidth struct{}

func (p BorderLeftWidth) Key() string { return "border-left-width" }
func (p BorderLeftWidth) Sort() int   { return 1 }

// A shorthand property for the four border-*-radius properties
type BorderRadius struct{}

func (p BorderRadius) Key() string { return "border-radius" }
func (p BorderRadius) Sort() int   { return 1 }

// A shorthand property for all the border-right-* properties
type BorderRight struct{}

func (p BorderRight) Key() string { return "border-right" }
func (p BorderRight) Sort() int   { return 1 }

// Sets the color of the right border
type BorderRightColor struct{}

func (p BorderRightColor) Key() string { return "border-right-color" }
func (p BorderRightColor) Sort() int   { return 1 }

// Sets the style of the right border
type BorderRightStyle struct{}

func (p BorderRightStyle) Key() string { return "border-right-style" }
func (p BorderRightStyle) Sort() int   { return 1 }

// Sets the width of the right border
type BorderRightWidth struct{}

func (p BorderRightWidth) Key() string { return "border-right-width" }
func (p BorderRightWidth) Sort() int   { return 1 }

// Sets the distance between the borders of adjacent cells
type BorderSpacing struct{}

func (p BorderSpacing) Key() string { return "border-spacing" }
func (p BorderSpacing) Sort() int   { return 1 }

// Sets the style of the four borders
type BorderStyle struct{}

func (p BorderStyle) Key() string { return "border-style" }
func (p BorderStyle) Sort() int   { return 1 }

// A shorthand property for border-top-width, border-top-style and border-top-color
type BorderTop struct{}

func (p BorderTop) Key() string { return "border-top" }
func (p BorderTop) Sort() int   { return 1 }

// Sets the color of the top border
type BorderTopColor struct{}

func (p BorderTopColor) Key() string { return "border-top-color" }
func (p BorderTopColor) Sort() int   { return 1 }

// Defines the radius of the border of the top-left corner
type BorderTopLeftRadius struct{}

func (p BorderTopLeftRadius) Key() string { return "border-top-left-radius" }
func (p BorderTopLeftRadius) Sort() int   { return 1 }

// Defines the radius of the border of the top-right corner
type BorderTopRightRadius struct{}

func (p BorderTopRightRadius) Key() string { return "border-top-right-radius" }
func (p BorderTopRightRadius) Sort() int   { return 1 }

// Sets the style of the top border
type BorderTopStyle struct{}

func (p BorderTopStyle) Key() string { return "border-top-style" }
func (p BorderTopStyle) Sort() int   { return 1 }

// Sets the width of the top border
type BorderTopWidth struct{}

func (p BorderTopWidth) Key() string { return "border-top-width" }
func (p BorderTopWidth) Sort() int   { return 1 }

// Sets the width of the four borders
type BorderWidth struct{}

func (p BorderWidth) Key() string { return "border-width" }
func (p BorderWidth) Sort() int   { return 1 }

// Sets the elements position, from the bottom of its parent element
type Bottom struct{}

func (p Bottom) Key() string { return "bottom" }
func (p Bottom) Sort() int   { return 0 }

// Sets the behavior of the background and border of an element at page-break, or, for in-line elements, at line-break.
type BoxDecorationBreak struct{}

func (p BoxDecorationBreak) Key() string { return "box-decoration-break" }
func (p BoxDecorationBreak) Sort() int   { return 0 }

// The box-reflect property is used to create a reflection of an element.
type BoxReflect struct{}

func (p BoxReflect) Key() string { return "box-reflect" }
func (p BoxReflect) Sort() int   { return 0 }

// Attaches one or more shadows to an element
type BoxShadow struct{}

func (p BoxShadow) Key() string { return "box-shadow" }
func (p BoxShadow) Sort() int   { return 0 }

// Defines how the width and height of an element are calculated: should they include padding and borders, or not
type BoxSizing struct{}

func (p BoxSizing) Key() string { return "box-sizing" }
func (p BoxSizing) Sort() int   { return 0 }

// Specifies whether or not a page-, column-, or region-break should occur after the specified element
type BreakAfter struct{}

func (p BreakAfter) Key() string { return "break-after" }
func (p BreakAfter) Sort() int   { return 0 }

// Specifies whether or not a page-, column-, or region-break should occur before the specified element
type BreakBefore struct{}

func (p BreakBefore) Key() string { return "break-before" }
func (p BreakBefore) Sort() int   { return 0 }

// Specifies whether or not a page-, column-, or region-break should occur inside the specified element
type BreakInside struct{}

func (p BreakInside) Key() string { return "break-inside" }
func (p BreakInside) Sort() int   { return 0 }

// Specifies the placement of a table caption
type CaptionSide struct{}

func (p CaptionSide) Key() string { return "caption-side" }
func (p CaptionSide) Sort() int   { return 0 }

// Specifies the color of the cursor (caret) in inputs, textareas, or any element that is editable
type CaretColor struct{}

func (p CaretColor) Key() string { return "caret-color" }
func (p CaretColor) Sort() int   { return 0 }

// Specifies the character encoding used in the style sheet
type Charset struct{}

func (p Charset) Key() string { return "charset" }
func (p Charset) Sort() int   { return 0 }

// Specifies what should happen with the element that is next to a floating element
type Clear struct{}

func (p Clear) Key() string { return "clear" }
func (p Clear) Sort() int   { return 0 }

// Clips an absolutely positioned element
type Clip struct{}

func (p Clip) Key() string { return "clip" }
func (p Clip) Sort() int   { return 0 }

// Sets the color of text
type Color struct{}

func (p Color) Key() string { return "color" }
func (p Color) Sort() int   { return 0 }

// Specifies the number of columns an element should be divided into
type ColumnCount struct{}

func (p ColumnCount) Key() string { return "column-count" }
func (p ColumnCount) Sort() int   { return 0 }

// Specifies how to fill columns, balanced or not
type ColumnFill struct{}

func (p ColumnFill) Key() string { return "column-fill" }
func (p ColumnFill) Sort() int   { return 0 }

// Specifies the gap between the columns
type ColumnGap struct{}

func (p ColumnGap) Key() string { return "column-gap" }
func (p ColumnGap) Sort() int   { return 0 }

// A shorthand property for all the column-rule-* properties
type ColumnRule struct{}

func (p ColumnRule) Key() string { return "column-rule" }
func (p ColumnRule) Sort() int   { return 0 }

// Specifies the color of the rule between columns
type ColumnRuleColor struct{}

func (p ColumnRuleColor) Key() string { return "column-rule-color" }
func (p ColumnRuleColor) Sort() int   { return 0 }

// Specifies the style of the rule between columns
type ColumnRuleStyle struct{}

func (p ColumnRuleStyle) Key() string { return "column-rule-style" }
func (p ColumnRuleStyle) Sort() int   { return 0 }

// Specifies the width of the rule between columns
type ColumnRuleWidth struct{}

func (p ColumnRuleWidth) Key() string { return "column-rule-width" }
func (p ColumnRuleWidth) Sort() int   { return 0 }

// Specifies how many columns an element should span across
type ColumnSpan struct{}

func (p ColumnSpan) Key() string { return "column-span" }
func (p ColumnSpan) Sort() int   { return 0 }

// Specifies the column width
type ColumnWidth struct{}

func (p ColumnWidth) Key() string { return "column-width" }
func (p ColumnWidth) Sort() int   { return 0 }

// A shorthand property for column-width and column-count
type Columns struct{}

func (p Columns) Key() string { return "columns" }
func (p Columns) Sort() int   { return 0 }

// Used with the :before and :after pseudo-elements, to insert generated content
type Content struct{}

func (p Content) Key() string { return "content" }
func (p Content) Sort() int   { return 0 }

// Increases or decreases the value of one or more CSS counters
type CounterIncrement struct{}

func (p CounterIncrement) Key() string { return "counter-increment" }
func (p CounterIncrement) Sort() int   { return 0 }

// Creates or resets one or more CSS counters
type CounterReset struct{}

func (p CounterReset) Key() string { return "counter-reset" }
func (p CounterReset) Sort() int   { return 0 }

// Specifies the mouse cursor to be displayed when pointing over an element
type Cursor struct{}

func (p Cursor) Key() string { return "cursor" }
func (p Cursor) Sort() int   { return 0 }

// Specifies the text direction/writing direction
type Direction struct{}

func (p Direction) Key() string { return "direction" }
func (p Direction) Sort() int   { return 0 }

// Specifies how a certain HTML element should be displayed
type Display struct{}

func (p Display) Key() string { return "display" }
func (p Display) Sort() int   { return 0 }

// Specifies whether or not to display borders and background on empty cells in a table
type EmptyCells struct{}

func (p EmptyCells) Key() string { return "empty-cells" }
func (p EmptyCells) Sort() int   { return 0 }

// Defines effects (e.g. blurring or color shifting) on an element before the element is displayed
type Filter struct{}

func (p Filter) Key() string { return "filter" }
func (p Filter) Sort() int   { return 0 }

// A shorthand property for the flex-grow, flex-shrink, and the flex-basis properties
type Flex struct{}

func (p Flex) Key() string { return "flex" }
func (p Flex) Sort() int   { return 0 }

// Specifies the initial length of a flexible item
type FlexBasis struct{}

func (p FlexBasis) Key() string { return "flex-basis" }
func (p FlexBasis) Sort() int   { return 0 }

// Specifies the direction of the flexible items
type FlexDirection struct{}

func (p FlexDirection) Key() string { return "flex-direction" }
func (p FlexDirection) Sort() int   { return 0 }

// A shorthand property for the flex-direction and the flex-wrap properties
type FlexFlow struct{}

func (p FlexFlow) Key() string { return "flex-flow" }
func (p FlexFlow) Sort() int   { return 0 }

// Specifies how much the item will grow relative to the rest
type FlexGrow struct{}

func (p FlexGrow) Key() string { return "flex-grow" }
func (p FlexGrow) Sort() int   { return 0 }

// Specifies how the item will shrink relative to the rest
type FlexShrink struct{}

func (p FlexShrink) Key() string { return "flex-shrink" }
func (p FlexShrink) Sort() int   { return 0 }

// Specifies whether the flexible items should wrap or not
type FlexWrap struct{}

func (p FlexWrap) Key() string { return "flex-wrap" }
func (p FlexWrap) Sort() int   { return 0 }

// Specifies whether an element should float to the left, right, or not at all
type Float struct{}

func (p Float) Key() string { return "float" }
func (p Float) Sort() int   { return 0 }

// A shorthand property for the font-style, font-variant, font-weight, font-size/line-height, and the font-family properties
type Font struct{}

func (p Font) Key() string { return "font" }
func (p Font) Sort() int   { return 0 }

// A rule that allows websites to download and use fonts other than the "web-safe" fonts
type FontFace struct{}

func (p FontFace) Key() string { return "font-face" }
func (p FontFace) Sort() int   { return 0 }

// Specifies the font family for text
type FontFamily struct{}

func (p FontFamily) Key() string { return "font-family" }
func (p FontFamily) Sort() int   { return 0 }

// Allows control over advanced typographic features in OpenType fonts
type FontFeatureSettings struct{}

func (p FontFeatureSettings) Key() string { return "font-feature-settings" }
func (p FontFeatureSettings) Sort() int   { return 0 }

// Allows authors to use a common name in font-variant-alternate for feature activated differently in OpenType
type FontFeatureValues struct{}

func (p FontFeatureValues) Key() string { return "font-feature-values" }
func (p FontFeatureValues) Sort() int   { return 0 }

// Controls the usage of the kerning information (how letters are spaced)
type FontKerning struct{}

func (p FontKerning) Key() string { return "font-kerning" }
func (p FontKerning) Sort() int   { return 0 }

// Controls the usage of language-specific glyphs in a typeface
type FontLanguageOverride struct{}

func (p FontLanguageOverride) Key() string { return "font-language-override" }
func (p FontLanguageOverride) Sort() int   { return 0 }

// Specifies the font size of text
type FontSize struct{}

func (p FontSize) Key() string { return "font-size" }
func (p FontSize) Sort() int   { return 0 }

// Preserves the readability of text when font fallback occurs
type FontSizeAdjust struct{}

func (p FontSizeAdjust) Key() string { return "font-size-adjust" }
func (p FontSizeAdjust) Sort() int   { return 0 }

// Selects a normal, condensed, or expanded face from a font family
type FontStretch struct{}

func (p FontStretch) Key() string { return "font-stretch" }
func (p FontStretch) Sort() int   { return 0 }

// Specifies the font style for text
type FontStyle struct{}

func (p FontStyle) Key() string { return "font-style" }
func (p FontStyle) Sort() int   { return 0 }

// Controls which missing typefaces (bold or italic) may be synthesized by the browser
type FontSynthesis struct{}

func (p FontSynthesis) Key() string { return "font-synthesis" }
func (p FontSynthesis) Sort() int   { return 0 }

// Specifies whether or not a text should be displayed in a small-caps font
type FontVariant struct{}

func (p FontVariant) Key() string { return "font-variant" }
func (p FontVariant) Sort() int   { return 0 }

// Controls the usage of alternate glyphs associated to alternative names defined in @font-feature-values
type FontVariantAlternates struct{}

func (p FontVariantAlternates) Key() string { return "font-variant-alternates" }
func (p FontVariantAlternates) Sort() int   { return 0 }

// Controls the usage of alternate glyphs for capital letters
type FontVariantCaps struct{}

func (p FontVariantCaps) Key() string { return "font-variant-caps" }
func (p FontVariantCaps) Sort() int   { return 0 }

// Controls the usage of alternate glyphs for East Asian scripts (e.g Japanese and Chinese)
type FontVariantEastAsian struct{}

func (p FontVariantEastAsian) Key() string { return "font-variant-east-asian" }
func (p FontVariantEastAsian) Sort() int   { return 0 }

// Controls which ligatures and contextual forms are used in textual content of the elements it applies to
type FontVariantLigatures struct{}

func (p FontVariantLigatures) Key() string { return "font-variant-ligatures" }
func (p FontVariantLigatures) Sort() int   { return 0 }

// Controls the usage of alternate glyphs for numbers, fractions, and ordinal markers
type FontVariantNumeric struct{}

func (p FontVariantNumeric) Key() string { return "font-variant-numeric" }
func (p FontVariantNumeric) Sort() int   { return 0 }

// Controls the usage of alternate glyphs of smaller size positioned as superscript or subscript regarding the baseline of the font
type FontVariantPosition struct{}

func (p FontVariantPosition) Key() string { return "font-variant-position" }
func (p FontVariantPosition) Sort() int   { return 0 }

// Specifies the weight of a font
type FontWeight struct{}

func (p FontWeight) Key() string { return "font-weight" }
func (p FontWeight) Sort() int   { return 0 }

// A shorthand property for the row-gap and the column-gap properties
type Gap struct{}

func (p Gap) Key() string { return "gap" }
func (p Gap) Sort() int   { return 0 }

// A shorthand property for the grid-template-rows, grid-template-columns, grid-template-areas, grid-auto-rows, grid-auto-columns, and the grid-auto-flow properties
type Grid struct{}

func (p Grid) Key() string { return "grid" }
func (p Grid) Sort() int   { return 0 }

// Either specifies a name for the grid item, or this property is a shorthand property for the grid-row-start, grid-column-start, grid-row-end, and grid-column-end properties
type GridArea struct{}

func (p GridArea) Key() string { return "grid-area" }
func (p GridArea) Sort() int   { return 0 }

// Specifies a default column size
type GridAutoColumns struct{}

func (p GridAutoColumns) Key() string { return "grid-auto-columns" }
func (p GridAutoColumns) Sort() int   { return 0 }

// Specifies how auto-placed items are inserted in the grid
type GridAutoFlow struct{}

func (p GridAutoFlow) Key() string { return "grid-auto-flow" }
func (p GridAutoFlow) Sort() int   { return 0 }

// Specifies a default row size
type GridAutoRows struct{}

func (p GridAutoRows) Key() string { return "grid-auto-rows" }
func (p GridAutoRows) Sort() int   { return 0 }

// A shorthand property for the grid-column-start and the grid-column-end properties
type GridColumn struct{}

func (p GridColumn) Key() string { return "grid-column" }
func (p GridColumn) Sort() int   { return 0 }

// Specifies where to end the grid item
type GridColumnEnd struct{}

func (p GridColumnEnd) Key() string { return "grid-column-end" }
func (p GridColumnEnd) Sort() int   { return 0 }

// Specifies the size of the gap between columns
type GridColumnGap struct{}

func (p GridColumnGap) Key() string { return "grid-column-gap" }
func (p GridColumnGap) Sort() int   { return 0 }

// Specifies where to start the grid item
type GridColumnStart struct{}

func (p GridColumnStart) Key() string { return "grid-column-start" }
func (p GridColumnStart) Sort() int   { return 0 }

// A shorthand property for the grid-row-gap and grid-column-gap properties
type GridGap struct{}

func (p GridGap) Key() string { return "grid-gap" }
func (p GridGap) Sort() int   { return 0 }

// A shorthand property for the grid-row-start and the grid-row-end properties
type GridRow struct{}

func (p GridRow) Key() string { return "grid-row" }
func (p GridRow) Sort() int   { return 0 }

// Specifies where to end the grid item
type GridRowEnd struct{}

func (p GridRowEnd) Key() string { return "grid-row-end" }
func (p GridRowEnd) Sort() int   { return 0 }

// Specifies the size of the gap between rows
type GridRowGap struct{}

func (p GridRowGap) Key() string { return "grid-row-gap" }
func (p GridRowGap) Sort() int   { return 0 }

// Specifies where to start the grid item
type GridRowStart struct{}

func (p GridRowStart) Key() string { return "grid-row-start" }
func (p GridRowStart) Sort() int   { return 0 }

// A shorthand property for the grid-template-rows, grid-template-columns and grid-areas properties
type GridTemplate struct{}

func (p GridTemplate) Key() string { return "grid-template" }
func (p GridTemplate) Sort() int   { return 0 }

// Specifies how to display columns and rows, using named grid items
type GridTemplateAreas struct{}

func (p GridTemplateAreas) Key() string { return "grid-template-areas" }
func (p GridTemplateAreas) Sort() int   { return 0 }

// Specifies the size of the columns, and how many columns in a grid layout
type GridTemplateColumns struct{}

func (p GridTemplateColumns) Key() string { return "grid-template-columns" }
func (p GridTemplateColumns) Sort() int   { return 0 }

// Specifies the size of the rows in a grid layout
type GridTemplateRows struct{}

func (p GridTemplateRows) Key() string { return "grid-template-rows" }
func (p GridTemplateRows) Sort() int   { return 0 }

// Specifies whether a punctuation character may be placed outside the line box
type HangingPunctuation struct{}

func (p HangingPunctuation) Key() string { return "hanging-punctuation" }
func (p HangingPunctuation) Sort() int   { return 0 }

// Sets the height of an element
type Height struct{}

func (p Height) Key() string { return "height" }
func (p Height) Sort() int   { return 0 }

// Sets how to split words to improve the layout of paragraphs
type Hyphens struct{}

func (p Hyphens) Key() string { return "hyphens" }
func (p Hyphens) Sort() int   { return 0 }

// Specifies the type of algorithm to use for image scaling
type ImageRendering struct{}

func (p ImageRendering) Key() string { return "image-rendering" }
func (p ImageRendering) Sort() int   { return 0 }

// Allows you to import a style sheet into another style sheet
type Import struct{}

func (p Import) Key() string { return "import" }
func (p Import) Sort() int   { return 0 }

// Specifies the size of an element in the inline direction
type InlineSize struct{}

func (p InlineSize) Key() string { return "inline-size" }
func (p InlineSize) Sort() int   { return 0 }

// Specifies the distance between an element and the parent element
type Inset struct{}

func (p Inset) Key() string { return "inset" }
func (p Inset) Sort() int   { return 0 }

// Specifies the distance between an element and the parent element in the block direction
type InsetBlock struct{}

func (p InsetBlock) Key() string { return "inset-block" }
func (p InsetBlock) Sort() int   { return 0 }

// Specifies the distance between the end of an element and the parent element in the block direction
type InsetBlockEnd struct{}

func (p InsetBlockEnd) Key() string { return "inset-block-end" }
func (p InsetBlockEnd) Sort() int   { return 0 }

// Specifies the distance between the start of an element and the parent element in the block direction
type InsetBlockStart struct{}

func (p InsetBlockStart) Key() string { return "inset-block-start" }
func (p InsetBlockStart) Sort() int   { return 0 }

// Specifies the distance between an element and the parent element in the inline direction
type InsetInline struct{}

func (p InsetInline) Key() string { return "inset-inline" }
func (p InsetInline) Sort() int   { return 0 }

// Specifies the distance between the end of an element and the parent element in the inline direction
type InsetInlineEnd struct{}

func (p InsetInlineEnd) Key() string { return "inset-inline-end" }
func (p InsetInlineEnd) Sort() int   { return 0 }

// Specifies the distance between the start of an element and the parent element in the inline direction
type InsetInlineStart struct{}

func (p InsetInlineStart) Key() string { return "inset-inline-start" }
func (p InsetInlineStart) Sort() int   { return 0 }

// Defines whether an element must create a new stacking content
type Isolation struct{}

func (p Isolation) Key() string { return "isolation" }
func (p Isolation) Sort() int   { return 0 }

// Specifies the alignment between the items inside a flexible container when the items do not use all available space
type JustifyContent struct{}

func (p JustifyContent) Key() string { return "justify-content" }
func (p JustifyContent) Sort() int   { return 0 }

// Is set on the grid container. Specifies the alignment of grid items in the inline direction
type JustifyItems struct{}

func (p JustifyItems) Key() string { return "justify-items" }
func (p JustifyItems) Sort() int   { return 0 }

// Is set on the grid item. Specifies the alignment of the grid item in the inline direction
type JustifySelf struct{}

func (p JustifySelf) Key() string { return "justify-self" }
func (p JustifySelf) Sort() int   { return 0 }

// Specifies the animation code
type Keyframes struct{}

func (p Keyframes) Key() string { return "keyframes" }
func (p Keyframes) Sort() int   { return 0 }

// Specifies the left position of a positioned element
type Left struct{}

func (p Left) Key() string { return "left" }
func (p Left) Sort() int   { return 0 }

// Increases or decreases the space between characters in a text
type LetterSpacing struct{}

func (p LetterSpacing) Key() string { return "letter-spacing" }
func (p LetterSpacing) Sort() int   { return 0 }

// Specifies how/if to break lines
type LineBreak struct{}

func (p LineBreak) Key() string { return "line-break" }
func (p LineBreak) Sort() int   { return 0 }

// Sets the line height
type LineHeight struct{}

func (p LineHeight) Key() string { return "line-height" }
func (p LineHeight) Sort() int   { return 0 }

// Sets all the properties for a list in one declaration
type ListStyle struct{}

func (p ListStyle) Key() string { return "list-style" }
func (p ListStyle) Sort() int   { return 0 }

// Specifies an image as the list-item marker
type ListStyleImage struct{}

func (p ListStyleImage) Key() string { return "list-style-image" }
func (p ListStyleImage) Sort() int   { return 0 }

// Specifies the position of the list-item markers (bullet points)
type ListStylePosition struct{}

func (p ListStylePosition) Key() string { return "list-style-position" }
func (p ListStylePosition) Sort() int   { return 0 }

// Specifies the type of list-item marker
type ListStyleType struct{}

func (p ListStyleType) Key() string { return "list-style-type" }
func (p ListStyleType) Sort() int   { return 0 }

// Sets all the margin properties in one declaration
type Margin struct{}

func (p Margin) Key() string { return "margin" }
func (p Margin) Sort() int   { return 0 }

// Specifies the margin in the block direction
type MarginBlock struct{}

func (p MarginBlock) Key() string { return "margin-block" }
func (p MarginBlock) Sort() int   { return 0 }

// Specifies the margin at the end in the block direction
type MarginBlockEnd struct{}

func (p MarginBlockEnd) Key() string { return "margin-block-end" }
func (p MarginBlockEnd) Sort() int   { return 0 }

// Specifies the margin at the start in the block direction
type MarginBlockStart struct{}

func (p MarginBlockStart) Key() string { return "margin-block-start" }
func (p MarginBlockStart) Sort() int   { return 0 }

// Sets the bottom margin of an element
type MarginBottom struct{}

func (p MarginBottom) Key() string { return "margin-bottom" }
func (p MarginBottom) Sort() int   { return 0 }

// Specifies the margin in the inline direction
type MarginInline struct{}

func (p MarginInline) Key() string { return "margin-inline" }
func (p MarginInline) Sort() int   { return 0 }

// Specifies the margin at the end in the inline direction
type MarginInlineEnd struct{}

func (p MarginInlineEnd) Key() string { return "margin-inline-end" }
func (p MarginInlineEnd) Sort() int   { return 0 }

// Specifies the margin at the start in the inline direction
type MarginInlineStart struct{}

func (p MarginInlineStart) Key() string { return "margin-inline-start" }
func (p MarginInlineStart) Sort() int   { return 0 }

// Sets the left margin of an element
type MarginLeft struct{}

func (p MarginLeft) Key() string { return "margin-left" }
func (p MarginLeft) Sort() int   { return 0 }

// Sets the right margin of an element
type MarginRight struct{}

func (p MarginRight) Key() string { return "margin-right" }
func (p MarginRight) Sort() int   { return 0 }

// Sets the top margin of an element
type MarginTop struct{}

func (p MarginTop) Key() string { return "margin-top" }
func (p MarginTop) Sort() int   { return 0 }

// Hides parts of an element by masking or clipping an image at specific places
type Mask struct{}

func (p Mask) Key() string { return "mask" }
func (p Mask) Sort() int   { return 0 }

// Specifies the mask area
type MaskClip struct{}

func (p MaskClip) Key() string { return "mask-clip" }
func (p MaskClip) Sort() int   { return 0 }

// Represents a compositing operation used on the current mask layer with the mask layers below it
type MaskComposite struct{}

func (p MaskComposite) Key() string { return "mask-composite" }
func (p MaskComposite) Sort() int   { return 0 }

// Specifies an image to be used as a mask layer for an element
type MaskImage struct{}

func (p MaskImage) Key() string { return "mask-image" }
func (p MaskImage) Sort() int   { return 0 }

// Specifies whether the mask layer image is treated as a luminance mask or as an alpha mask
type MaskMode struct{}

func (p MaskMode) Key() string { return "mask-mode" }
func (p MaskMode) Sort() int   { return 0 }

// Specifies the origin position (the mask position area) of a mask layer image
type MaskOrigin struct{}

func (p MaskOrigin) Key() string { return "mask-origin" }
func (p MaskOrigin) Sort() int   { return 0 }

// Sets the starting position of a mask layer image (relative to the mask position area)
type MaskPosition struct{}

func (p MaskPosition) Key() string { return "mask-position" }
func (p MaskPosition) Sort() int   { return 0 }

// Specifies how the mask layer image is repeated
type MaskRepeat struct{}

func (p MaskRepeat) Key() string { return "mask-repeat" }
func (p MaskRepeat) Sort() int   { return 0 }

// Specifies the size of a mask layer image
type MaskSize struct{}

func (p MaskSize) Key() string { return "mask-size" }
func (p MaskSize) Sort() int   { return 0 }

// Specifies whether an SVG <mask> element is treated as a luminance mask or as an alpha mask
type MaskType struct{}

func (p MaskType) Key() string { return "mask-type" }
func (p MaskType) Sort() int   { return 0 }

// Sets the maximum height of an element
type MaxHeight struct{}

func (p MaxHeight) Key() string { return "max-height" }
func (p MaxHeight) Sort() int   { return 0 }

// Sets the maximum width of an element
type MaxWidth struct{}

func (p MaxWidth) Key() string { return "max-width" }
func (p MaxWidth) Sort() int   { return 0 }

// Sets the style rules for different media types/devices/sizes
type Media struct{}

func (p Media) Key() string { return "media" }
func (p Media) Sort() int   { return 0 }

// Sets the maximum size of an element in the block direction
type MaxBlockSize struct{}

func (p MaxBlockSize) Key() string { return "max-block-size" }
func (p MaxBlockSize) Sort() int   { return 0 }

// Sets the maximum size of an element in the inline direction
type MaxInlineSize struct{}

func (p MaxInlineSize) Key() string { return "max-inline-size" }
func (p MaxInlineSize) Sort() int   { return 0 }

// Sets the minimum size of an element in the block direction
type MinBlockSize struct{}

func (p MinBlockSize) Key() string { return "min-block-size" }
func (p MinBlockSize) Sort() int   { return 0 }

// Sets the minimum size of an element in the inline direction
type MinInlineSize struct{}

func (p MinInlineSize) Key() string { return "min-inline-size" }
func (p MinInlineSize) Sort() int   { return 0 }

// Sets the minimum height of an element
type MinHeight struct{}

func (p MinHeight) Key() string { return "min-height" }
func (p MinHeight) Sort() int   { return 0 }

// Sets the minimum width of an element
type MinWidth struct{}

func (p MinWidth) Key() string { return "min-width" }
func (p MinWidth) Sort() int   { return 0 }

// Specifies how an element's content should blend with its direct parent background
type MixBlendMode struct{}

func (p MixBlendMode) Key() string { return "mix-blend-mode" }
func (p MixBlendMode) Sort() int   { return 0 }

// Specifies how the contents of a replaced element should be fitted to the box established by its used height and width
type ObjectFit struct{}

func (p ObjectFit) Key() string { return "object-fit" }
func (p ObjectFit) Sort() int   { return 0 }

// Specifies the alignment of the replaced element inside its box
type ObjectPosition struct{}

func (p ObjectPosition) Key() string { return "object-position" }
func (p ObjectPosition) Sort() int   { return 0 }

// Is a shorthand, and specifies how to animate an element along a path
type Offset struct{}

func (p Offset) Key() string { return "offset" }
func (p Offset) Sort() int   { return 0 }

// Specifies a point on an element that is fixed to the path it is animated along
type OffsetAnchor struct{}

func (p OffsetAnchor) Key() string { return "offset-anchor" }
func (p OffsetAnchor) Sort() int   { return 0 }

// Specifies the position along a path where an animated element is placed
type OffsetDistance struct{}

func (p OffsetDistance) Key() string { return "offset-distance" }
func (p OffsetDistance) Sort() int   { return 0 }

// Specifies the path an element is animated along
type OffsetPath struct{}

func (p OffsetPath) Key() string { return "offset-path" }
func (p OffsetPath) Sort() int   { return 0 }

// Specifies rotation of an element as it is animated along a path
type OffsetRotate struct{}

func (p OffsetRotate) Key() string { return "offset-rotate" }
func (p OffsetRotate) Sort() int   { return 0 }

// Sets the opacity level for an element
type Opacity struct{}

func (p Opacity) Key() string { return "opacity" }
func (p Opacity) Sort() int   { return 0 }

// Sets the order of the flexible item, relative to the rest
type Order struct{}

func (p Order) Key() string { return "order" }
func (p Order) Sort() int   { return 0 }

// Sets the minimum number of lines that must be left at the bottom of a page or column
type Orphans struct{}

func (p Orphans) Key() string { return "orphans" }
func (p Orphans) Sort() int   { return 0 }

// A shorthand property for the outline-width, outline-style, and the outline-color properties
type Outline struct{}

func (p Outline) Key() string { return "outline" }
func (p Outline) Sort() int   { return 0 }

// Sets the color of an outline
type OutlineColor struct{}

func (p OutlineColor) Key() string { return "outline-color" }
func (p OutlineColor) Sort() int   { return 0 }

// Offsets an outline, and draws it beyond the border edge
type OutlineOffset struct{}

func (p OutlineOffset) Key() string { return "outline-offset" }
func (p OutlineOffset) Sort() int   { return 0 }

// Sets the style of an outline
type OutlineStyle struct{}

func (p OutlineStyle) Key() string { return "outline-style" }
func (p OutlineStyle) Sort() int   { return 0 }

// Sets the width of an outline
type OutlineWidth struct{}

func (p OutlineWidth) Key() string { return "outline-width" }
func (p OutlineWidth) Sort() int   { return 0 }

// Specifies what happens if content overflows an element's box
type Overflow struct{}

func (p Overflow) Key() string { return "overflow" }
func (p Overflow) Sort() int   { return 0 }

// Specifies whether or not content in viewable area in a scrollable contianer should be pushed down when new content is loaded above
type OverflowAnchor struct{}

func (p OverflowAnchor) Key() string { return "overflow-anchor" }
func (p OverflowAnchor) Sort() int   { return 0 }

// Specifies whether or not the browser can break lines with long words, if they overflow the container
type OverflowWrap struct{}

func (p OverflowWrap) Key() string { return "overflow-wrap" }
func (p OverflowWrap) Sort() int   { return 0 }

// Specifies whether or not to clip the left/right edges of the content, if it overflows the element's content area
type OverflowX struct{}

func (p OverflowX) Key() string { return "overflow-x" }
func (p OverflowX) Sort() int   { return 0 }

// Specifies whether or not to clip the top/bottom edges of the content, if it overflows the element's content area
type OverflowY struct{}

func (p OverflowY) Key() string { return "overflow-y" }
func (p OverflowY) Sort() int   { return 0 }

// Specifies whether to have scroll chaining or overscroll affordance in x- and y-directions
type OverscrollBehavior struct{}

func (p OverscrollBehavior) Key() string { return "overscroll-behavior" }
func (p OverscrollBehavior) Sort() int   { return 0 }

// Specifies whether to have scroll chaining or overscroll affordance in the block direction
type OverscrollBehaviorBlock struct{}

func (p OverscrollBehaviorBlock) Key() string { return "overscroll-behavior-block" }
func (p OverscrollBehaviorBlock) Sort() int   { return 0 }

// Specifies whether to have scroll chaining or overscroll affordance in the inline direction
type OverscrollBehaviorInline struct{}

func (p OverscrollBehaviorInline) Key() string { return "overscroll-behavior-inline" }
func (p OverscrollBehaviorInline) Sort() int   { return 0 }

// Specifies whether to have scroll chaining or overscroll affordance in x-direction
type OverscrollBehaviorX struct{}

func (p OverscrollBehaviorX) Key() string { return "overscroll-behavior-x" }
func (p OverscrollBehaviorX) Sort() int   { return 0 }

// Specifies whether to have scroll chaining or overscroll affordance in y-directions
type OverscrollBehaviorY struct{}

func (p OverscrollBehaviorY) Key() string { return "overscroll-behavior-y" }
func (p OverscrollBehaviorY) Sort() int   { return 0 }

// A shorthand property for all the padding-* properties
type Padding struct{}

func (p Padding) Key() string { return "padding" }
func (p Padding) Sort() int   { return 1 }

// Specifies the padding in the block direction
type PaddingBlock struct{}

func (p PaddingBlock) Key() string { return "padding-block" }
func (p PaddingBlock) Sort() int   { return 1 }

// Specifies the padding at the end in the block direction
type PaddingBlockEnd struct{}

func (p PaddingBlockEnd) Key() string { return "padding-block-end" }
func (p PaddingBlockEnd) Sort() int   { return 1 }

// Specifies the padding at the start in the block direction
type PaddingBlockStart struct{}

func (p PaddingBlockStart) Key() string { return "padding-block-start" }
func (p PaddingBlockStart) Sort() int   { return 1 }

// Sets the bottom padding of an element
type PaddingBottom struct{}

func (p PaddingBottom) Key() string { return "padding-bottom" }
func (p PaddingBottom) Sort() int   { return 1 }

// Specifies the padding in the inline direction
type PaddingInline struct{}

func (p PaddingInline) Key() string { return "padding-inline" }
func (p PaddingInline) Sort() int   { return 1 }

// Specifies the padding at the end in the inline direction
type PaddingInlineEnd struct{}

func (p PaddingInlineEnd) Key() string { return "padding-inline-end" }
func (p PaddingInlineEnd) Sort() int   { return 1 }

// Specifies the padding at the start in the inline direction
type PaddingInlineStart struct{}

func (p PaddingInlineStart) Key() string { return "padding-inline-start" }
func (p PaddingInlineStart) Sort() int   { return 1 }

// Sets the left padding of an element
type PaddingLeft struct{}

func (p PaddingLeft) Key() string { return "padding-left" }
func (p PaddingLeft) Sort() int   { return 1 }

// Sets the right padding of an element
type PaddingRight struct{}

func (p PaddingRight) Key() string { return "padding-right" }
func (p PaddingRight) Sort() int   { return 1 }

// Sets the top padding of an element
type PaddingTop struct{}

func (p PaddingTop) Key() string { return "padding-top" }
func (p PaddingTop) Sort() int   { return 1 }

// Sets the page-break behavior after an element
type PageBreakAfter struct{}

func (p PageBreakAfter) Key() string { return "page-break-after" }
func (p PageBreakAfter) Sort() int   { return 0 }

// Sets the page-break behavior before an element
type PageBreakBefore struct{}

func (p PageBreakBefore) Key() string { return "page-break-before" }
func (p PageBreakBefore) Sort() int   { return 0 }

// Sets the page-break behavior inside an element
type PageBreakInside struct{}

func (p PageBreakInside) Key() string { return "page-break-inside" }
func (p PageBreakInside) Sort() int   { return 0 }

// Sets the order of how an SVG element or text is painted.
type PaintOrder struct{}

func (p PaintOrder) Key() string { return "paint-order" }
func (p PaintOrder) Sort() int   { return 0 }

// Gives a 3D-positioned element some perspective
type Perspective struct{}

func (p Perspective) Key() string { return "perspective" }
func (p Perspective) Sort() int   { return 0 }

// Defines at which position the user is looking at the 3D-positioned element
type PerspectiveOrigin struct{}

func (p PerspectiveOrigin) Key() string { return "perspective-origin" }
func (p PerspectiveOrigin) Sort() int   { return 0 }

// Specifies align-content and justify-content property values for flexbox and grid layouts
type PlaceContent struct{}

func (p PlaceContent) Key() string { return "place-content" }
func (p PlaceContent) Sort() int   { return 0 }

// Specifies align-items and justify-items property values for grid layouts
type PlaceItems struct{}

func (p PlaceItems) Key() string { return "place-items" }
func (p PlaceItems) Sort() int   { return 0 }

// Specifies align-self and justify-self property values for grid layouts
type PlaceSelf struct{}

func (p PlaceSelf) Key() string { return "place-self" }
func (p PlaceSelf) Sort() int   { return 0 }

// Defines whether or not an element reacts to pointer events
type PointerEvents struct{}

func (p PointerEvents) Key() string { return "pointer-events" }
func (p PointerEvents) Sort() int   { return 0 }

// Specifies the type of positioning method used for an element (static, relative, absolute or fixed)
type Position struct{}

func (p Position) Key() string { return "position" }
func (p Position) Sort() int   { return 0 }

// Sets the type of quotation marks for embedded quotations
type Quotes struct{}

func (p Quotes) Key() string { return "quotes" }
func (p Quotes) Sort() int   { return 0 }

// Defines if (and how) an element is resizable by the user
type Resize struct{}

func (p Resize) Key() string { return "resize" }
func (p Resize) Sort() int   { return 0 }

// Specifies the right position of a positioned element
type Right struct{}

func (p Right) Key() string { return "right" }
func (p Right) Sort() int   { return 0 }

// Specifies the rotation of an element
type Rotate struct{}

func (p Rotate) Key() string { return "rotate" }
func (p Rotate) Sort() int   { return 0 }

// Specifies the gap between the grid rows
type RowGap struct{}

func (p RowGap) Key() string { return "row-gap" }
func (p RowGap) Sort() int   { return 0 }

// Specifies the size of an element by scaling up or down
type Scale struct{}

func (p Scale) Key() string { return "scale" }
func (p Scale) Sort() int   { return 0 }

// Specifies whether to smoothly animate the scroll position in a scrollable box, instead of a straight jump
type ScrollBehavior struct{}

func (p ScrollBehavior) Key() string { return "scroll-behavior" }
func (p ScrollBehavior) Sort() int   { return 0 }

// Specifies the margin between the snap position and the container
type ScrollMargin struct{}

func (p ScrollMargin) Key() string { return "scroll-margin" }
func (p ScrollMargin) Sort() int   { return 0 }

// Specifies the margin between the snap position and the container in the block direction
type ScrollMarginBlock struct{}

func (p ScrollMarginBlock) Key() string { return "scroll-margin-block" }
func (p ScrollMarginBlock) Sort() int   { return 0 }

// Specifies the end margin between the snap position and the container in the block direction
type ScrollMarginBlockEnd struct{}

func (p ScrollMarginBlockEnd) Key() string { return "scroll-margin-block-end" }
func (p ScrollMarginBlockEnd) Sort() int   { return 0 }

// Specifies the start margin between the snap position and the container in the block direction
type ScrollMarginBlockStart struct{}

func (p ScrollMarginBlockStart) Key() string { return "scroll-margin-block-start" }
func (p ScrollMarginBlockStart) Sort() int   { return 0 }

// Specifies the margin between the snap position on the bottom side and the container
type ScrollMarginBottom struct{}

func (p ScrollMarginBottom) Key() string { return "scroll-margin-bottom" }
func (p ScrollMarginBottom) Sort() int   { return 0 }

// Specifies the margin between the snap position and the container in the inline direction
type ScrollMarginInline struct{}

func (p ScrollMarginInline) Key() string { return "scroll-margin-inline" }
func (p ScrollMarginInline) Sort() int   { return 0 }

// Specifies the end margin between the snap position and the container in the inline direction
type ScrollMarginInlineEnd struct{}

func (p ScrollMarginInlineEnd) Key() string { return "scroll-margin-inline-end" }
func (p ScrollMarginInlineEnd) Sort() int   { return 0 }

// Specifies the start margin between the snap position and the container in the inline direction
type ScrollMarginInlineStart struct{}

func (p ScrollMarginInlineStart) Key() string { return "scroll-margin-inline-start" }
func (p ScrollMarginInlineStart) Sort() int   { return 0 }

// Specifies the margin between the snap position on the left side and the container
type ScrollMarginLeft struct{}

func (p ScrollMarginLeft) Key() string { return "scroll-margin-left" }
func (p ScrollMarginLeft) Sort() int   { return 0 }

// Specifies the margin between the snap position on the right side and the container
type ScrollMarginRight struct{}

func (p ScrollMarginRight) Key() string { return "scroll-margin-right" }
func (p ScrollMarginRight) Sort() int   { return 0 }

// Specifies the margin between the snap position on the top side and the container
type ScrollMarginTop struct{}

func (p ScrollMarginTop) Key() string { return "scroll-margin-top" }
func (p ScrollMarginTop) Sort() int   { return 0 }

// Specifies the distance from the container to the snap position on the child elements
type ScrollPadding struct{}

func (p ScrollPadding) Key() string { return "scroll-padding" }
func (p ScrollPadding) Sort() int   { return 0 }

// Specifies the distance in block direction from the container to the snap position on the child elements
type ScrollPaddingBlock struct{}

func (p ScrollPaddingBlock) Key() string { return "scroll-padding-block" }
func (p ScrollPaddingBlock) Sort() int   { return 0 }

// Specifies the distance in block direction from the end of the container to the snap position on the child elements
type ScrollPaddingBlockEnd struct{}

func (p ScrollPaddingBlockEnd) Key() string { return "scroll-padding-block-end" }
func (p ScrollPaddingBlockEnd) Sort() int   { return 0 }

// Specifies the distance in block direction from the start of the container to the snap position on the child elements
type ScrollPaddingBlockStart struct{}

func (p ScrollPaddingBlockStart) Key() string { return "scroll-padding-block-start" }
func (p ScrollPaddingBlockStart) Sort() int   { return 0 }

// Specifies the distance from the bottom of the container to the snap position on the child elements
type ScrollPaddingBottom struct{}

func (p ScrollPaddingBottom) Key() string { return "scroll-padding-bottom" }
func (p ScrollPaddingBottom) Sort() int   { return 0 }

// Specifies the distance in inline direction from the container to the snap position on the child elements
type ScrollPaddingInline struct{}

func (p ScrollPaddingInline) Key() string { return "scroll-padding-inline" }
func (p ScrollPaddingInline) Sort() int   { return 0 }

// Specifies the distance in inline direction from the end of the container to the snap position on the child elements
type ScrollPaddingInlineEnd struct{}

func (p ScrollPaddingInlineEnd) Key() string { return "scroll-padding-inline-end" }
func (p ScrollPaddingInlineEnd) Sort() int   { return 0 }

// Specifies the distance in inline direction from the start of the container to the snap position on the child elements
type ScrollPaddingInlineStart struct{}

func (p ScrollPaddingInlineStart) Key() string { return "scroll-padding-inline-start" }
func (p ScrollPaddingInlineStart) Sort() int   { return 0 }

// Specifies the distance from the left side of the container to the snap position on the child elements
type ScrollPaddingLeft struct{}

func (p ScrollPaddingLeft) Key() string { return "scroll-padding-left" }
func (p ScrollPaddingLeft) Sort() int   { return 0 }

// Specifies the distance from the right side of the container to the snap position on the child elements
type ScrollPaddingRight struct{}

func (p ScrollPaddingRight) Key() string { return "scroll-padding-right" }
func (p ScrollPaddingRight) Sort() int   { return 0 }

// Specifies the distance from the top of the container to the snap position on the child elements
type ScrollPaddingTop struct{}

func (p ScrollPaddingTop) Key() string { return "scroll-padding-top" }
func (p ScrollPaddingTop) Sort() int   { return 0 }

// Specifies where to position elements when the user stops scrolling
type ScrollSnapAlign struct{}

func (p ScrollSnapAlign) Key() string { return "scroll-snap-align" }
func (p ScrollSnapAlign) Sort() int   { return 0 }

// Specifies scroll behaviour after fast swipe on trackpad or touch screen
type ScrollSnapStop struct{}

func (p ScrollSnapStop) Key() string { return "scroll-snap-stop" }
func (p ScrollSnapStop) Sort() int   { return 0 }

// Specifies how snap behaviour should be when scrolling
type ScrollSnapType struct{}

func (p ScrollSnapType) Key() string { return "scroll-snap-type" }
func (p ScrollSnapType) Sort() int   { return 0 }

// Specifies the color of the scrollbar of an element
type ScrollbarColor struct{}

func (p ScrollbarColor) Key() string { return "scrollbar-color" }
func (p ScrollbarColor) Sort() int   { return 0 }

// Specifies the width of a tab character
type TabSize struct{}

func (p TabSize) Key() string { return "tab-size" }
func (p TabSize) Sort() int   { return 0 }

// Defines the algorithm used to lay out table cells, rows, and columns
type TableLayout struct{}

func (p TableLayout) Key() string { return "table-layout" }
func (p TableLayout) Sort() int   { return 0 }

// Specifies the horizontal alignment of text
type TextAlign struct{}

func (p TextAlign) Key() string { return "text-align" }
func (p TextAlign) Sort() int   { return 0 }

// Describes how the last line of a block or a line right before a forced line break is aligned when text-align is "justify"
type TextAlignLast struct{}

func (p TextAlignLast) Key() string { return "text-align-last" }
func (p TextAlignLast) Sort() int   { return 0 }

// Specifies the combination of multiple characters into the space of a single character
type TextCombineUpright struct{}

func (p TextCombineUpright) Key() string { return "text-combine-upright" }
func (p TextCombineUpright) Sort() int   { return 0 }

// Specifies the decoration added to text
type TextDecoration struct{}

func (p TextDecoration) Key() string { return "text-decoration" }
func (p TextDecoration) Sort() int   { return 0 }

// Specifies the color of the text-decoration
type TextDecorationColor struct{}

func (p TextDecorationColor) Key() string { return "text-decoration-color" }
func (p TextDecorationColor) Sort() int   { return 0 }

// Specifies the type of line in a text-decoration
type TextDecorationLine struct{}

func (p TextDecorationLine) Key() string { return "text-decoration-line" }
func (p TextDecorationLine) Sort() int   { return 0 }

// Specifies the style of the line in a text decoration
type TextDecorationStyle struct{}

func (p TextDecorationStyle) Key() string { return "text-decoration-style" }
func (p TextDecorationStyle) Sort() int   { return 0 }

// Specifies the thickness of the decoration line
type TextDecorationThickness struct{}

func (p TextDecorationThickness) Key() string { return "text-decoration-thickness" }
func (p TextDecorationThickness) Sort() int   { return 0 }

// Applies emphasis marks to text
type TextEmphasis struct{}

func (p TextEmphasis) Key() string { return "text-emphasis" }
func (p TextEmphasis) Sort() int   { return 0 }

// Specifies the indentation of the first line in a text-block
type TextIndent struct{}

func (p TextIndent) Key() string { return "text-indent" }
func (p TextIndent) Sort() int   { return 0 }

// Specifies the justification method used when text-align is "justify"
type TextJustify struct{}

func (p TextJustify) Key() string { return "text-justify" }
func (p TextJustify) Sort() int   { return 0 }

// Defines the orientation of characters in a line
type TextOrientation struct{}

func (p TextOrientation) Key() string { return "text-orientation" }
func (p TextOrientation) Sort() int   { return 0 }

// Specifies what should happen when text overflows the containing element
type TextOverflow struct{}

func (p TextOverflow) Key() string { return "text-overflow" }
func (p TextOverflow) Sort() int   { return 0 }

// Adds shadow to text
type TextShadow struct{}

func (p TextShadow) Key() string { return "text-shadow" }
func (p TextShadow) Sort() int   { return 0 }

// Controls the capitalization of text
type TextTransform struct{}

func (p TextTransform) Key() string { return "text-transform" }
func (p TextTransform) Sort() int   { return 0 }

// Specifies the position of the underline which is set using the text-decoration property
type TextUnderlinePosition struct{}

func (p TextUnderlinePosition) Key() string { return "text-underline-position" }
func (p TextUnderlinePosition) Sort() int   { return 0 }

// Specifies the top position of a positioned element
type Top struct{}

func (p Top) Key() string { return "top" }
func (p Top) Sort() int   { return 0 }

// Applies a 2D or 3D transformation to an element
type Transform struct{}

func (p Transform) Key() string { return "transform" }
func (p Transform) Sort() int   { return 0 }

// Allows you to change the position on transformed elements
type TransformOrigin struct{}

func (p TransformOrigin) Key() string { return "transform-origin" }
func (p TransformOrigin) Sort() int   { return 0 }

// Specifies how nested elements are rendered in 3D space
type TransformStyle struct{}

func (p TransformStyle) Key() string { return "transform-style" }
func (p TransformStyle) Sort() int   { return 0 }

// A shorthand property for all the transition-* properties
type Transition struct{}

func (p Transition) Key() string { return "transition" }
func (p Transition) Sort() int   { return 0 }

// Specifies when the transition effect will start
type TransitionDelay struct{}

func (p TransitionDelay) Key() string { return "transition-delay" }
func (p TransitionDelay) Sort() int   { return 0 }

// Specifies how many seconds or milliseconds a transition effect takes to complete
type TransitionDuration struct{}

func (p TransitionDuration) Key() string { return "transition-duration" }
func (p TransitionDuration) Sort() int   { return 0 }

// Specifies the name of the CSS property the transition effect is for
type TransitionProperty struct{}

func (p TransitionProperty) Key() string { return "transition-property" }
func (p TransitionProperty) Sort() int   { return 0 }

// Specifies the speed curve of the transition effect
type TransitionTimingFunction struct{}

func (p TransitionTimingFunction) Key() string { return "transition-timing-function" }
func (p TransitionTimingFunction) Sort() int   { return 0 }

// Specifies the position of an element
type Translate struct{}

func (p Translate) Key() string { return "translate" }
func (p Translate) Sort() int   { return 0 }

// Used together with the direction property to set or return whether the text should be overridden to support multiple languages in the same document
type UnicodeBidi struct{}

func (p UnicodeBidi) Key() string { return "unicode-bidi" }
func (p UnicodeBidi) Sort() int   { return 0 }

// Specifies whether the text of an element can be selected
type UserSelect struct{}

func (p UserSelect) Key() string { return "user-select" }
func (p UserSelect) Sort() int   { return 0 }

// Sets the vertical alignment of an element
type VerticalAlign struct{}

func (p VerticalAlign) Key() string { return "vertical-align" }
func (p VerticalAlign) Sort() int   { return 0 }

// Specifies whether or not an element is visible
type Visibility struct{}

func (p Visibility) Key() string { return "visibility" }
func (p Visibility) Sort() int   { return 0 }

// Specifies how white-space inside an element is handled
type WhiteSpace struct{}

func (p WhiteSpace) Key() string { return "white-space" }
func (p WhiteSpace) Sort() int   { return 0 }

// Sets the minimum number of lines that must be left at the top of a page or column
type Widows struct{}

func (p Widows) Key() string { return "widows" }
func (p Widows) Sort() int   { return 0 }

// Sets the width of an element
type Width struct{}

func (p Width) Key() string { return "width" }
func (p Width) Sort() int   { return 0 }

// Specifies how words should break when reaching the end of a line
type WordBreak struct{}

func (p WordBreak) Key() string { return "word-break" }
func (p WordBreak) Sort() int   { return 0 }

// Increases or decreases the space between words in a text
type WordSpacing struct{}

func (p WordSpacing) Key() string { return "word-spacing" }
func (p WordSpacing) Sort() int   { return 0 }

// Allows long, unbreakable words to be broken and wrap to the next line
type WordWrap struct{}

func (p WordWrap) Key() string { return "word-wrap" }
func (p WordWrap) Sort() int   { return 0 }

// Specifies whether lines of text are laid out horizontally or vertically
type WritingMode struct{}

func (p WritingMode) Key() string { return "writing-mode" }
func (p WritingMode) Sort() int   { return 0 }

// Sets the stack order of a positioned element
type ZIndex struct{}

func (p ZIndex) Key() string { return "z-index" }
func (p ZIndex) Sort() int   { return 0 }
