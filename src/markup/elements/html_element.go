/*****************************************************************************/
/* html_element.go                                                           */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package elements

type Element interface {
	Key() string
}

var ElementMap = map[string]Element{
	"a":          A{},
	"abbr":       Abbr{},
	"address":    Address{},
	"area":       Area{},
	"article":    Article{},
	"aside":      Aside{},
	"audio":      Audio{},
	"b":          B{},
	"base":       Base{},
	"bdi":        Bdi{},
	"bdo":        Bdo{},
	"blockquote": Blockquote{},
	"body":       Body{},
	"br":         Br{},
	"button":     Button{},
	"canvas":     Canvas{},
	"caption":    Caption{},
	"cite":       Cite{},
	"code":       Code{},
	"col":        Col{},
	"colgroup":   Colgroup{},
	"data":       Data{},
	"datalist":   Datalist{},
	"dd":         Dd{},
	"del":        Del{},
	"details":    Details{},
	"dfn":        Dfn{},
	"dialog":     Dialog{},
	"div":        Div{},
	"dl":         Dl{},
	"dt":         Dt{},
	"em":         Em{},
	"embed":      Embed{},
	"fieldset":   Fieldset{},
	"figcaption": Figcaption{},
	"figure":     Figure{},
	"footer":     Footer{},
	"form":       Form{},
	"h1":         H1{},
	"h2":         H2{},
	"h3":         H3{},
	"h4":         H4{},
	"h5":         H5{},
	"h6":         H6{},
	"head":       Head{},
	"header":     Header{},
	"hgroup":     Hgroup{},
	"hr":         Hr{},
	"html":       Html{},
	"i":          I{},
	"iframe":     Iframe{},
	"img":        Img{},
	"input":      Input{},
	"ins":        Ins{},
	"kbd":        Kbd{},
	"label":      Label{},
	"legend":     Legend{},
	"li":         Li{},
	"link":       Link{},
	"main":       Main{},
	"map":        Map{},
	"mark":       Mark{},
	"menu":       Menu{},
	"meta":       Meta{},
	"meter":      Meter{},
	"nav":        Nav{},
	"noscript":   Noscript{},
	"object":     Object{},
	"ol":         Ol{},
	"optgroup":   Optgroup{},
	"option":     Option{},
	"output":     Output{},
	"p":          P{},
	"picture":    Picture{},
	"pre":        Pre{},
	"progress":   Progress{},
	"q":          Q{},
	"rp":         Rp{},
	"rt":         Rt{},
	"ruby":       Ruby{},
	"s":          S{},
	"samp":       Samp{},
	"script":     Script{},
	"search":     Search{},
	"section":    Section{},
	"select":     Select{},
	"slot":       Slot{},
	"small":      Small{},
	"source":     Source{},
	"span":       Span{},
	"strong":     Strong{},
	"style":      Style{},
	"sub":        Sub{},
	"summary":    Summary{},
	"sup":        Sup{},
	"table":      Table{},
	"tbody":      Tbody{},
	"td":         Td{},
	"template":   Template{},
	"textarea":   Textarea{},
	"tfoot":      Tfoot{},
	"th":         Th{},
	"thead":      Thead{},
	"time":       Time{},
	"title":      Title{},
	"tr":         Tr{},
	"track":      Track{},
	"u":          U{},
	"ul":         Ul{},
	"var":        Var{},
	"video":      Video{},
	"wbr":        Wbr{},
}
