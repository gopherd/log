package log

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Recorder holds context recorder
type Recorder struct {
	level   Level
	prefix  string
	encoder encoder
}

var recorderPool = sync.Pool{
	New: func() interface{} {
		return new(Recorder)
	},
}

func getRecorder(level Level, prefix Prefix) *Recorder {
	if gprinter.GetLevel() < level {
		return nil
	}
	recorder := recorderPool.Get().(*Recorder)
	recorder.reset(level, string(prefix))
	return recorder
}

func putRecorder(recorder *Recorder) {
	if recorder.encoder.Cap() < 1024 {
		recorderPool.Put(recorder)
	}
}

func (recorder *Recorder) reset(level Level, prefix string) {
	recorder.level = level
	recorder.prefix = prefix
	recorder.encoder.reset()
}

// Print prints logging with context recorder. After this call,
// the recorder not available.
func (recorder *Recorder) Print(s string) {
	if recorder == nil {
		return
	}
	recorder.encoder.finish()
	recorder.encoder.writeString(s)
	var (
		file string
		line int
	)
	if gprinter.GetFlags()&(Lshortfile|Llongfile) != 0 {
		_, file, line, _ = runtime.Caller(1)
	}
	gprinter.Printf(file, line, recorder.level, recorder.prefix, recorder.encoder.String())
	putRecorder(recorder)
}

func (recorder *Recorder) Int(key string, value int) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt(int64(value))
	}
	return recorder
}

func (recorder *Recorder) Int8(key string, value int8) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt(int64(value))
	}
	return recorder
}

func (recorder *Recorder) Int16(key string, value int16) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt(int64(value))
	}
	return recorder
}

func (recorder *Recorder) Int32(key string, value int32) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt(int64(value))
	}
	return recorder
}

func (recorder *Recorder) Int64(key string, value int64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeInt(value)
	}
	return recorder
}

func (recorder *Recorder) Uint(key string, value uint) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint(uint64(value))
	}
	return recorder
}

func (recorder *Recorder) Uint8(key string, value uint8) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint(uint64(value))
	}
	return recorder
}

func (recorder *Recorder) Uint16(key string, value uint16) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint(uint64(value))
	}
	return recorder
}

func (recorder *Recorder) Uint32(key string, value uint32) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint(uint64(value))
	}
	return recorder
}

func (recorder *Recorder) Uint64(key string, value uint64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeUint(value)
	}
	return recorder
}

func (recorder *Recorder) Float32(key string, value float32) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeFloat32(value)
	}
	return recorder
}

func (recorder *Recorder) Float64(key string, value float64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeFloat64(value)
	}
	return recorder
}

func (recorder *Recorder) Complex64(key string, value complex64) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeComplex64(value)
	}
	return recorder
}

func (recorder *Recorder) Complex128(key string, value complex128) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeComplex128(value)
	}
	return recorder
}

func (recorder *Recorder) Byte(key string, value byte) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeByte(value)
	}
	return recorder
}

func (recorder *Recorder) Rune(key string, value rune) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeRune(value)
	}
	return recorder
}

func (recorder *Recorder) Bool(key string, value bool) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeBool(value)
	}
	return recorder
}

func (recorder *Recorder) String(key string, value string) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeString(value)
	}
	return recorder
}

func (recorder *Recorder) Error(key string, value error) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		if value == nil {
			recorder.encoder.encodeNil()
		} else {
			recorder.encoder.buf = strconv.AppendQuote(recorder.encoder.buf, value.Error())
		}
	}
	return recorder
}

func (recorder *Recorder) Any(key string, value interface{}) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		if value == nil {
			recorder.encoder.encodeNil()
		} else {
			switch x := value.(type) {
			case error:
				recorder.encoder.encodeString(x.Error())
			case fmt.Stringer:
				recorder.encoder.encodeString(x.String())
			case string:
				recorder.encoder.encodeString(x)
			case appendFormatter:
				recorder.encoder.buf = x.AppendFormat(recorder.encoder.buf)
			default:
				if !recorder.encoder.encodeScalar(value) {
					recorder.encoder.encodeString(fmt.Sprintf("%v", value))
				}
			}
		}
	}
	return recorder
}

func (recorder *Recorder) Type(key string, value interface{}) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		if value == nil {
			recorder.encoder.encodeString("nil")
		} else {
			recorder.encoder.encodeString(reflect.TypeOf(value).String())
		}
	}
	return recorder
}

func (recorder *Recorder) Exec(key string, stringer func() string) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.encodeString(stringer())
	}
	return recorder
}

func (recorder *Recorder) writeTime(key string, value time.Time, layout string) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		recorder.encoder.buf = append(recorder.encoder.buf, '"')
		recorder.encoder.buf = value.AppendFormat(recorder.encoder.buf, layout)
		recorder.encoder.buf = append(recorder.encoder.buf, '"')
	}
	return recorder
}

func (recorder *Recorder) Date(key string, value time.Time) *Recorder {
	return recorder.writeTime(key, value, "2006-01-02Z07:00")
}

func (recorder *Recorder) Time(key string, value time.Time) *Recorder {
	return recorder.writeTime(key, value, time.RFC3339Nano)
}

func (recorder *Recorder) Seconds(key string, value time.Time) *Recorder {
	return recorder.writeTime(key, value, time.RFC3339)
}

func (recorder *Recorder) Milliseconds(key string, value time.Time) *Recorder {
	return recorder.writeTime(key, value, "2006-01-02T15:04:05.999Z07:00")
}

func (recorder *Recorder) Microseconds(key string, value time.Time) *Recorder {
	return recorder.writeTime(key, value, "2006-01-02T15:04:05.999999Z07:00")
}

func (recorder *Recorder) Duration(key string, value time.Duration) *Recorder {
	if recorder != nil {
		recorder.encoder.encodeKey(key)
		const reserved = 32
		l := len(recorder.encoder.buf)
		if cap(recorder.encoder.buf)-l < reserved {
			recorder.encoder.grow(reserved)
		}
		n := formatDuration(recorder.encoder.buf[l:l+reserved], value)
		recorder.encoder.buf = recorder.encoder.buf[:l+n]
	}
	return recorder
}
