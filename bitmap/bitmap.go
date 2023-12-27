package bitmap

const bitsInByte = 8

type Bitmap []byte

func New(length int) Bitmap {
	return make([]byte, LengthFor(length))
}

func LengthFor(byteCount int) int {
	return (byteCount / bitsInByte) + 1
}

func (b Bitmap) Check(index int) bool {
	return b[index/bitsInByte]&0x01<<(index%bitsInByte) != 0
}

func (b Bitmap) Set(index int) {
	b[index/bitsInByte] |= 0x01 << (index % bitsInByte)
}

func (b Bitmap) Assign(index int, value bool) {
	if value {
		b.Set(index)
	} else {
		b.Reset(index)
	}
}

func (b Bitmap) Reset(index int) {
	b[index/bitsInByte] &= ^(0x01 << (index % bitsInByte))
}

func (b Bitmap) Toggle(index int) {
	b[index/bitsInByte] ^= 0x01 << (index % bitsInByte)
}

func (b Bitmap) Count() int {
	count := 0
	length := len(b) * bitsInByte
	for i := 0; i < length; i++ {
		if b.Check(i) {
			count++
		}
	}
	return count
}

func (b Bitmap) CountInverse() int {
	return len(b)*bitsInByte - b.Count()
}

func (b Bitmap) Clear() {
	clear(b)
}
