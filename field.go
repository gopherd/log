package log

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unicode"
)

// Fields holds context fields
type Fields struct {
	level   Level
	prefix  string
	builder builder
}

var fieldsPool = sync.Pool{
	New: func() interface{} {
		return new(Fields)
	},
}

func getFields(level Level, prefix Prefix) *Fields {
	if gprinter.GetLevel() < level {
		return nil
	}
	fields := fieldsPool.Get().(*Fields)
	fields.reset(level, string(prefix))
	return fields
}

func putFields(fields *Fields) {
	if fields.builder.Cap() < 1024 {
		fieldsPool.Put(fields)
	}
}

func (fields *Fields) reset(level Level, prefix string) {
	fields.level = level
	fields.prefix = prefix
	fields.builder.reset()
}

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

func (fields *Fields) writeKey(key string) {
	if fields.builder.Len() == 0 {
		fields.builder.writeByte('{')
	} else {
		fields.builder.writeByte(' ')
	}
	if isIdent(key) {
		fields.builder.writeString(key)
	} else {
		fields.builder.writeQuotedString(key)
	}
	fields.builder.writeByte(':')
}

// Print prints logging with context fields. After this call,
// the fields not available.
func (fields *Fields) Print(s string) {
	if fields == nil {
		return
	}
	if fields.builder.Len() > 0 {
		fields.builder.writeString("} ")
	}
	fields.builder.writeString(s)
	gprinter.Printf(1, fields.level, fields.prefix, fields.builder.String())
	putFields(fields)
}

func (fields *Fields) Int(key string, value int) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int8(key string, value int8) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int16(key string, value int16) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int32(key string, value int32) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(int64(value))
	}
	return fields
}

func (fields *Fields) Int64(key string, value int64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeInt(value)
	}
	return fields
}

func (fields *Fields) Uint(key string, value uint) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint8(key string, value uint8) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint16(key string, value uint16) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint32(key string, value uint32) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(uint64(value))
	}
	return fields
}

func (fields *Fields) Uint64(key string, value uint64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeUint(value)
	}
	return fields
}

func (fields *Fields) Float32(key string, value float32) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeFloat32(value)
	}
	return fields
}

func (fields *Fields) Float64(key string, value float64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeFloat64(value)
	}
	return fields
}

func (fields *Fields) Complex64(key string, value complex64) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeComplex64(value)
	}
	return fields
}

func (fields *Fields) Complex128(key string, value complex128) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeComplex128(value)
	}
	return fields
}

func (fields *Fields) Byte(key string, value byte) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeByte('\'')
		fields.builder.writeByte(value)
		fields.builder.writeByte('\'')
	}
	return fields
}

func (fields *Fields) Rune(key string, value rune) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.buf = strconv.AppendQuoteRune(fields.builder.buf, value)
	}
	return fields
}

func (fields *Fields) Bool(key string, value bool) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeBool(value)
	}
	return fields
}

func (fields *Fields) String(key string, value string) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeQuotedString(value)
	}
	return fields
}

func (fields *Fields) Error(key string, value error) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("nil")
		} else {
			fields.builder.buf = strconv.AppendQuote(fields.builder.buf, value.Error())
		}
	}
	return fields
}

func (fields *Fields) Any(key string, value interface{}) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString("nil")
		} else {
			switch x := value.(type) {
			case error:
				fields.builder.writeQuotedString(x.Error())
			case fmt.Stringer:
				fields.builder.writeQuotedString(x.String())
			case string:
				fields.builder.writeQuotedString(x)
			case appendFormatter:
				fields.builder.buf = x.AppendFormat(fields.builder.buf)
			default:
				if !fields.builder.tryWriteScalar(value) {
					fields.builder.writeQuotedString(fmt.Sprintf("%v", value))
				}
			}
		}
	}
	return fields
}

func (fields *Fields) Type(key string, value interface{}) *Fields {
	if fields != nil {
		fields.writeKey(key)
		if value == nil {
			fields.builder.writeString(`"nil"`)
		} else {
			fields.builder.writeQuotedString(reflect.TypeOf(value).String())
		}
	}
	return fields
}

func (fields *Fields) Exec(key string, stringer func() string) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.writeQuotedString(stringer())
	}
	return fields
}

func (fields *Fields) writeTime(key string, value time.Time, layout string) *Fields {
	if fields != nil {
		fields.writeKey(key)
		fields.builder.buf = value.AppendFormat(fields.builder.buf, layout)
	}
	return fields
}

func (fields *Fields) Date(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, "2006-01-02Z07:00")
}

func (fields *Fields) Time(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, time.RFC3339Nano)
}

func (fields *Fields) Seconds(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, time.RFC3339)
}

func (fields *Fields) Milliseconds(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, "2006-01-02T15:04:05.999Z07:00")
}

func (fields *Fields) Microseconds(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, "2006-01-02T15:04:05.999999Z07:00")
}

func (fields *Fields) Duration(key string, value time.Duration) *Fields {
	if fields != nil {
		fields.writeKey(key)
		const reserved = 32
		l := len(fields.builder.buf)
		if cap(fields.builder.buf)-l < reserved {
			fields.builder.grow(reserved)
		}
		n := formatDuration(fields.builder.buf[l:l+reserved], value)
		fields.builder.buf = fields.builder.buf[:l+n]
	}
	return fields
}
