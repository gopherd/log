package log

import (
	"strconv"
	"time"
	"unicode"
	"unsafe"
)

const hex = "0123456789abcdef"

func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if !isIdentRune(c, i) {
			return false
		}
	}
	return true
}

func isIdentRune(ch rune, i int) bool {
	return ch == '_' || ch == '-' || ch == '.' || ch == '#' || ch == '$' || ch == '/' ||
		unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

// jsonx used to build json with some extra features:
//
// 1. support unqutoed key
// 2. support literal complex, e.g. 1.5+2i, 2i
// 3. support literal duration, e.g. 1s, 1ms
// 4. support literal nil
// 5. support bytes starts with 0x
type jsonx struct {
	buf []byte
}

// String returns the accumulated string.
func (b *jsonx) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (b *jsonx) Len() int { return len(b.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (b *jsonx) Cap() int { return cap(b.buf) }

// Write implements io.Writer Write method
func (b *jsonx) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *jsonx) reset() {
	b.buf = b.buf[:0]
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (b *jsonx) grow(n int) {
	buf := make([]byte, len(b.buf), 2*cap(b.buf)+n)
	copy(buf, b.buf)
	b.buf = buf
}

func (b *jsonx) writeByte(c byte) {
	b.buf = append(b.buf, c)
}

func (b *jsonx) writeString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *jsonx) encodeKey(key string) {
	if len(b.buf) == 0 {
		b.writeByte('{')
	} else {
		b.writeByte(',')
	}
	if isIdent(key) {
		b.buf = append(b.buf, key...)
	} else {
		b.encodeString(key)
	}
	b.writeByte(':')
}

func (b *jsonx) finish() {
	if len(b.buf) > 0 {
		b.buf = append(b.buf, '}', ' ')
	}
}

func (b *jsonx) encodeNil() {
	b.buf = append(b.buf, "nil"...)
}

func (b *jsonx) encodeByte(c byte) {
	b.buf = append(b.buf, '\'', c, '\'')
}

func (b *jsonx) encodeRune(r rune) {
	b.buf = strconv.AppendQuoteRune(b.buf, r)
}

func (b *jsonx) encodeString(s string) {
	b.buf = strconv.AppendQuote(b.buf, s)
}

func (b *jsonx) encodeInt(i int64) {
	b.buf = strconv.AppendInt(b.buf, i, 10)
}

func (b *jsonx) encodeUint(i uint64) {
	b.buf = strconv.AppendUint(b.buf, i, 10)
}

func (b *jsonx) encodeFloat(f float64, bits int) {
	b.buf = strconv.AppendFloat(b.buf, f, 'f', -1, bits)
}

func (b *jsonx) encodeFloat32(f float32) {
	b.encodeFloat(float64(f), 32)
}

func (b *jsonx) encodeFloat64(f float64) {
	b.encodeFloat(f, 64)
}

func (b *jsonx) encodeBool(v bool) {
	b.buf = strconv.AppendBool(b.buf, v)
}

func (b *jsonx) encodeComplex(r, i float64, bits int) {
	if r != 0 {
		b.encodeFloat(r, bits)
	}
	if i != 0 {
		if r != 0 {
			b.buf = append(b.buf, '+')
		}
		b.encodeFloat(i, bits)
		b.buf = append(b.buf, 'i')
	} else if r == 0 {
		b.buf = append(b.buf, '0')
	}
}

func (b *jsonx) encodeComplex64(c complex64) {
	r, i := real(c), imag(c)
	b.encodeComplex(float64(r), float64(i), 32)
}

func (b *jsonx) encodeComplex128(c complex128) {
	r, i := real(c), imag(c)
	b.encodeComplex(r, i, 64)
}

func (b *jsonx) encodeScalar(value interface{}) bool {
	switch x := value.(type) {
	case int:
		b.encodeInt(int64(x))
	case int8:
		b.encodeInt(int64(x))
	case int16:
		b.encodeInt(int64(x))
	case int32:
		b.encodeInt(int64(x))
	case int64:
		b.encodeInt(x)
	case uint:
		b.encodeUint(uint64(x))
	case uint8:
		b.encodeUint(uint64(x))
	case uint16:
		b.encodeUint(uint64(x))
	case uint32:
		b.encodeUint(uint64(x))
	case uint64:
		b.encodeUint(x)
	case float32:
		b.encodeFloat32(x)
	case float64:
		b.encodeFloat64(x)
	case bool:
		b.encodeBool(x)
	case complex64:
		b.encodeComplex64(x)
	case complex128:
		b.encodeComplex128(x)
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
