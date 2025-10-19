/******************************************************************************/
/* html_element_types.go                                                      */
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

package elements

type A struct{}

func (p A) Key() string { return "a" }

type Abbr struct{}

func (p Abbr) Key() string { return "abbr" }

type Address struct{}

func (p Address) Key() string { return "address" }

type Area struct{}

func (p Area) Key() string { return "area" }

type Article struct{}

func (p Article) Key() string { return "article" }

type Aside struct{}

func (p Aside) Key() string { return "aside" }

type Audio struct{}

func (p Audio) Key() string { return "audio" }

type B struct{}

func (p B) Key() string { return "b" }

type Base struct{}

func (p Base) Key() string { return "base" }

type Bdi struct{}

func (p Bdi) Key() string { return "bdi" }

type Bdo struct{}

func (p Bdo) Key() string { return "bdo" }

type Blockquote struct{}

func (p Blockquote) Key() string { return "blockquote" }

type Body struct{}

func (p Body) Key() string { return "body" }

type Br struct{}

func (p Br) Key() string { return "br" }

type Button struct{}

func (p Button) Key() string { return "button" }

type Canvas struct{}

func (p Canvas) Key() string { return "canvas" }

type Caption struct{}

func (p Caption) Key() string { return "caption" }

type Cite struct{}

func (p Cite) Key() string { return "cite" }

type Code struct{}

func (p Code) Key() string { return "code" }

type Col struct{}

func (p Col) Key() string { return "col" }

type Colgroup struct{}

func (p Colgroup) Key() string { return "colgroup" }

type Data struct{}

func (p Data) Key() string { return "data" }

type Datalist struct{}

func (p Datalist) Key() string { return "datalist" }

type Dd struct{}

func (p Dd) Key() string { return "dd" }

type Del struct{}

func (p Del) Key() string { return "del" }

type Details struct{}

func (p Details) Key() string { return "details" }

type Dfn struct{}

func (p Dfn) Key() string { return "dfn" }

type Dialog struct{}

func (p Dialog) Key() string { return "dialog" }

type Div struct{}

func (p Div) Key() string { return "div" }

type Dl struct{}

func (p Dl) Key() string { return "dl" }

type Dt struct{}

func (p Dt) Key() string { return "dt" }

type Em struct{}

func (p Em) Key() string { return "em" }

type Embed struct{}

func (p Embed) Key() string { return "embed" }

type Fieldset struct{}

func (p Fieldset) Key() string { return "fieldset" }

type Figcaption struct{}

func (p Figcaption) Key() string { return "figcaption" }

type Figure struct{}

func (p Figure) Key() string { return "figure" }

type Footer struct{}

func (p Footer) Key() string { return "footer" }

type Form struct{}

func (p Form) Key() string { return "form" }

type H1 struct{}

func (p H1) Key() string { return "h1" }

type H2 struct{}

func (p H2) Key() string { return "h2" }

type H3 struct{}

func (p H3) Key() string { return "h3" }

type H4 struct{}

func (p H4) Key() string { return "h4" }

type H5 struct{}

func (p H5) Key() string { return "h5" }

type H6 struct{}

func (p H6) Key() string { return "h6" }

type Head struct{}

func (p Head) Key() string { return "head" }

type Header struct{}

func (p Header) Key() string { return "header" }

type Hgroup struct{}

func (p Hgroup) Key() string { return "hgroup" }

type Hr struct{}

func (p Hr) Key() string { return "hr" }

type Html struct{}

func (p Html) Key() string { return "html" }

type I struct{}

func (p I) Key() string { return "i" }

type Iframe struct{}

func (p Iframe) Key() string { return "iframe" }

type Img struct{}

func (p Img) Key() string { return "img" }

type Input struct{}

func (p Input) Key() string { return "input" }

type Ins struct{}

func (p Ins) Key() string { return "ins" }

type Kbd struct{}

func (p Kbd) Key() string { return "kbd" }

type Label struct{}

func (p Label) Key() string { return "label" }

type Legend struct{}

func (p Legend) Key() string { return "legend" }

type Li struct{}

func (p Li) Key() string { return "li" }

type Link struct{}

func (p Link) Key() string { return "link" }

type Main struct{}

func (p Main) Key() string { return "main" }

type Map struct{}

func (p Map) Key() string { return "map" }

type Mark struct{}

func (p Mark) Key() string { return "mark" }

type Menu struct{}

func (p Menu) Key() string { return "menu" }

type Meta struct{}

func (p Meta) Key() string { return "meta" }

type Meter struct{}

func (p Meter) Key() string { return "meter" }

type Nav struct{}

func (p Nav) Key() string { return "nav" }

type Noscript struct{}

func (p Noscript) Key() string { return "noscript" }

type Object struct{}

func (p Object) Key() string { return "object" }

type Ol struct{}

func (p Ol) Key() string { return "ol" }

type Optgroup struct{}

func (p Optgroup) Key() string { return "optgroup" }

type Option struct{}

func (p Option) Key() string { return "option" }

type Output struct{}

func (p Output) Key() string { return "output" }

type P struct{}

func (p P) Key() string { return "p" }

type Picture struct{}

func (p Picture) Key() string { return "picture" }

type Pre struct{}

func (p Pre) Key() string { return "pre" }

type Progress struct{}

func (p Progress) Key() string { return "progress" }

type Q struct{}

func (p Q) Key() string { return "q" }

type Rp struct{}

func (p Rp) Key() string { return "rp" }

type Rt struct{}

func (p Rt) Key() string { return "rt" }

type Ruby struct{}

func (p Ruby) Key() string { return "ruby" }

type S struct{}

func (p S) Key() string { return "s" }

type Samp struct{}

func (p Samp) Key() string { return "samp" }

type Script struct{}

func (p Script) Key() string { return "script" }

type Search struct{}

func (p Search) Key() string { return "search" }

type Section struct{}

func (p Section) Key() string { return "section" }

type Select struct{}

func (p Select) Key() string { return "select" }

type Slot struct{}

func (p Slot) Key() string { return "slot" }

type Small struct{}

func (p Small) Key() string { return "small" }

type Source struct{}

func (p Source) Key() string { return "source" }

type Span struct{}

func (p Span) Key() string { return "span" }

type Strong struct{}

func (p Strong) Key() string { return "strong" }

type Style struct{}

func (p Style) Key() string { return "style" }

type Sub struct{}

func (p Sub) Key() string { return "sub" }

type Summary struct{}

func (p Summary) Key() string { return "summary" }

type Sup struct{}

func (p Sup) Key() string { return "sup" }

type Table struct{}

func (p Table) Key() string { return "table" }

type Tbody struct{}

func (p Tbody) Key() string { return "tbody" }

type Td struct{}

func (p Td) Key() string { return "td" }

type Template struct{}

func (p Template) Key() string { return "template" }

type Textarea struct{}

func (p Textarea) Key() string { return "textarea" }

type Tfoot struct{}

func (p Tfoot) Key() string { return "tfoot" }

type Th struct{}

func (p Th) Key() string { return "th" }

type Thead struct{}

func (p Thead) Key() string { return "thead" }

type Time struct{}

func (p Time) Key() string { return "time" }

type Title struct{}

func (p Title) Key() string { return "title" }

type Tr struct{}

func (p Tr) Key() string { return "tr" }

type Track struct{}

func (p Track) Key() string { return "track" }

type U struct{}

func (p U) Key() string { return "u" }

type Ul struct{}

func (p Ul) Key() string { return "ul" }

type Var struct{}

func (p Var) Key() string { return "var" }

type Video struct{}

func (p Video) Key() string { return "video" }

type Wbr struct{}

func (p Wbr) Key() string { return "wbr" }
