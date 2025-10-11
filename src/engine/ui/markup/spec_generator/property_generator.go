/******************************************************************************/
/* property_generator.go                                                      */
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

package spec_generator

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type propertyData struct {
	name        string
	description string
}

func (d propertyData) SafeName() string {
	return strings.ReplaceAll(d.name, "@", "")
}

func (d propertyData) StructName() string {
	titleCase := cases.Title(language.English)
	return strings.ReplaceAll(titleCase.String(strings.ReplaceAll(d.SafeName(), "-", " ")), " ", "")
}

var genProps = []propertyData{
	{"accent-color", "Specifies an accent color for user-interface controls"},
	{"align-content", "Specifies the alignment between the lines inside a flexible container when the items do not use all available space"},
	{"align-items", "Specifies the alignment for items inside a flexible container"},
	{"align-self", "Specifies the alignment for selected items inside a flexible container"},
	{"all", "Resets all properties (except unicode-bidi and direction)"},
	{"animation", "A shorthand property for all the animation-* properties"},
	{"animation-delay", "Specifies a delay for the start of an animation"},
	{"animation-direction", "Specifies whether an animation should be played forwards, backwards or in alternate cycles"},
	{"animation-duration", "Specifies how long an animation should take to complete one cycle"},
	{"animation-fill-mode", "Specifies a style for the element when the animation is not playing (before it starts, after it ends, or both)"},
	{"animation-iteration-count", "Specifies the number of times an animation should be played"},
	{"animation-name", "Specifies a name for the @keyframes animation"},
	{"animation-play-state", "Specifies whether the animation is running or paused"},
	{"animation-timing-function", "Specifies the speed curve of an animation"},
	{"aspect-ratio", "Specifies preferred aspect ratio of an element"},
	{"backdrop-filter", "Defines a graphical effect to the area behind an element"},
	{"backface-visibility", "Defines whether or not the back face of an element should be visible when facing the user"},
	{"background", "A shorthand property for all the background-* properties"},
	{"background-attachment", "Sets whether a background image scrolls with the rest of the page, or is fixed"},
	{"background-blend-mode", "Specifies the blending mode of each background layer (color/image)"},
	{"background-clip", "Defines how far the background (color or image) should extend within an element"},
	{"background-color", "Specifies the background color of an element"},
	{"background-image", "Specifies one or more background images for an element"},
	{"background-origin", "Specifies the origin position of a background image"},
	{"background-position", "Specifies the position of a background image"},
	{"background-position-x", "Specifies the position of a background image on x-axis"},
	{"background-position-y", "Specifies the position of a background image on y-axis"},
	{"background-repeat", "Sets if/how a background image will be repeated"},
	{"background-size", "Specifies the size of the background images"},
	{"block-size", "Specifies the size of an element in block direction"},
	{"border", "A shorthand property for border-width, border-style and border-color"},
	{"border-block", "A shorthand property for border-block-width, border-block-style and border-block-color"},
	{"border-block-color", "Sets the color of the borders at start and end in the block direction"},
	{"border-block-end-color", "Sets the color of the border at the end in the block direction"},
	{"border-block-end-style", "Sets the style of the border at the end in the block direction"},
	{"border-block-end-width", "Sets the width of the border at the end in the block direction"},
	{"border-block-start-color", "Sets the color of the border at the start in the block direction"},
	{"border-block-start-style", "Sets the style of the border at the start in the block direction"},
	{"border-block-start-width", "Sets the width of the border at the start in the block direction"},
	{"border-block-style", "Sets the style of the borders at start and end in the block direction"},
	{"border-block-width", "Sets the width of the borders at start and end in the block direction"},
	{"border-bottom", "A shorthand property for border-bottom-width, border-bottom-style and border-bottom-color"},
	{"border-bottom-color", "Sets the color of the bottom border"},
	{"border-bottom-left-radius", "Defines the radius of the border of the bottom-left corner"},
	{"border-bottom-right-radius", "Defines the radius of the border of the bottom-right corner"},
	{"border-bottom-style", "Sets the style of the bottom border"},
	{"border-bottom-width", "Sets the width of the bottom border"},
	{"border-collapse", "Sets whether table borders should collapse into a single border or be separated"},
	{"border-color", "Sets the color of the four borders"},
	{"border-image", "A shorthand property for all the border-image-* properties"},
	{"border-image-outset", "Specifies the amount by which the border image area extends beyond the border box"},
	{"border-image-repeat", "Specifies whether the border image should be repeated, rounded or stretched"},
	{"border-image-slice", "Specifies how to slice the border image"},
	{"border-image-source", "Specifies the path to the image to be used as a border"},
	{"border-image-width", "Specifies the width of the border image"},
	{"border-inline", "A shorthand property for border-inline-width, border-inline-style and border-inline-color"},
	{"border-inline-color", "Sets the color of the borders at start and end in the inline direction"},
	{"border-inline-end-color", "Sets the color of the border at the end in the inline direction"},
	{"border-inline-end-style", "Sets the style of the border at the end in the inline direction"},
	{"border-inline-end-width", "Sets the width of the border at the end in the inline direction"},
	{"border-inline-start-color", "Sets the color of the border at the start in the inline direction"},
	{"border-inline-start-style", "Sets the style of the border at the start in the inline direction"},
	{"border-inline-start-width", "Sets the width of the border at the start in the inline direction"},
	{"border-inline-style", "Sets the style of the borders at start and end in the inline direction"},
	{"border-inline-width", "Sets the width of the borders at start and end in the inline direction"},
	{"border-left", "A shorthand property for all the border-left-* properties"},
	{"border-left-color", "Sets the color of the left border"},
	{"border-left-style", "Sets the style of the left border"},
	{"border-left-width", "Sets the width of the left border"},
	{"border-radius", "A shorthand property for the four border-*-radius properties"},
	{"border-right", "A shorthand property for all the border-right-* properties"},
	{"border-right-color", "Sets the color of the right border"},
	{"border-right-style", "Sets the style of the right border"},
	{"border-right-width", "Sets the width of the right border"},
	{"border-spacing", "Sets the distance between the borders of adjacent cells"},
	{"border-style", "Sets the style of the four borders"},
	{"border-top", "A shorthand property for border-top-width, border-top-style and border-top-color"},
	{"border-top-color", "Sets the color of the top border"},
	{"border-top-left-radius", "Defines the radius of the border of the top-left corner"},
	{"border-top-right-radius", "Defines the radius of the border of the top-right corner"},
	{"border-top-style", "Sets the style of the top border"},
	{"border-top-width", "Sets the width of the top border"},
	{"border-width", "Sets the width of the four borders"},
	{"bottom", "Sets the elements position, from the bottom of its parent element"},
	{"box-decoration-break", "Sets the behavior of the background and border of an element at page-break, or, for in-line elements, at line-break."},
	{"box-reflect", "The box-reflect property is used to create a reflection of an element."},
	{"box-shadow", "Attaches one or more shadows to an element"},
	{"box-sizing", "Defines how the width and height of an element are calculated: should they include padding and borders, or not"},
	{"break-after", "Specifies whether or not a page-, column-, or region-break should occur after the specified element"},
	{"break-before", "Specifies whether or not a page-, column-, or region-break should occur before the specified element"},
	{"break-inside", "Specifies whether or not a page-, column-, or region-break should occur inside the specified element"},
	{"caption-side", "Specifies the placement of a table caption"},
	{"caret-color", "Specifies the color of the cursor (caret) in inputs, textareas, or any element that is editable"},
	{"@charset", "Specifies the character encoding used in the style sheet"},
	{"clear", "Specifies what should happen with the element that is next to a floating element"},
	{"clip", "Clips an absolutely positioned element"},
	{"color", "Sets the color of text"},
	{"column-count", "Specifies the number of columns an element should be divided into"},
	{"column-fill", "Specifies how to fill columns, balanced or not"},
	{"column-gap", "Specifies the gap between the columns"},
	{"column-rule", "A shorthand property for all the column-rule-* properties"},
	{"column-rule-color", "Specifies the color of the rule between columns"},
	{"column-rule-style", "Specifies the style of the rule between columns"},
	{"column-rule-width", "Specifies the width of the rule between columns"},
	{"column-span", "Specifies how many columns an element should span across"},
	{"column-width", "Specifies the column width"},
	{"columns", "A shorthand property for column-width and column-count"},
	{"content", "Used with the :before and :after pseudo-elements, to insert generated content"},
	{"counter-increment", "Increases or decreases the value of one or more CSS counters"},
	{"counter-reset", "Creates or resets one or more CSS counters"},
	{"cursor", "Specifies the mouse cursor to be displayed when pointing over an element"},
	{"direction", "Specifies the text direction/writing direction"},
	{"display", "Specifies how a certain HTML element should be displayed"},
	{"empty-cells", "Specifies whether or not to display borders and background on empty cells in a table"},
	{"filter", "Defines effects (e.g. blurring or color shifting) on an element before the element is displayed"},
	{"flex", "A shorthand property for the flex-grow, flex-shrink, and the flex-basis properties"},
	{"flex-basis", "Specifies the initial length of a flexible item"},
	{"flex-direction", "Specifies the direction of the flexible items"},
	{"flex-flow", "A shorthand property for the flex-direction and the flex-wrap properties"},
	{"flex-grow", "Specifies how much the item will grow relative to the rest"},
	{"flex-shrink", "Specifies how the item will shrink relative to the rest"},
	{"flex-wrap", "Specifies whether the flexible items should wrap or not"},
	{"float", "Specifies whether an element should float to the left, right, or not at all"},
	{"font", "A shorthand property for the font-style, font-variant, font-weight, font-size/line-height, and the font-family properties"},
	{"@font-face", `A rule that allows websites to download and use fonts other than the "web-safe" fonts`},
	{"font-family", "Specifies the font family for text"},
	{"font-feature-settings", "Allows control over advanced typographic features in OpenType fonts"},
	{"@font-feature-values", "Allows authors to use a common name in font-variant-alternate for feature activated differently in OpenType"},
	{"font-kerning", "Controls the usage of the kerning information (how letters are spaced)"},
	{"font-language-override", "Controls the usage of language-specific glyphs in a typeface"},
	{"font-size", "Specifies the font size of text"},
	{"font-size-adjust", "Preserves the readability of text when font fallback occurs"},
	{"font-stretch", "Selects a normal, condensed, or expanded face from a font family"},
	{"font-style", "Specifies the font style for text"},
	{"font-synthesis", "Controls which missing typefaces (bold or italic) may be synthesized by the browser"},
	{"font-variant", "Specifies whether or not a text should be displayed in a small-caps font"},
	{"font-variant-alternates", "Controls the usage of alternate glyphs associated to alternative names defined in @font-feature-values"},
	{"font-variant-caps", "Controls the usage of alternate glyphs for capital letters"},
	{"font-variant-east-asian", "Controls the usage of alternate glyphs for East Asian scripts (e.g Japanese and Chinese)"},
	{"font-variant-ligatures", "Controls which ligatures and contextual forms are used in textual content of the elements it applies to"},
	{"font-variant-numeric", "Controls the usage of alternate glyphs for numbers, fractions, and ordinal markers"},
	{"font-variant-position", "Controls the usage of alternate glyphs of smaller size positioned as superscript or subscript regarding the baseline of the font"},
	{"font-weight", "Specifies the weight of a font"},
	{"gap", "A shorthand property for the row-gap and the column-gap properties"},
	{"grid", "A shorthand property for the grid-template-rows, grid-template-columns, grid-template-areas, grid-auto-rows, grid-auto-columns, and the grid-auto-flow properties"},
	{"grid-area", "Either specifies a name for the grid item, or this property is a shorthand property for the grid-row-start, grid-column-start, grid-row-end, and grid-column-end properties"},
	{"grid-auto-columns", "Specifies a default column size"},
	{"grid-auto-flow", "Specifies how auto-placed items are inserted in the grid"},
	{"grid-auto-rows", "Specifies a default row size"},
	{"grid-column", "A shorthand property for the grid-column-start and the grid-column-end properties"},
	{"grid-column-end", "Specifies where to end the grid item"},
	{"grid-column-gap", "Specifies the size of the gap between columns"},
	{"grid-column-start", "Specifies where to start the grid item"},
	{"grid-gap", "A shorthand property for the grid-row-gap and grid-column-gap properties"},
	{"grid-row", "A shorthand property for the grid-row-start and the grid-row-end properties"},
	{"grid-row-end", "Specifies where to end the grid item"},
	{"grid-row-gap", "Specifies the size of the gap between rows"},
	{"grid-row-start", "Specifies where to start the grid item"},
	{"grid-template", "A shorthand property for the grid-template-rows, grid-template-columns and grid-areas properties"},
	{"grid-template-areas", "Specifies how to display columns and rows, using named grid items"},
	{"grid-template-columns", "Specifies the size of the columns, and how many columns in a grid layout"},
	{"grid-template-rows", "Specifies the size of the rows in a grid layout"},
	{"hanging-punctuation", "Specifies whether a punctuation character may be placed outside the line box"},
	{"height", "Sets the height of an element"},
	{"hyphens", "Sets how to split words to improve the layout of paragraphs"},
	{"image-rendering", "Specifies the type of algorithm to use for image scaling"},
	{"@import", "Allows you to import a style sheet into another style sheet"},
	{"inline-size", "Specifies the size of an element in the inline direction"},
	{"inset", "Specifies the distance between an element and the parent element"},
	{"inset-block", "Specifies the distance between an element and the parent element in the block direction"},
	{"inset-block-end", "Specifies the distance between the end of an element and the parent element in the block direction"},
	{"inset-block-start", "Specifies the distance between the start of an element and the parent element in the block direction"},
	{"inset-inline", "Specifies the distance between an element and the parent element in the inline direction"},
	{"inset-inline-end", "Specifies the distance between the end of an element and the parent element in the inline direction"},
	{"inset-inline-start", "Specifies the distance between the start of an element and the parent element in the inline direction"},
	{"isolation", "Defines whether an element must create a new stacking content"},
	{"justify-content", "Specifies the alignment between the items inside a flexible container when the items do not use all available space"},
	{"justify-items", "Is set on the grid container. Specifies the alignment of grid items in the inline direction"},
	{"justify-self", "Is set on the grid item. Specifies the alignment of the grid item in the inline direction"},
	{"@keyframes", "Specifies the animation code"},
	{"left", "Specifies the left position of a positioned element"},
	{"letter-spacing", "Increases or decreases the space between characters in a text"},
	{"line-break", "Specifies how/if to break lines"},
	{"line-height", "Sets the line height"},
	{"list-style", "Sets all the properties for a list in one declaration"},
	{"list-style-image", "Specifies an image as the list-item marker"},
	{"list-style-position", "Specifies the position of the list-item markers (bullet points)"},
	{"list-style-type", "Specifies the type of list-item marker"},
	{"margin", "Sets all the margin properties in one declaration"},
	{"margin-block", "Specifies the margin in the block direction"},
	{"margin-block-end", "Specifies the margin at the end in the block direction"},
	{"margin-block-start", "Specifies the margin at the start in the block direction"},
	{"margin-bottom", "Sets the bottom margin of an element"},
	{"margin-inline", "Specifies the margin in the inline direction"},
	{"margin-inline-end", "Specifies the margin at the end in the inline direction"},
	{"margin-inline-start", "Specifies the margin at the start in the inline direction"},
	{"margin-left", "Sets the left margin of an element"},
	{"margin-right", "Sets the right margin of an element"},
	{"margin-top", "Sets the top margin of an element"},
	{"mask", "Hides parts of an element by masking or clipping an image at specific places"},
	{"mask-clip", "Specifies the mask area"},
	{"mask-composite", "Represents a compositing operation used on the current mask layer with the mask layers below it"},
	{"mask-image", "Specifies an image to be used as a mask layer for an element"},
	{"mask-mode", "Specifies whether the mask layer image is treated as a luminance mask or as an alpha mask"},
	{"mask-origin", "Specifies the origin position (the mask position area) of a mask layer image"},
	{"mask-position", "Sets the starting position of a mask layer image (relative to the mask position area)"},
	{"mask-repeat", "Specifies how the mask layer image is repeated"},
	{"mask-size", "Specifies the size of a mask layer image"},
	{"mask-type", "Specifies whether an SVG <mask> element is treated as a luminance mask or as an alpha mask"},
	{"max-height", "Sets the maximum height of an element"},
	{"max-width", "Sets the maximum width of an element"},
	{"@media", "Sets the style rules for different media types/devices/sizes"},
	{"max-block-size", "Sets the maximum size of an element in the block direction"},
	{"max-inline-size", "Sets the maximum size of an element in the inline direction"},
	{"min-block-size", "Sets the minimum size of an element in the block direction"},
	{"min-inline-size", "Sets the minimum size of an element in the inline direction"},
	{"min-height", "Sets the minimum height of an element"},
	{"min-width", "Sets the minimum width of an element"},
	{"mix-blend-mode", "Specifies how an element's content should blend with its direct parent background"},
	{"object-fit", "Specifies how the contents of a replaced element should be fitted to the box established by its used height and width"},
	{"object-position", "Specifies the alignment of the replaced element inside its box"},
	{"offset", "Is a shorthand, and specifies how to animate an element along a path"},
	{"offset-anchor", "Specifies a point on an element that is fixed to the path it is animated along"},
	{"offset-distance", "Specifies the position along a path where an animated element is placed"},
	{"offset-path", "Specifies the path an element is animated along"},
	{"offset-rotate", "Specifies rotation of an element as it is animated along a path"},
	{"opacity", "Sets the opacity level for an element"},
	{"order", "Sets the order of the flexible item, relative to the rest"},
	{"orphans", "Sets the minimum number of lines that must be left at the bottom of a page or column"},
	{"outline", "A shorthand property for the outline-width, outline-style, and the outline-color properties"},
	{"outline-color", "Sets the color of an outline"},
	{"outline-offset", "Offsets an outline, and draws it beyond the border edge"},
	{"outline-style", "Sets the style of an outline"},
	{"outline-width", "Sets the width of an outline"},
	{"overflow", "Specifies what happens if content overflows an element's box"},
	{"overflow-anchor", "Specifies whether or not content in viewable area in a scrollable contianer should be pushed down when new content is loaded above"},
	{"overflow-wrap", "Specifies whether or not the browser can break lines with long words, if they overflow the container"},
	{"overflow-x", "Specifies whether or not to clip the left/right edges of the content, if it overflows the element's content area"},
	{"overflow-y", "Specifies whether or not to clip the top/bottom edges of the content, if it overflows the element's content area"},
	{"overscroll-behavior", "Specifies whether to have scroll chaining or overscroll affordance in x- and y-directions"},
	{"overscroll-behavior-block", "Specifies whether to have scroll chaining or overscroll affordance in the block direction"},
	{"overscroll-behavior-inline", "Specifies whether to have scroll chaining or overscroll affordance in the inline direction"},
	{"overscroll-behavior-x", "Specifies whether to have scroll chaining or overscroll affordance in x-direction"},
	{"overscroll-behavior-y", "Specifies whether to have scroll chaining or overscroll affordance in y-directions"},
	{"padding", "A shorthand property for all the padding-* properties"},
	{"padding-block", "Specifies the padding in the block direction"},
	{"padding-block-end", "Specifies the padding at the end in the block direction"},
	{"padding-block-start", "Specifies the padding at the start in the block direction"},
	{"padding-bottom", "Sets the bottom padding of an element"},
	{"padding-inline", "Specifies the padding in the inline direction"},
	{"padding-inline-end", "Specifies the padding at the end in the inline direction"},
	{"padding-inline-start", "Specifies the padding at the start in the inline direction"},
	{"padding-left", "Sets the left padding of an element"},
	{"padding-right", "Sets the right padding of an element"},
	{"padding-top", "Sets the top padding of an element"},
	{"page-break-after", "Sets the page-break behavior after an element"},
	{"page-break-before", "Sets the page-break behavior before an element"},
	{"page-break-inside", "Sets the page-break behavior inside an element"},
	{"paint-order", "Sets the order of how an SVG element or text is painted."},
	{"perspective", "Gives a 3D-positioned element some perspective"},
	{"perspective-origin", "Defines at which position the user is looking at the 3D-positioned element"},
	{"place-content", "Specifies align-content and justify-content property values for flexbox and grid layouts"},
	{"place-items", "Specifies align-items and justify-items property values for grid layouts"},
	{"place-self", "Specifies align-self and justify-self property values for grid layouts"},
	{"pointer-events", "Defines whether or not an element reacts to pointer events"},
	{"position", "Specifies the type of positioning method used for an element (static, relative, absolute or fixed)"},
	{"quotes", "Sets the type of quotation marks for embedded quotations"},
	{"resize", "Defines if (and how) an element is resizable by the user"},
	{"right", "Specifies the right position of a positioned element"},
	{"rotate", "Specifies the rotation of an element"},
	{"row-gap", "Specifies the gap between the grid rows"},
	{"scale", "Specifies the size of an element by scaling up or down"},
	{"scroll-behavior", "Specifies whether to smoothly animate the scroll position in a scrollable box, instead of a straight jump"},
	{"scroll-margin", "Specifies the margin between the snap position and the container"},
	{"scroll-margin-block", "Specifies the margin between the snap position and the container in the block direction"},
	{"scroll-margin-block-end", "Specifies the end margin between the snap position and the container in the block direction"},
	{"scroll-margin-block-start", "Specifies the start margin between the snap position and the container in the block direction"},
	{"scroll-margin-bottom", "Specifies the margin between the snap position on the bottom side and the container"},
	{"scroll-margin-inline", "Specifies the margin between the snap position and the container in the inline direction"},
	{"scroll-margin-inline-end", "Specifies the end margin between the snap position and the container in the inline direction"},
	{"scroll-margin-inline-start", "Specifies the start margin between the snap position and the container in the inline direction"},
	{"scroll-margin-left", "Specifies the margin between the snap position on the left side and the container"},
	{"scroll-margin-right", "Specifies the margin between the snap position on the right side and the container"},
	{"scroll-margin-top", "Specifies the margin between the snap position on the top side and the container"},
	{"scroll-padding", "Specifies the distance from the container to the snap position on the child elements"},
	{"scroll-padding-block", "Specifies the distance in block direction from the container to the snap position on the child elements"},
	{"scroll-padding-block-end", "Specifies the distance in block direction from the end of the container to the snap position on the child elements"},
	{"scroll-padding-block-start", "Specifies the distance in block direction from the start of the container to the snap position on the child elements"},
	{"scroll-padding-bottom", "Specifies the distance from the bottom of the container to the snap position on the child elements"},
	{"scroll-padding-inline", "Specifies the distance in inline direction from the container to the snap position on the child elements"},
	{"scroll-padding-inline-end", "Specifies the distance in inline direction from the end of the container to the snap position on the child elements"},
	{"scroll-padding-inline-start", "Specifies the distance in inline direction from the start of the container to the snap position on the child elements"},
	{"scroll-padding-left", "Specifies the distance from the left side of the container to the snap position on the child elements"},
	{"scroll-padding-right", "Specifies the distance from the right side of the container to the snap position on the child elements"},
	{"scroll-padding-top", "Specifies the distance from the top of the container to the snap position on the child elements"},
	{"scroll-snap-align", "Specifies where to position elements when the user stops scrolling"},
	{"scroll-snap-stop", "Specifies scroll behaviour after fast swipe on trackpad or touch screen"},
	{"scroll-snap-type", "Specifies how snap behaviour should be when scrolling"},
	{"scrollbar-color", "Specifies the color of the scrollbar of an element"},
	{"tab-size", "Specifies the width of a tab character"},
	{"table-layout", "Defines the algorithm used to lay out table cells, rows, and columns"},
	{"text-align", "Specifies the horizontal alignment of text"},
	{"text-align-last", `Describes how the last line of a block or a line right before a forced line break is aligned when text-align is "justify"`},
	{"text-combine-upright", "Specifies the combination of multiple characters into the space of a single character"},
	{"text-decoration", "Specifies the decoration added to text"},
	{"text-decoration-color", "Specifies the color of the text-decoration"},
	{"text-decoration-line", "Specifies the type of line in a text-decoration"},
	{"text-decoration-style", "Specifies the style of the line in a text decoration"},
	{"text-decoration-thickness", "Specifies the thickness of the decoration line"},
	{"text-emphasis", "Applies emphasis marks to text"},
	{"text-indent", "Specifies the indentation of the first line in a text-block"},
	{"text-justify", `Specifies the justification method used when text-align is "justify"`},
	{"text-orientation", "Defines the orientation of characters in a line"},
	{"text-overflow", "Specifies what should happen when text overflows the containing element"},
	{"text-shadow", "Adds shadow to text"},
	{"text-transform", "Controls the capitalization of text"},
	{"text-underline-position", "Specifies the position of the underline which is set using the text-decoration property"},
	{"top", "Specifies the top position of a positioned element"},
	{"transform", "Applies a 2D or 3D transformation to an element"},
	{"transform-origin", "Allows you to change the position on transformed elements"},
	{"transform-style", "Specifies how nested elements are rendered in 3D space"},
	{"transition", "A shorthand property for all the transition-* properties"},
	{"transition-delay", "Specifies when the transition effect will start"},
	{"transition-duration", "Specifies how many seconds or milliseconds a transition effect takes to complete"},
	{"transition-property", "Specifies the name of the CSS property the transition effect is for"},
	{"transition-timing-function", "Specifies the speed curve of the transition effect"},
	{"translate", "Specifies the position of an element"},
	{"unicode-bidi", "Used together with the direction property to set or return whether the text should be overridden to support multiple languages in the same document"},
	{"user-select", "Specifies whether the text of an element can be selected"},
	{"vertical-align", "Sets the vertical alignment of an element"},
	{"visibility", "Specifies whether or not an element is visible"},
	{"white-space", "Specifies how white-space inside an element is handled"},
	{"widows", "Sets the minimum number of lines that must be left at the top of a page or column"},
	{"width", "Sets the width of an element"},
	{"word-break", "Specifies how words should break when reaching the end of a line"},
	{"word-spacing", "Increases or decreases the space between words in a text"},
	{"word-wrap", "Allows long, unbreakable words to be broken and wrap to the next line"},
	{"writing-mode", "Specifies whether lines of text are laid out horizontally or vertically"},
	{"z-index", "Sets the stack order of a positioned element"},
}

func writePropertyFile() error {
	if err := writeBaseFile(propFolder); err != nil {
		return err
	}
	pf, err := os.Create(propFolder + "/css_property.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package properties

import (
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/css/functions"
	"kaiju/engine/ui/markup/markup"
	"kaiju/engine/ui"
)

type Property interface {
	Key() string
	Process(panel *ui.Panel, elm document.DocumentElement, values []rules.PropertyValue, host *engine.Host) error
}

var PropertyMap = map[string]Property{
`)
	for _, p := range genProps {
		pf.WriteString(fmt.Sprintf(`	"%s": %s{},`, p.SafeName(), p.StructName()))
		pf.WriteString("\n")
	}
	pf.WriteString(`}

func runValueFuncs(panel *ui.Panel, values []rules.PropertyValue) []error {
	problems := make([]error, 0)
	for i, v := range values {
		if v.IsFunction() {
			if f, ok := functions.FunctionMap[v.Str]; ok {
				if res, err := f.Process(panel, v); err == nil {
					values[i].Str = res
					values[i].Args = values[i].Args[:0]
				} else {
					problems = append(problems, err)
				}
			}
		}
	}
	return problems
}
`)
	return nil
}

func writeProperties() error {
	pf, err := os.Create(propFolder + "/css_property_types.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package properties
`)
	for _, p := range genProps {
		pf.WriteString(fmt.Sprintf(`
// %s
type %s struct{}

func (p %s) Key() string { return "%s" }
`, p.description, p.StructName(), p.StructName(), p.SafeName()))
	}
	for _, p := range genProps {
		fName := propFolder + "/css_" + strings.ReplaceAll(p.SafeName(), "-", "_") + ".go"
		if _, err := os.Stat(fName); err != nil {
			if os.IsNotExist(err) {
				f, err := os.Create(fName)
				if err != nil {
					return err
				}
				defer f.Close()
				f.WriteString(fmt.Sprintf(`package properties

import (
	"errors"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/markup"
)

func (p %s) Process(panel *ui.Panel, elm document.DocumentElement, values []rules.PropertyValue, host *engine.Host) error {
	problems := []error{errors.New("%s not implemented")}
	return problems[0]
}
`, p.StructName(), p.StructName()))
			}
		}
	}
	return nil
}
