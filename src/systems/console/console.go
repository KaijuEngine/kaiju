/******************************************************************************/
/* console.go                                                                 */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package console

import (
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
	"strings"
)

type ConsoleFunc func(*engine.Host, string) string

var consoles = map[*engine.Host]*Console{}

type history struct {
	data []string
	idx  int
}

func newHistory() history { return history{data: []string{}} }

func (h *history) add(cmd string) {
	if h.idx > 0 && len(h.data) > 0 && h.data[h.idx-1] == cmd {
		return
	}
	h.data = h.data[:h.idx]
	h.data = append(h.data, cmd)
	h.idx++
}

func (h *history) back() string {
	if len(h.data) == 0 {
		return ""
	}
	if h.idx > 0 {
		h.idx--
	}
	return h.data[h.idx]
}

func (h *history) forward() string {
	if len(h.data) == 0 {
		return ""
	}
	if h.idx < len(h.data) {
		h.idx++
	} else {
		h.idx = len(h.data) - 1
	}
	return h.data[h.idx]
}

type ConsoleData interface{}

type consoleCommand struct {
	description string
	fn          ConsoleFunc
}

type Console struct {
	doc        *document.Document
	host       *engine.Host
	commands   map[string]consoleCommand
	history    history
	historyIdx int
	updateId   int
	isActive   bool
	input      *ui.Input
	data       map[string]ConsoleData
}

func For(host *engine.Host) *Console {
	c, ok := consoles[host]
	if !ok {
		host.CreatingEditorEntities()
		c = initialize(host)
		host.DoneCreatingEditorEntities()
		consoles[host] = c
	}
	return c
}

func initialize(host *engine.Host) *Console {
	console := &Console{
		host:     host,
		commands: map[string]consoleCommand{},
		history:  newHistory(),
		data:     make(map[string]ConsoleData),
	}
	consoleHTML, _ := host.AssetDatabase().ReadText("ui/console.html")
	console.doc = markup.DocumentFromHTMLString(host,
		string(consoleHTML), "", nil, nil)
	console.updateId = host.Updater.AddUpdate(console.update)
	console.doc.Elements[0].UI.Entity().OnDestroy.Add(func() {
		host.Updater.RemoveUpdate(console.updateId)
	})
	inputElm, _ := console.doc.GetElementById("consoleInput")
	input := inputElm.UI.(*ui.Input)
	input.Data().OnSubmit.Add(func() { console.submit(input) })
	console.input = input
	input.Clean()
	console.hide()
	console.AddCommand("help", "Display list of commands and their descriptions", console.help)
	console.AddCommand("clear", "Clears the console text", console.clear)
	return console
}

func UnlinkHost(host *engine.Host) { delete(consoles, host) }

func (c *Console) Host() *engine.Host                   { return c.host }
func (c *Console) SetData(key string, data ConsoleData) { c.data[key] = data }
func (c *Console) HasData(key string) bool              { _, ok := c.data[key]; return ok }
func (c *Console) Data(key string) ConsoleData          { return c.data[key] }
func (c *Console) DeleteData(key string)                { delete(c.data, key) }

func (c *Console) toggle() {
	if c.isActive {
		c.hide()
	} else {
		c.show()
	}
}

func (c *Console) show() {
	for i := range c.doc.Elements {
		c.doc.Elements[i].UI.Entity().Activate()
	}
	c.isActive = true
	c.input.Select()
}

func (c *Console) hide() {
	c.input.SetText(strings.TrimSuffix(c.input.Text(), "`"))
	for i := range c.doc.Elements {
		c.doc.Elements[i].UI.Entity().Deactivate()
	}
	c.isActive = false
}

func (c *Console) IsActive() bool {
	return c.isActive
}

func (c *Console) AddCommand(key, description string, fn ConsoleFunc) {
	c.commands[strings.ToLower(key)] = consoleCommand{description, fn}
}

func (c *Console) Write(message string) {
	lbl := c.outputLabel()
	lbl.SetText(lbl.Text() + "\n" + message)
}

func (c *Console) help(*engine.Host, string) string {
	sb := strings.Builder{}
	sb.WriteString("Available Commands:\n")
	for name, cmd := range c.commands {
		sb.WriteString(name)
		sb.WriteString(":\t")
		sb.WriteString(cmd.description)
		sb.WriteRune('\n')
	}

	return sb.String()
}

func (c *Console) clear(*engine.Host, string) string {
	c.outputLabel().SetText("")
	return ""
}

func (c *Console) outputLabel() *ui.Label {
	cc, _ := c.doc.GetElementById("consoleContent")
	return ui.FirstOnEntity(cc.HTML.Children[0].DocumentElement.UI.Entity()).(*ui.Label)
}

func (c *Console) submit(input *ui.Input) {
	cmdStr := strings.TrimSpace(input.Text())
	if cmdStr == "" {
		return
	}
	input.SetText("")
	c.history.add(cmdStr)
	head := strings.Index(cmdStr, " ")
	var key, value string
	if head != -1 {
		key = cmdStr[:head]
		value = strings.TrimSpace(cmdStr[head+1:])
	} else {
		key = cmdStr
	}
	var res string
	if cmd, ok := c.commands[key]; ok {
		res = strings.TrimSpace(cmd.fn(c.host, value))
	}
	lblParent, _ := c.doc.GetElementById("consoleContent")
	lbl := c.outputLabel()
	if res != "" {
		lbl.SetText(lbl.Text() + "\n" + cmdStr + "\n" + res)
	} else {
		lbl.SetText(lbl.Text() + "\n" + cmdStr)
	}
	lblParent.UIPanel.SetScrollY(matrix.FloatMax)
}

func (c *Console) update(deltaTime float64) {
	kb := &c.host.Window.Keyboard
	if kb.KeyDown(hid.KeyboardKeyBackQuote) {
		c.toggle()
	} else if kb.KeyDown(hid.KeyboardKeyUp) {
		c.input.SetText(c.history.back())
	} else if kb.KeyDown(hid.KeyboardKeyDown) {
		c.input.SetText(c.history.forward())
	}
}
