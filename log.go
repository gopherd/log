package log

import (
	"errors"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

// These flags define which text to prefix to each log entry generated by the Logger.
// Bits are or'ed together to control what's printed.
const (
	Ltimestamp    = 1 << iota                  // the timestamp in the local time zone: 2001/02/03 01:23:23
	LUTC                                       // if Ltimestamp is set, use UTC rather than the local time zone
	Lmicroseconds                              // microsecond resolution: 01:23:23.123123.  assumes Ltimestamp.
	Lshortfile                                 // final file name element and line number: d.go:23. overrides Llongfile
	Llongfile                                  // full file name and line number: /a/b/c/d.go:23
	LdefaultFlags = Ltimestamp | Lmicroseconds // default values for the standard logger
)

// Level represents log level
type Level int32

// Level constants
const (
	_          Level = iota // 0
	LevelFatal              // 1
	LevelError              // 2
	LevelWarn               // 3
	LevelInfo               // 4
	LevelDebug              // 5
	LevelTrace              // 6

	numLevel = 6
)

const levelBytes = "FEWIDT"

func getLevelByte(level Level) byte {
	i := int(level - LevelFatal)
	if i < 0 || i >= len(levelBytes) {
		return 'X'
	}
	return levelBytes[i]
}

var errUnrecognizedLevel = errors.New("log: unrecognized level")

func (level Level) index() int { return int(level - 1) }

// Set implements flag.Value interface such that you can use level  as a command as following:
//
//	var level log.Level
//	flag.Var(&level, "log_level", "log level: trace/debug/info/warn/error/fatal")
func (level *Level) Set(s string) error {
	lv, ok := ParseLevel(s)
	*level = lv
	if !ok {
		return errUnrecognizedLevel
	}
	return nil
}

// Literal returns literal value which is a number
func (level Level) Literal() string {
	return strconv.Itoa(int(level))
}

// String returns a serialized string of level
func (level Level) String() string {
	switch level {
	case LevelFatal:
		return "FATAL"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelTrace:
		return "TRACE"
	}
	return "(" + strconv.Itoa(int(level)) + ")"
}

// MarshalJSON implements json.Marshaler
func (level Level) MarshalJSON() ([]byte, error) {
	return []byte(`"` + level.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (level *Level) UnmarshalJSON(data []byte) error {
	var (
		s   string
		err error
	)
	if len(data) >= 2 {
		s, err = strconv.Unquote(string(data))
		if err != nil {
			return err
		}
	} else {
		s = string(data)
	}
	return level.Set(s)
}

// MoreVerboseThan returns whether level more verbose than other
func (level Level) MoreVerboseThan(other Level) bool { return level > other }

// ParseLevel parses log level from string
func ParseLevel(s string) (lv Level, ok bool) {
	s = strings.ToUpper(s)
	switch s {
	case "FATAL", "F", LevelFatal.Literal():
		return LevelFatal, true
	case "ERROR", "E", LevelError.Literal():
		return LevelError, true
	case "WARN", "W", LevelWarn.Literal():
		return LevelWarn, true
	case "INFO", "I", LevelInfo.Literal():
		return LevelInfo, true
	case "DEBUG", "D", LevelDebug.Literal():
		return LevelDebug, true
	case "TRACE", "T", LevelTrace.Literal():
		return LevelTrace, true
	}
	return LevelInfo, false
}

type options struct {
	flags   int
	sync    bool
	level   Level
	prefix  string
	printer Printer
	writers []Writer
	errors  []error
}

func defaultOptions() options {
	return options{
		flags: LdefaultFlags,
		level: LevelInfo,
	}
}

func (opt *options) apply(options []Option) {
	for i := range options {
		if options[i] != nil {
			options[i](opt)
		}
	}
}

// Option is option for Start
type Option func(*options)

func errOption(err error) Option {
	return func(opt *options) {
		opt.errors = append(opt.errors, err)
	}
}

// WithSync synchronize outputs log or not
func WithSync(yes bool) Option {
	return func(opt *options) {
		opt.sync = yes
	}
}

// WithFlags enable or disable flags information
func WithFlags(flags int) Option {
	return func(opt *options) {
		opt.flags = flags
	}
}

// WithLevel sets log level
func WithLevel(level Level) Option {
	return func(opt *options) {
		opt.level = level
	}
}

// WithPrinter specify custom printer
func WithPrinter(printer Printer) Option {
	if printer == nil {
		panic("log: with a nil printer")
	}
	return func(opt *options) {
		if opt.printer != nil {
			panic("log: printer already specified")
		}
		if len(opt.writers) > 0 {
			panic("log: couldn't specify printer if any writer specified")
		}
		opt.printer = printer
	}
}

// WithWriters appends the writers
func WithWriters(writers ...Writer) Option {
	if len(writers) == 0 {
		return nil
	}
	for i, writer := range writers {
		if writer == nil {
			panic("log: with a nil(" + strconv.Itoa(i+1) + "th) writer")
		}
	}
	var copied = make([]Writer, len(writers))
	copy(copied, writers)
	return func(opt *options) {
		if opt.printer != nil {
			panic("log: couldn't specify writer if a printer specified")
		}
		opt.writers = append(opt.writers, copied...)
	}
}

// WithOutput appends a console writer with specified io.Writer
func WithOutput(w io.Writer) Option {
	return WithWriters(newConsole(w))
}

// WithFile appends a file writer
func WithFile(fileOptions FileOptions) Option {
	f, err := newFile(fileOptions)
	if err != nil {
		return errOption(err)
	}
	return WithWriters(f)
}

// WithMultiFile appends a multifile writer
func WithMultiFile(multiFileOptions MultiFileOptions) Option {
	return WithWriters(newMultiFile(multiFileOptions))
}

// Logger is the top-level object for outputing log message
type Logger struct {
	printer Printer
	prefix  string
	level   int32
	flags   int32
}

// NewLogger creates logger
func NewLogger(prefix string) *Logger {
	return &Logger{
		printer: empty,
		level:   int32(LevelInfo),
		prefix:  prefix,
	}
}

// Start starts logging with options
func (logger *Logger) Start(options ...Option) error {
	var opt = defaultOptions()
	opt.apply(options)
	if len(opt.errors) > 0 {
		for i := range opt.writers {
			opt.writers[i].Close()
		}
		return opt.errors[0]
	}
	async := !opt.sync
	changed := true
	if opt.printer == nil {
		switch len(opt.writers) {
		case 0:
			opt.printer = logger.printer
			changed = false
		case 1:
			opt.printer = newPrinter(opt.writers[0], async)
		default:
			opt.printer = newPrinter(multiWriter{opt.writers}, async)
		}
	}
	if opt.level != 0 {
		logger.SetLevel(opt.level)
	}
	logger.SetFlags(opt.flags)

	if changed {
		logger.Shutdown()
		logger.printer = opt.printer
		logger.printer.Start()
	}
	return nil
}

// Shutdown shutdowns the logger
func (logger *Logger) Shutdown() {
	logger.printer.Shutdown()
}

// GetFlags returns the output flags
func (logger *Logger) GetFlags() int {
	return int(atomic.LoadInt32(&logger.flags))
}

// SetFlags sets the output flags
func (logger *Logger) SetFlags(flags int) {
	atomic.StoreInt32(&logger.flags, int32(flags))
}

// GetLevel returns the log level
func (logger *Logger) GetLevel() Level {
	return Level(atomic.LoadInt32(&logger.level))
}

// SetLevel sets the log level
func (logger *Logger) SetLevel(level Level) {
	atomic.StoreInt32(&logger.level, int32(level))
}

// Trace creates a context with level trace
func (logger *Logger) Trace() *Context { return getContext(logger, LevelTrace, logger.prefix) }

// Debug creates a context with level debug
func (logger *Logger) Debug() *Context { return getContext(logger, LevelDebug, logger.prefix) }

// Info creates a context with level info
func (logger *Logger) Info() *Context { return getContext(logger, LevelInfo, logger.prefix) }

// Warn creates a context with level warn
func (logger *Logger) Warn() *Context { return getContext(logger, LevelWarn, logger.prefix) }

// Error creates a context with level error
func (logger *Logger) Error() *Context { return getContext(logger, LevelError, logger.prefix) }

// Fatal creates a context with level fatal
func (logger *Logger) Fatal() *Context { return getContext(logger, LevelFatal, logger.prefix) }

// Log creates a context with specified level
func (logger *Logger) Log(level Level) *Context { return getContext(logger, level, logger.prefix) }

// Print is a low-level API to print log.
func (logger *Logger) Print(calldepth int, level Level, msg string) {
	if logger.GetLevel() < level {
		return
	}
	var (
		caller Caller
		flags  = logger.GetFlags()
	)
	if flags&(Lshortfile|Llongfile) != 0 {
		_, caller.Filename, caller.Line, _ = runtime.Caller(calldepth)
	}
	logger.printer.Print(level, flags, caller, logger.prefix, msg)
}

// default global logger
var DefaultLogger = NewLogger("")

// Start starts the global logger
func Start(options ...Option) error {
	return DefaultLogger.Start(options...)
}

// Shutdown shutdowns the global logger
func Shutdown() {
	DefaultLogger.Shutdown()
}

// GetFlags returns the output flags
func GetFlags() {
	DefaultLogger.GetFlags()
}

// SetFlags sets the output flags
func SetFlags(flags int) {
	DefaultLogger.SetFlags(flags)
}

// GetLevel returns the log level
func GetLevel() Level {
	return DefaultLogger.GetLevel()
}

// SetLevel sets the log level
func SetLevel(level Level) {
	DefaultLogger.SetLevel(level)
}

// Trace creates a context with level trace
func Trace() *Context { return getContext(DefaultLogger, LevelTrace, DefaultLogger.prefix) }

// Debug creates a context with level debug
func Debug() *Context { return getContext(DefaultLogger, LevelDebug, DefaultLogger.prefix) }

// Info creates a context with level info
func Info() *Context { return getContext(DefaultLogger, LevelInfo, DefaultLogger.prefix) }

// Warn creates a context with level warn
func Warn() *Context { return getContext(DefaultLogger, LevelWarn, DefaultLogger.prefix) }

// Error creates a context with level error
func Error() *Context { return getContext(DefaultLogger, LevelError, DefaultLogger.prefix) }

// Fatal creates a context with level fatal
func Fatal() *Context { return getContext(DefaultLogger, LevelFatal, DefaultLogger.prefix) }

// Log creates a context with specified level
func Log(level Level) *Context { return getContext(DefaultLogger, level, DefaultLogger.prefix) }

// Print is a low-level API to print log.
func Print(calldepth int, level Level, msg string) {
	if DefaultLogger.GetLevel() < level {
		return
	}
	var (
		caller Caller
		flags  = DefaultLogger.GetFlags()
	)
	if flags&(Lshortfile|Llongfile) != 0 {
		_, caller.Filename, caller.Line, _ = runtime.Caller(calldepth)
	}
	DefaultLogger.printer.Print(level, flags, caller, DefaultLogger.prefix, msg)
}

// ContextLogger holds a prefixed logger
type ContextLogger struct {
	logger *Logger
	prefix string
}

// Prefix creates a context logger with prefix
func Prefix(logger *Logger, prefix string) *ContextLogger {
	if logger == nil {
		logger = DefaultLogger
	}
	if logger.prefix != "" {
		prefix = logger.prefix + "/" + prefix
	}
	return &ContextLogger{
		logger: logger,
		prefix: prefix,
	}
}

// Logger returns underlying logger
func (p *ContextLogger) Logger() *Logger {
	return p.logger
}

// Prefix returns prefix of context logger
func (p *ContextLogger) Prefix() string {
	return p.prefix
}

// Trace creates a context with level trace
func (p *ContextLogger) Trace() *Context {
	return getContext(p.logger, LevelTrace, p.prefix)
}

// Debug creates a context with level debug
func (p *ContextLogger) Debug() *Context {
	return getContext(p.logger, LevelDebug, p.prefix)
}

// Info creates a context with level info
func (p *ContextLogger) Info() *Context {
	return getContext(p.logger, LevelInfo, p.prefix)
}

// Warn creates a context with level warn
func (p *ContextLogger) Warn() *Context {
	return getContext(p.logger, LevelWarn, p.prefix)
}

// Error creates a context with level error
func (p *ContextLogger) Error() *Context {
	return getContext(p.logger, LevelError, p.prefix)
}

// Fatal creates a context with level fatal
func (p *ContextLogger) Fatal() *Context {
	return getContext(p.logger, LevelFatal, p.prefix)
}

// Log creates a context with specified level
func (p *ContextLogger) Log(level Level) *Context {
	return getContext(p.logger, level, p.prefix)
}

// Print is a low-level API to print log.
func (p *ContextLogger) Print(calldepth int, level Level, msg string) {
	if p.logger.GetLevel() < level {
		return
	}
	var (
		caller Caller
		flags  = p.logger.GetFlags()
	)
	if flags&(Lshortfile|Llongfile) != 0 {
		_, caller.Filename, caller.Line, _ = runtime.Caller(calldepth)
	}
	p.logger.printer.Print(level, flags, caller, p.prefix, msg)
}
