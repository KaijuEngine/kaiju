package ui

import (
	"kaiju/rendering"
)

const (
	offIdleTexture  = "textures/checkbox-off-idle.png"
	offDownTexture  = "textures/checkbox-off-down.png"
	offHoverTexture = "textures/checkbox-off-hover.png"
	onIdleTexture   = "textures/checkbox-on-idle.png"
	onDownTexture   = "textures/checkbox-on-down.png"
	onHoverTexture  = "textures/checkbox-on-hover.png"
)

const (
	texOffIdle = iota
	texOffDown
	texOffHover
	texOnIdle
	texOnDown
	texOnHover
)

const (
	defaultCheckboxSize = 25
)

type localCheckboxData struct {
	label     *Label
	textures  [6]*rendering.Texture
	isChecked bool
}

type Checkbox Panel

func (cb *Checkbox) data() *localCheckboxData {
	return cb.localData.(*localCheckboxData)
}

func (p *Panel) ConvertToCheckbox() *Checkbox {
	ld := &localCheckboxData{}
	tc := p.host.TextureCache()
	ld.textures[texOffIdle], _ = tc.Texture(
		offIdleTexture, rendering.TextureFilterLinear)
	ld.textures[texOffDown], _ = tc.Texture(
		offDownTexture, rendering.TextureFilterLinear)
	ld.textures[texOffHover], _ = tc.Texture(
		offHoverTexture, rendering.TextureFilterLinear)
	ld.textures[texOnIdle], _ = tc.Texture(
		onIdleTexture, rendering.TextureFilterLinear)
	ld.textures[texOnDown], _ = tc.Texture(
		onDownTexture, rendering.TextureFilterLinear)
	ld.textures[texOnHover], _ = tc.Texture(
		onHoverTexture, rendering.TextureFilterLinear)
	cb := (*Checkbox)(p)
	p.AddEvent(EventTypeEnter, cb.onHover)
	p.AddEvent(EventTypeExit, cb.onBlur)
	p.AddEvent(EventTypeDown, cb.onDown)
	p.AddEvent(EventTypeUp, cb.onUp)
	p.AddEvent(EventTypeClick, cb.onClick)
	p.localData = ld
	cb.layout.Scale(defaultCheckboxSize, defaultCheckboxSize)
	(*Panel)(cb).SetBackground(ld.textures[texOffIdle])
	return cb
}

func (cb *Checkbox) onHover() {
	var target *rendering.Texture = nil
	data := cb.data()
	if cb.isDown {
		if data.isChecked {
			target = data.textures[texOnDown]
		} else {
			target = data.textures[texOffDown]
		}
	} else {
		if data.isChecked {
			target = data.textures[texOnHover]
		} else {
			target = data.textures[texOffHover]
		}
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onBlur() {
	data := cb.data()
	var target *rendering.Texture = nil
	if data.isChecked {
		target = data.textures[texOnIdle]
	} else {
		target = data.textures[texOffIdle]
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onDown() {
	data := cb.data()
	var target *rendering.Texture = nil
	if data.isChecked {
		target = data.textures[texOnDown]
	} else {
		target = data.textures[texOffDown]
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onUp() {
	data := cb.data()
	var target *rendering.Texture = nil
	if data.isChecked {
		target = data.textures[texOnHover]
	} else {
		target = data.textures[texOffHover]
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb *Checkbox) onClick() {
	data := cb.data()
	cb.SetChecked(!data.isChecked)
}

func (cb *Checkbox) SetChecked(isChecked bool) {
	data := cb.data()
	data.isChecked = isChecked
	var target *rendering.Texture = nil
	if data.isChecked {
		if cb.hovering {
			target = data.textures[texOnHover]
		} else {
			target = data.textures[texOnIdle]
		}
	} else {
		if cb.hovering {
			target = data.textures[texOffHover]
		} else {
			target = data.textures[texOffIdle]
		}
	}
	(*Panel)(cb).SetBackground(target)
}

func (cb Checkbox) IsChecked() bool {
	return cb.data().isChecked
}
