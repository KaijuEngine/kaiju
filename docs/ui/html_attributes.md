---
title: Kaiju Engine | UI HTML Attributes
---

# UI HTML Attributes
Some standard as well as non-standard HTML attributes are available for developers to use. Below are some tables going over different attributes that are currently available and what they are used for.

## General
| Attribute    | Description                                                        |
| ------------ | ------------------------------------------------------------------ |
| id           | A unique id for this element, useful for searching in Go           |
| group        | The group that this element belongs to, useful for searching in Go |
| class        | The CSS class style to use                                         |
| style        | An inline override style to use for this element                   |

## Events
We have both standard and non-standard events built into the UI elements. For the most part, all automatic events will be prefixed with `on` for the HTML attribute name.

| Attribute    | Description                                                             |
| ------------ | ----------------------------------------------------------------------- |
| onclick      | The element was clicked or tapped via finger or stylus                  |
| onrightclick | The element was right-clicked                                           |
| onmiss       | The anything but the element was clicked or tapped via finger or stylus |
| onsubmit     | The enter key is pressed in an input field                              |
| onkeydown    | A key is pressed while the element is focused                           |
| onkeyup      | A key is released while the element is focused                          |
| ondblclick   | The element was double clicked or tapped via finger or stylus           |
| onmouseover  | The mouse or stylus hovers over this element                            |
| onmouseenter | The mouse or stylus hovers over this element                            |
| onmouseleave | The mouse or stylus exits hovering over this element                    |
| onmouseexit  | The mouse or stylus exits hovering over this element                    |
| onmousedown  | The mouse button, touch, or stylus is pressed on the element            |
| onmouseup    | The mouse button, touch, or stylus is released on the element           |
| onmousewheel | The mouse wheel was moved while hovering over the element               |
| onchange     | The value of the element changed (input, select, checkbox, etc.)        |
| ondragenter  | The cursor is currently dragging something and hovers over this         |
| ondragleave  | The cursor is currently dragging something and stops hovering over this |
| ondragstart  | The cursor started dragging this element                                |
| ondrop       | The cursor was dragging something and dropped it onto this element      |
| ondragend    | The cursor stopped dragging this element                                |