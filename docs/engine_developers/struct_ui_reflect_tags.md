---
title: Struct UI reflect tags | Kaiju Engine
---

# Struct UI reflect tags
We have a few different tags that help with reflecting the structure fields in
the editor.

|   Key   |      Arguments       | Description |
| ------- | -------------------- | ----------- |
| clamp   | number,number,number | Clamps the value between 2 numbers: default, min, max |
| default | any                  | Sets the default/starting value |
| visible | "false"              | Makes the field invisible in the UI, omit if true |
| label   | string               | Explicitly set what the labe is in the UI |
