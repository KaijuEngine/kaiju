package framework

import "math/rand/v2"

type EncryptedInt32 struct {
	RawValue int32
	seed     int32
}

func (e EncryptedInt32) Value() int32 {
	return e.RawValue ^ e.seed
}

func (e *EncryptedInt32) SetValue(value int32) {
	e.seed = rand.Int32()
	e.RawValue = value ^ e.seed
}

func (e *EncryptedInt32) Increment() {
	e.SetValue(e.Value() + 1)
}

func (e *EncryptedInt32) Decrement() {
	e.SetValue(e.Value() - 1)
}

func (e *EncryptedInt32) Add(amount int32) {
	e.SetValue(e.Value() + amount)
}

func (e *EncryptedInt32) Subtract(amount int32) {
	e.SetValue(e.Value() - amount)
}

func (e *EncryptedInt32) Decrypt() {
	e.RawValue ^= e.seed
	e.seed = 0
}

func (e *EncryptedInt32) Encrypt() {
	e.SetValue(e.RawValue)
}
