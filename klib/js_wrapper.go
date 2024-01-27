//go:build js

package klib

import "syscall/js"

func JSUint8ArrayToBytes(jsArray js.Value) []byte {
	jsBin := js.Global().Get("Uint8Array").New(jsArray)
	data := make([]byte, jsBin.Get("length").Int())
	js.CopyBytesToGo(data, jsBin)
	return data
}

func JSBytesToUint8Array(data []byte) js.Value {
	jsArray := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(jsArray, data)
	return jsArray
}
