package log

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

// Fields holds context fields
type Fields struct {
	level   Level
	prefix  string
	encoder jsonx
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
	if fields.encoder.Cap() < 1024 {
		fieldsPool.Put(fields)
	}
}

func (fields *Fields) reset(level Level, prefix string) {
	fields.level = level
	fields.prefix = prefix
	fields.encoder.reset()
}

// Print prints logging with context fields. After this call,
// the fields not available.
func (fields *Fields) Print(s string) {
	if fields == nil {
		return
	}
	fields.encoder.finish()
	fields.encoder.writeString(s)
	gprinter.Printf(1, fields.level, fields.prefix, fields.encoder.String())
	putFields(fields)
}

//loglint:method Int
func (fields *Fields) Int(key string, value int) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt(int64(value))
	}
	return fields
}

//loglint:method Int8
func (fields *Fields) Int8(key string, value int8) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt(int64(value))
	}
	return fields
}

//loglint:method Int16
func (fields *Fields) Int16(key string, value int16) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt(int64(value))
	}
	return fields
}

//loglint:method Int32
func (fields *Fields) Int32(key string, value int32) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt(int64(value))
	}
	return fields
}

//loglint:method Int64
func (fields *Fields) Int64(key string, value int64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeInt(value)
	}
	return fields
}

//loglint:method Uint
func (fields *Fields) Uint(key string, value uint) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint(uint64(value))
	}
	return fields
}

//loglint:method Uint8
func (fields *Fields) Uint8(key string, value uint8) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint(uint64(value))
	}
	return fields
}

//loglint:method Uint16
func (fields *Fields) Uint16(key string, value uint16) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint(uint64(value))
	}
	return fields
}

//loglint:method Uint32
func (fields *Fields) Uint32(key string, value uint32) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint(uint64(value))
	}
	return fields
}

//loglint:method Uint64
func (fields *Fields) Uint64(key string, value uint64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeUint(value)
	}
	return fields
}

//loglint:method Float32
func (fields *Fields) Float32(key string, value float32) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeFloat32(value)
	}
	return fields
}

//loglint:method Float64
func (fields *Fields) Float64(key string, value float64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeFloat64(value)
	}
	return fields
}

//loglint:method Complex64
func (fields *Fields) Complex64(key string, value complex64) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeComplex64(value)
	}
	return fields
}

//loglint:method Complex128
func (fields *Fields) Complex128(key string, value complex128) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeComplex128(value)
	}
	return fields
}

//loglint:method Byte
func (fields *Fields) Byte(key string, value byte) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeByte(value)
	}
	return fields
}

//loglint:method Rune
func (fields *Fields) Rune(key string, value rune) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeRune(value)
	}
	return fields
}

//loglint:method Bool
func (fields *Fields) Bool(key string, value bool) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeBool(value)
	}
	return fields
}

//loglint:method String
func (fields *Fields) String(key string, value string) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeString(value)
	}
	return fields
}

//loglint:method Error
func (fields *Fields) Error(key string, value error) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		if value == nil {
			fields.encoder.encodeNil()
		} else {
			fields.encoder.buf = strconv.AppendQuote(fields.encoder.buf, value.Error())
		}
	}
	return fields
}

//loglint:method Any
func (fields *Fields) Any(key string, value interface{}) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		if value == nil {
			fields.encoder.encodeNil()
		} else {
			switch x := value.(type) {
			case error:
				fields.encoder.encodeString(x.Error())
			case fmt.Stringer:
				fields.encoder.encodeString(x.String())
			case string:
				fields.encoder.encodeString(x)
			case appendFormatter:
				fields.encoder.buf = x.AppendFormat(fields.encoder.buf)
			default:
				if !fields.encoder.encodeScalar(value) {
					fields.encoder.encodeString(fmt.Sprintf("%v", value))
				}
			}
		}
	}
	return fields
}

//loglint:method Type
func (fields *Fields) Type(key string, value interface{}) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		if value == nil {
			fields.encoder.encodeString("nil")
		} else {
			fields.encoder.encodeString(reflect.TypeOf(value).String())
		}
	}
	return fields
}

//loglint:method Exec
func (fields *Fields) Exec(key string, stringer func() string) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.encodeString(stringer())
	}
	return fields
}

func (fields *Fields) writeTime(key string, value time.Time, layout string) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		fields.encoder.buf = append(fields.encoder.buf, '"')
		fields.encoder.buf = value.AppendFormat(fields.encoder.buf, layout)
		fields.encoder.buf = append(fields.encoder.buf, '"')
	}
	return fields
}

//loglint:method Date
func (fields *Fields) Date(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, "2006-01-02Z07:00")
}

//loglint:method Time
func (fields *Fields) Time(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, time.RFC3339Nano)
}

//loglint:method Seconds
func (fields *Fields) Seconds(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, time.RFC3339)
}

//loglint:method Milliseconds
func (fields *Fields) Milliseconds(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, "2006-01-02T15:04:05.999Z07:00")
}

//loglint:method Microseconds
func (fields *Fields) Microseconds(key string, value time.Time) *Fields {
	return fields.writeTime(key, value, "2006-01-02T15:04:05.999999Z07:00")
}

//loglint:method Duration
func (fields *Fields) Duration(key string, value time.Duration) *Fields {
	if fields != nil {
		fields.encoder.encodeKey(key)
		const reserved = 32
		l := len(fields.encoder.buf)
		if cap(fields.encoder.buf)-l < reserved {
			fields.encoder.grow(reserved)
		}
		n := formatDuration(fields.encoder.buf[l:l+reserved], value)
		fields.encoder.buf = fields.encoder.buf[:l+n]
	}
	return fields
}
