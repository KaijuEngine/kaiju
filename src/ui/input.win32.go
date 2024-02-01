//go:build windows

package ui

func (input *Input) internalCopyToClipboard() {
	input.host.Window.CopyToClipboard(input.Text())
}

func (input *Input) internalCutToClipboard() {
	input.host.Window.CopyToClipboard(input.Text())
	input.setText("")
}

func (input *Input) internalPasteFromClipboard() {
	text := input.host.Window.ClipboardContents()
	input.SetText(text)
}
