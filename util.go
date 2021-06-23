package log

import (
	"strconv"
	"time"
	"unicode/utf8"
	"unsafe"
)

const hex = "0123456789abcdef"

// builder used to build string
type builder struct {
	buf []byte
}

// String returns the accumulated string.
func (b *builder) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *builder) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *builder) Cap() int { return cap(b.buf) }

func (b *builder) reset() {
	b.buf = b.buf[:0]
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (b *builder) grow(n int) {
	buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
	copy(buf, b.buf)
	b.buf = buf
}

func (b *builder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *builder) writeByte(c byte) {
	b.buf = append(b.buf, c)
}

func (b *builder) writeRune(r rune) {
	if r < utf8.RuneSelf {
		b.buf = append(b.buf, byte(r))
		return
	}
	l := len(b.buf)
	if cap(b.buf)-l < utf8.UTFMax {
		b.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(b.buf[l:l+utf8.UTFMax], r)
	b.buf = b.buf[:l+n]
}

func (b *builder) writeString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *builder) writeQuotedString(s string) {
	b.buf = strconv.AppendQuote(b.buf, s)
}

func (b *builder) writeInt(i int64) {
	b.buf = strconv.AppendInt(b.buf, i, 10)
}

func (b *builder) writeUint(i uint64) {
	b.buf = strconv.AppendUint(b.buf, i, 10)
}

func (b *builder) writeFloat32(f float32) {
	b.buf = strconv.AppendFloat(b.buf, float64(f), 'f', -1, 32)
}

func (b *builder) writeFloat64(f float64) {
	b.buf = strconv.AppendFloat(b.buf, f, 'f', -1, 64)
}

func (b *builder) writeBool(v bool) {
	b.buf = strconv.AppendBool(b.buf, v)
}

func (b *builder) writeComplex64(c complex64) {
	r, i := real(c), imag(c)
	if r != 0 {
		b.writeFloat32(r)
	}
	if i != 0 {
		if r != 0 {
			b.buf = append(b.buf, '+')
		}
		b.writeFloat32(i)
		b.buf = append(b.buf, 'i')
	} else if r == 0 {
		b.buf = append(b.buf, '0')
	}
}

func (b *builder) writeComplex128(c complex128) {
	r, i := real(c), imag(c)
	if r != 0 {
		b.writeFloat64(r)
	}
	if i != 0 {
		if r != 0 {
			b.buf = append(b.buf, '+')
		}
		b.writeFloat64(i)
		b.buf = append(b.buf, 'i')
	} else if r == 0 {
		b.buf = append(b.buf, '0')
	}
}

func (b *builder) tryWriteScalar(value interface{}) bool {
	switch x := value.(type) {
	case int:
		b.writeInt(int64(x))
	case int8:
		b.writeInt(int64(x))
	case int16:
		b.writeInt(int64(x))
	case int32:
		b.writeInt(int64(x))
	case int64:
		b.writeInt(x)
	case uint:
		b.writeUint(uint64(x))
	case uint8:
		b.writeUint(uint64(x))
	case uint16:
		b.writeUint(uint64(x))
	case uint32:
		b.writeUint(uint64(x))
	case uint64:
		b.writeUint(x)
	case float32:
		b.writeFloat32(x)
	case float64:
		b.writeFloat64(x)
	case bool:
		b.writeBool(x)
	case complex64:
		b.writeComplex64(x)
	case complex128:
		b.writeComplex128(x)
	default:
		return false
	}
	return true
}

// String returns a string representing the duration in the form "72h3m0.5s".
// Leading zero units are omitted. As a special case, durations less than one
// second format use a smaller unit (milli-, micro-, or nanoseconds) to ensure
// that the leading digit is non-zero. The zero duration formats as 0s.
func formatDuration(buf []byte, d time.Duration) int {
	// Largest time is 2540400h10m10.000000000s
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	if u < uint64(time.Second) {
		// Special case: if duration is smaller than a second,
		// use smaller units, like 1.2ms
		var prec int
		w--
		buf[w] = 's'
		w--
		switch {
		case u == 0:
			buf[0] = '0'
			buf[1] = 's'
			return 2
		case u < uint64(time.Microsecond):
			// print nanoseconds
			prec = 0
			buf[w] = 'n'
		case u < uint64(time.Millisecond):
			// print microseconds
			prec = 3
			// U+00B5 'µ' micro sign == 0xC2 0xB5
			w-- // Need room for two bytes.
			copy(buf[w:], "µ")
		default:
			// print milliseconds
			prec = 6
			buf[w] = 'm'
		}
		w, u = fmtFrac(buf[:w], u, prec)
		w = fmtInt(buf[:w], u)
	} else {
		w--
		buf[w] = 's'

		w, u = fmtFrac(buf[:w], u, 9)

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60

		// u is now integer minutes
		if u > 0 {
			w--
			buf[w] = 'm'
			w = fmtInt(buf[:w], u%60)
			u /= 60

			// u is now integer hours
			// Stop at hours because days can be different lengths.
			if u > 0 {
				w--
				buf[w] = 'h'
				w = fmtInt(buf[:w], u)
			}
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}

	copy(buf, buf[w:])
	return len(buf) - w
}

// fmtFrac formats the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf, omitting trailing zeros. It omits the decimal
// point too when the fraction is 0. It returns the index where the
// output bytes begin and the value v/10**prec.
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	print := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		print = print || digit != 0
		if print {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if print {
		w--
		buf[w] = '.'
	}
	return w, v
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

type appendFormatter interface {
	AppendFormat(buf []byte) []byte
}
