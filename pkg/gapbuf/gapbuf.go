package gapbuf

const (
	bufferMin = 64*1024 - 1
)

type gapBuffer struct {
	gapStart int
	gapEnd   int
	buf      []rune
}

func (b *gapBuffer) gapSize() int {
	return b.gapEnd - b.gapStart
}

func (b *gapBuffer) moveGap(position int) {
	if position == b.gapStart {
		return
	} else if position < b.gapStart {
		d := b.gapStart - position
		b.gapStart -= d
		b.gapEnd -= d
		copy(b.buf[b.gapEnd:], b.buf[b.gapStart:b.gapStart+d])
	} else {
		d := position - b.gapStart
		b.gapStart += d
		b.gapEnd += d
		copy(b.buf[b.gapStart-d:b.gapStart], b.buf[b.gapEnd-d:b.gapEnd])
	}
}

func (b *gapBuffer) ensureGap(length int) {
	if b.gapSize() >= length {
		return
	}

	newSize := len(b.buf) + length + bufferMin
	endSize := len(b.buf) - b.gapEnd

	buf := make([]rune, newSize)
	copy(buf, b.buf[:b.gapStart])
	copy(buf[newSize-endSize:], b.buf[b.gapEnd:])

	b.buf = buf
	b.gapEnd = newSize - endSize
}

func (b *gapBuffer) deleteBuffer(position, length int) {
	b.moveGap(position)
	b.gapEnd += length
}

func (b *gapBuffer) readBuffer(c []rune, position int) (n int) {
	if position < b.gapStart {
		d := b.gapStart - position
		n += copy(c[:d], b.buf[position:b.gapStart])
		n += copy(c[d:], b.buf[b.gapEnd:])

		return n
	}

	position += b.gapSize()
	return copy(c, b.buf[position:])
}

func (b *gapBuffer) insertBuffer(position int, s []rune) {
	b.ensureGap(len(s))
	b.moveGap(position)
	copy(b.buf[b.gapStart:], s)
	b.gapStart += len(s)
}

type Buffer struct {
	numRunes int
	gb       *gapBuffer
}

func New() Buffer {
	return Buffer{
		gb: &gapBuffer{
			buf:    make([]rune, bufferMin),
			gapEnd: bufferMin,
		},
	}
}

func (b *Buffer) Read(p []rune, length, position int) (n int) {
	if position+length > b.numRunes {
		length = b.numRunes - position
	}
	if length <= 0 {
		return 0
	}

	return b.gb.readBuffer(p, position)
}

func (b *Buffer) Insert(runes []rune, position int) {
	if len(runes) > 0 {
		b.gb.insertBuffer(position, runes)
		b.numRunes += len(runes)
	}
}

func (b *Buffer) Delete(startPosition, endPosition int) {
	length := endPosition - startPosition
	b.gb.deleteBuffer(startPosition, length)
	b.numRunes -= length
}
