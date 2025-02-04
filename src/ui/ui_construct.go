package ui

import "kaiju/rendering"

type ConstructButton struct {
	label *Label
}

type ConstructCheckbox struct {
	checked bool
}

type ConstructInput struct {
	placeholder string
}

type ConstructLabel struct {
	text string
}

type ConstructPanel struct {
	texture          *rendering.Texture
	shaderDefinition string
}

type ConstructSelect struct {
	label   string
	options []string
}

type ConstructSlider struct {
	value float32
}

type ConstructSprite struct {
	texture         *rendering.Texture
	framesPerSecond float32
	flipTextures    []*rendering.Texture
}
