package console

import (
	"kaiju/engine"
	"kaiju/hid"
	"kaiju/ui"
	"kaiju/uimarkup"
	"kaiju/uimarkup/markup"
	"strings"
)

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
	}
	return h.data[h.idx]
}

type Console struct {
	doc        *markup.Document
	host       *engine.Host
	commands   map[string]func(string) string
	history    history
	historyIdx int
	updateId   int
	isActive   bool
	input      *ui.Input
}

func For(host *engine.Host) *Console {
	c, ok := consoles[host]
	if !ok {
		c = initialize(host)
		consoles[host] = c
	}
	return c
}

func UnlinkHost(host *engine.Host) { delete(consoles, host) }

func initialize(host *engine.Host) *Console {
	console := &Console{
		host:     host,
		commands: map[string]func(string) string{},
		history:  newHistory(),
	}
	consoleHTML, _ := host.AssetDatabase().ReadText("ui/console.html")
	console.doc = uimarkup.DocumentFromHTMLString(host,
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
	console.AddCommand("help", console.help)
	return console
}

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

func (c *Console) AddCommand(key string, fn func(cmd string) string) {
	c.commands[key] = fn
}

func (c *Console) help(arg string) string {
	sb := strings.Builder{}
	sb.WriteString("Available Commands:\n")
	for name := range c.commands {
		sb.WriteString(name)
		sb.WriteRune('\n')
	}

	return sb.String()
}

func (c *Console) submit(input *ui.Input) {
	cmd := strings.TrimSpace(input.Text())
	if cmd == "" {
		return
	}
	input.SetText("")
	c.history.add(cmd)
	head := strings.Index(cmd, " ")
	var key, value string
	if head != -1 {
		key = cmd[:head]
		value = strings.TrimSpace(cmd[head+1:])
	} else {
		key = cmd
	}
	var res string
	if fn, ok := c.commands[key]; ok {
		res = strings.TrimSpace(fn(value))
	}
	cc, _ := c.doc.GetElementById("consoleContent")
	lbl := ui.FirstOnEntity(cc.HTML.Children[0].DocumentElement.UI.Entity()).(*ui.Label)
	if res != "" {
		lbl.SetText(lbl.Text() + "\n" + cmd + "\n" + res)
	} else {
		lbl.SetText(lbl.Text() + "\n" + cmd)
	}
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
