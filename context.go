package log

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Context holds context ctx
type Context struct {
	logger  *Logger
	level   Level
	prefix  string
	encoder encoder
}

var ctxPool = sync.Pool{
	New: func() interface{} {
		return new(Context)
	},
}

func getContext(logger *Logger, level Level, prefix string) *Context {
	if logger == nil || logger.GetLevel() < level {
		return nil
	}
	ctx := ctxPool.Get().(*Context)
	ctx.reset(logger, level, prefix)
	return ctx
}

func putContext(ctx *Context) {
	if ctx.encoder.Cap() < 1024 {
		ctxPool.Put(ctx)
	}
}

func (ctx *Context) reset(logger *Logger, level Level, prefix string) {
	ctx.logger = logger
	ctx.level = level
	ctx.prefix = prefix
	ctx.encoder.reset()
}

// Print prints logging with context ctx. After this call,
// the ctx not available.
func (ctx *Context) Print(msg string) {
	if ctx == nil {
		return
	}
	ctx.encoder.finish()
	ctx.encoder.writeString(msg)
	var (
		caller Caller
		flags  = ctx.logger.GetFlags()
	)
	if flags&(Lshortfile|Llongfile) != 0 {
		_, caller.Filename, caller.Line, _ = runtime.Caller(1)
	}
	ctx.logger.provider.Print(ctx.level, flags, caller, ctx.prefix, ctx.encoder.String())
	putContext(ctx)
}

// Printf prints logging with context ctx by format. After this call,
// the ctx not available.
func (ctx *Context) Printf(msg string, a ...interface{}) {
	if ctx == nil {
		return
	}
	ctx.encoder.finish()
	fmt.Fprintf(&ctx.encoder, msg, a...)
	var (
		caller Caller
		flags  = ctx.logger.GetFlags()
	)
	if flags&(Lshortfile|Llongfile) != 0 {
		_, caller.Filename, caller.Line, _ = runtime.Caller(1)
	}
	ctx.logger.provider.Print(ctx.level, flags, caller, ctx.prefix, ctx.encoder.String())
	putContext(ctx)
}

// Int puts an integer value for key
func (ctx *Context) Int(key string, value int) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeInt(int64(value))
	}
	return ctx
}

// Int8 puts an 8-bits integer value for key
func (ctx *Context) Int8(key string, value int8) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeInt(int64(value))
	}
	return ctx
}

// Int16 puts a 16-bits integer value for key
func (ctx *Context) Int16(key string, value int16) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeInt(int64(value))
	}
	return ctx
}

// Int32 puts a 32-bits integer value for key
func (ctx *Context) Int32(key string, value int32) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeInt(int64(value))
	}
	return ctx
}

// Int64 puts a 64-bits integer value for key
func (ctx *Context) Int64(key string, value int64) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeInt(value)
	}
	return ctx
}

// Uint puts an unsigned integer value for key
func (ctx *Context) Uint(key string, value uint) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeUint(uint64(value))
	}
	return ctx
}

// Uint8 puts an 8-bits unsigned integer value for key
func (ctx *Context) Uint8(key string, value uint8) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeUint(uint64(value))
	}
	return ctx
}

// Uint16 puts a 16-bits unsigned integer value for key
func (ctx *Context) Uint16(key string, value uint16) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeUint(uint64(value))
	}
	return ctx
}

// Uint32 puts a 32-bits unsigned integer value for key
func (ctx *Context) Uint32(key string, value uint32) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeUint(uint64(value))
	}
	return ctx
}

// Uint64 puts a 64-bits unsigned integer value for key
func (ctx *Context) Uint64(key string, value uint64) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeUint(value)
	}
	return ctx
}

// Float32 puts a 32-bits floating value for key
func (ctx *Context) Float32(key string, value float32) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeFloat32(value)
	}
	return ctx
}

// Float64 puts a 64-bits floating value for key
func (ctx *Context) Float64(key string, value float64) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeFloat64(value)
	}
	return ctx
}

// Complex64 puts a 64-bits complex value for key
func (ctx *Context) Complex64(key string, value complex64) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeComplex64(value)
	}
	return ctx
}

// Complex128 puts a 128-bits complex value for key
func (ctx *Context) Complex128(key string, value complex128) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeComplex128(value)
	}
	return ctx
}

// Byte puts a byte value for key
func (ctx *Context) Byte(key string, value byte) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeByte(value)
	}
	return ctx
}

// Rune puts a rune value for key
func (ctx *Context) Rune(key string, value rune) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeRune(value)
	}
	return ctx
}

// Bool puts a boolean value for key
func (ctx *Context) Bool(key string, value bool) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeBool(value)
	}
	return ctx
}

// String puts a string value for key
func (ctx *Context) String(key string, value string) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeString(value)
	}
	return ctx
}

// Error puts an error value for key
func (ctx *Context) Error(key string, value error) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		if value == nil {
			ctx.encoder.encodeNil()
		} else {
			ctx.encoder.buf = strconv.AppendQuote(ctx.encoder.buf, value.Error())
		}
	}
	return ctx
}

// Any puts an any value for key
func (ctx *Context) Any(key string, value interface{}) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		if value == nil {
			ctx.encoder.encodeNil()
		} else {
			switch x := value.(type) {
			case error:
				ctx.encoder.encodeString(x.Error())
			case fmt.Stringer:
				ctx.encoder.encodeString(x.String())
			case string:
				ctx.encoder.encodeString(x)
			case appendFormatter:
				ctx.encoder.buf = x.AppendFormat(ctx.encoder.buf)
			default:
				if !ctx.encoder.encodeScalar(value) {
					ctx.encoder.encodeString(fmt.Sprintf("%v", value))
				}
			}
		}
	}
	return ctx
}

// Type puts a type info of value for key
func (ctx *Context) Type(key string, value interface{}) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		if value == nil {
			ctx.encoder.encodeString("nil")
		} else {
			ctx.encoder.encodeString(reflect.TypeOf(value).String())
		}
	}
	return ctx
}

// Exec puts a result value of function for key
func (ctx *Context) Exec(key string, stringer func() string) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.encodeString(stringer())
	}
	return ctx
}

func (ctx *Context) writeTime(key string, value time.Time, layout string) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		ctx.encoder.buf = append(ctx.encoder.buf, '"')
		ctx.encoder.buf = value.AppendFormat(ctx.encoder.buf, layout)
		ctx.encoder.buf = append(ctx.encoder.buf, '"')
	}
	return ctx
}

// Date puts a date for key
func (ctx *Context) Date(key string, value time.Time) *Context {
	return ctx.writeTime(key, value, "2006-01-02Z07:00")
}

// Time puts a time for key
func (ctx *Context) Time(key string, value time.Time) *Context {
	return ctx.writeTime(key, value, time.RFC3339Nano)
}

// Clock puts a clock for key
func (ctx *Context) Clock(key string, value time.Time) *Context {
	return ctx.writeTime(key, value, "15:04:05")
}

// Seconds puts a time accurate to the second for key
func (ctx *Context) Seconds(key string, value time.Time) *Context {
	return ctx.writeTime(key, value, time.RFC3339)
}

// Milliseconds puts a time accurate to the millisecond for key
func (ctx *Context) Milliseconds(key string, value time.Time) *Context {
	return ctx.writeTime(key, value, "2006-01-02T15:04:05.999Z07:00")
}

// Microseconds puts a time accurate to the microsecond for key
func (ctx *Context) Microseconds(key string, value time.Time) *Context {
	return ctx.writeTime(key, value, "2006-01-02T15:04:05.999999Z07:00")
}

// Duration puts a duration value for key
func (ctx *Context) Duration(key string, value time.Duration) *Context {
	if ctx != nil {
		ctx.encoder.encodeKey(key)
		const reserved = 32
		l := len(ctx.encoder.buf)
		if cap(ctx.encoder.buf)-l < reserved {
			ctx.encoder.grow(reserved)
		}
		n := formatDuration(ctx.encoder.buf[l:l+reserved], value)
		ctx.encoder.buf = ctx.encoder.buf[:l+n]
	}
	return ctx
}
