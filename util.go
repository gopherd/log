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

// encoder used to build json with some extra features:
//
// 1. support unqutoed key
// 2. support literal complex, e.g. 1.5+2i, 2i
// 3. support literal duration, e.g. 1s, 1ms
// 4. support literal nil
// 5. support bytes starts with 0x
type encoder struct {
	buf []byte
}

// String returns the accumulated string.
func (enc *encoder) String() string {
	return *(*string)(unsafe.Pointer(&enc.buf))
}

// Len returns the number of accumulated bytes; enc.Len() == len(enc.String()).
func (enc *encoder) Len() int { return len(enc.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (enc *encoder) Cap() int { return cap(enc.buf) }

// Write implements io.Writer Write method
func (enc *encoder) Write(p []byte) (int, error) {
	enc.buf = append(enc.buf, p...)
	return len(p), nil
}

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(b.buf).
func (enc *encoder) grow(n int) {
	buf := make([]byte, len(enc.buf), 2*cap(enc.buf)+n)
	copy(buf, enc.buf)
	enc.buf = buf
}

func (enc *encoder) reset() {
	enc.buf = enc.buf[:0]
}

func (enc *encoder) writeByte(c byte) {
	enc.buf = append(enc.buf, c)
}

func (enc *encoder) writeString(s string) {
	enc.buf = append(enc.buf, s...)
}

func (enc *encoder) encodeKey(key string) {
	if len(enc.buf) == 0 {
		enc.writeByte('{')
	} else {
		enc.writeByte(',')
	}
	if isIdent(key) {
		enc.buf = append(enc.buf, key...)
	} else {
		enc.encodeString(key)
	}
	enc.writeByte(':')
}

func (enc *encoder) finish() {
	if len(enc.buf) > 0 {
		enc.buf = append(enc.buf, '}', ' ')
	}
}

func (enc *encoder) encodeNil() {
	enc.buf = append(enc.buf, "nil"...)
}

func (enc *encoder) encodeByte(c byte) {
	enc.buf = append(enc.buf, '\'', c, '\'')
}

func (enc *encoder) encodeRune(r rune) {
	enc.buf = strconv.AppendQuoteRune(enc.buf, r)
}

func (enc *encoder) encodeString(s string) {
	enc.buf = strconv.AppendQuote(enc.buf, s)
}

func (enc *encoder) encodeInt(i int64) {
	enc.buf = strconv.AppendInt(enc.buf, i, 10)
}

func (enc *encoder) encodeUint(i uint64) {
	enc.buf = strconv.AppendUint(enc.buf, i, 10)
}

func (enc *encoder) encodeFloat(f float64, bits int) {
	enc.buf = strconv.AppendFloat(enc.buf, f, 'f', -1, bits)
}

func (enc *encoder) encodeFloat32(f float32) {
	enc.encodeFloat(float64(f), 32)
}

func (enc *encoder) encodeFloat64(f float64) {
	enc.encodeFloat(f, 64)
}

func (enc *encoder) encodeBool(v bool) {
	enc.buf = strconv.AppendBool(enc.buf, v)
}

func (enc *encoder) encodeComplex(r, i float64, bits int) {
	if r != 0 {
		enc.encodeFloat(r, bits)
	}
	if i != 0 {
		if r != 0 {
			enc.buf = append(enc.buf, '+')
		}
		enc.encodeFloat(i, bits)
		enc.buf = append(enc.buf, 'i')
	} else if r == 0 {
		enc.buf = append(enc.buf, '0')
	}
}

func (enc *encoder) encodeComplex64(c complex64) {
	r, i := real(c), imag(c)
	enc.encodeComplex(float64(r), float64(i), 32)
}

func (enc *encoder) encodeComplex128(c complex128) {
	r, i := real(c), imag(c)
	enc.encodeComplex(r, i, 64)
}

func (enc *encoder) encodeScalar(value any) bool {
	switch x := value.(type) {
	case int:
		enc.encodeInt(int64(x))
	case int8:
		enc.encodeInt(int64(x))
	case int16:
		enc.encodeInt(int64(x))
	case int32:
		enc.encodeInt(int64(x))
	case int64:
		enc.encodeInt(x)
	case uint:
		enc.encodeUint(uint64(x))
	case uint8:
		enc.encodeUint(uint64(x))
	case uint16:
		enc.encodeUint(uint64(x))
	case uint32:
		enc.encodeUint(uint64(x))
	case uint64:
		enc.encodeUint(x)
	case float32:
		enc.encodeFloat32(x)
	case float64:
		enc.encodeFloat64(x)
	case bool:
		enc.encodeBool(x)
	case complex64:
		enc.encodeComplex64(x)
	case complex128:
		enc.encodeComplex128(x)
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
