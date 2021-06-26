package log

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// These flags define which text to prefix to each log entry generated by the Logger.
// Bits are or'ed together to control what's printed.
const (
	Ldatetime     = 1 << iota // the datetime in the local time zone: 2001/02/03 01:23:23
	Lshortfile                // final file name element and line number: d.go:23. overrides Llongfile
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	LUTC                      // if Ldatetime is set, use UTC rather than the local time zone
	LdefaultFlags = Ldatetime // default values for the standard logger
)

// Level represents log level
type Level int32

// Level constants
const (
	_       Level = iota // 0
	LvFATAL              // 1
	LvERROR              // 2
	LvWARN               // 3
	LvINFO               // 4
	LvDEBUG              // 5
	LvTRACE              // 6

	numLevel = 6
)

const levelBytes = "FEWIDT"

func getLevelByte(level Level) byte {
	i := int(level - LvFATAL)
	if i < 0 || i >= len(levelBytes) {
		return 'X'
	}
	return levelBytes[i]
}

var errUnrecognizedLevel = errors.New("log: unrecognized level")

func (level Level) index() int { return int(level - 1) }

// Set implements flag.Value interface such that you can use level  as a command as following:
//
//	var level logger.Level
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
	case LvFATAL:
		return "FATAL"
	case LvERROR:
		return "ERROR"
	case LvWARN:
		return "WARN"
	case LvINFO:
		return "INFO"
	case LvDEBUG:
		return "DEBUG"
	case LvTRACE:
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
	case "FATAL", "F", LvFATAL.Literal():
		return LvFATAL, true
	case "ERROR", "E", LvERROR.Literal():
		return LvERROR, true
	case "WARN", "W", LvWARN.Literal():
		return LvWARN, true
	case "INFO", "I", LvINFO.Literal():
		return LvINFO, true
	case "DEBUG", "D", LvDEBUG.Literal():
		return LvDEBUG, true
	case "TRACE", "T", LvTRACE.Literal():
		return LvTRACE, true
	}
	return LvINFO, false
}

// httpHandlerGetLevel returns a http handler for getting log level.
func httpHandlerGetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, GetLevel().String())
	})
}

// httpHandlerSetLevel sets new log level and returns old log level,
// Returns status code `StatusBadRequest` if parse log level fail.
func httpHandlerSetLevel() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		level := r.FormValue("level")
		lv, ok := ParseLevel(level)
		// invalid parameter
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid log level: "+level)
			return
		}
		// not modified
		oldLevel := GetLevel()
		if lv == oldLevel {
			w.WriteHeader(http.StatusNotModified)
			io.WriteString(w, oldLevel.String())
			return
		}
		// updated
		SetLevel(lv)
		io.WriteString(w, oldLevel.String())
	})
}

var (
	registerHTTPHandlersOnce sync.Once
)

func registerHTTPHandlers() {
	registerHTTPHandlersOnce.Do(func() {
		http.Handle("/log/level/get", httpHandlerGetLevel())
		http.Handle("/log/level/set", httpHandlerSetLevel())
	})
}

// Printer represents the printer for logging
type Printer interface {
	// Start starts the printer
	Start()
	// Quit quits the printer
	Shutdown()
	// Flush flushs all queued logs
	Flush()
	// GetFlags returns the flags
	GetFlags() int
	// SetFlags sets the flags
	SetFlags(flags int)
	// GetLevel returns log level
	GetLevel() Level
	// SetLevel sets log level
	SetLevel(Level)
	// SetPrefix sets log prefix
	SetPrefix(string)
	// Printf outputs leveled logs with file, line and extra prefix.
	// If line <= 0, then file and line both are invalid.
	Printf(file string, line int, level Level, prefix, format string, args ...interface{})
}

// stack returns the call stack
func stack(calldepth int) []byte {
	var (
		e             = make([]byte, 1<<16) // 64k
		nbytes        = runtime.Stack(e, false)
		ignorelinenum = 2*calldepth + 1
		count         = 0
		startIndex    = 0
	)
	for i := range e {
		if e[i] == '\n' {
			count++
			if count == ignorelinenum {
				startIndex = i + 1
				break
			}
		}
	}
	return e[startIndex:nbytes]
}

// printer implements Printer
type printer struct {
	level  Level
	prefix string
	writer Writer
	flags  int32

	entryListLocker sync.Mutex
	entryList       *entry

	async bool

	// used if async==false
	writeLocker sync.Mutex

	// used if async==true
	running int32
	queue   *queue
	queueMu sync.Mutex
	cond    *sync.Cond
	flush   chan chan struct{}
	quit    chan struct{}
	wait    chan struct{}
}

// newPrinter creates built in printer
func newPrinter(writer Writer, async bool) Printer {
	p := &printer{
		writer:    writer,
		entryList: new(entry),
		async:     async,
	}
	if async {
		p.queue = newQueue()
		p.cond = sync.NewCond(&p.queueMu)
		p.flush = make(chan chan struct{}, 1)
		p.quit = make(chan struct{})
		p.wait = make(chan struct{})
	}
	return p
}

// Start implements Printer Start method
func (p *printer) Start() {
	if p.queue == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return
	}
	go p.run()
}

func (p *printer) run() {
	for {
		p.cond.L.Lock()
		if p.queue.size() == 0 {
			p.cond.Wait()
		}
		entries := p.queue.popAll()
		p.cond.L.Unlock()
		p.writeEntries(entries)
		if p.consumeSignals() {
			break
		}
	}
}

func (p *printer) consumeSignals() bool {
	for {
		select {
		case resp := <-p.flush:
			p.flushAll()
			close(resp)
			continue
		case <-p.quit:
			p.flushAll()
			close(p.wait)
			return true
		default:
			return false
		}
	}
}

func (p *printer) flushAll() {
	p.cond.L.Lock()
	entries := p.queue.popAll()
	p.cond.L.Unlock()
	p.writeEntries(entries)
}

func (p *printer) writeEntries(entries []*entry) {
	for _, e := range entries {
		p.writeEntry(e)
	}
}

// Shutdown implements Printer Shutdown method
func (p *printer) Shutdown() {
	if p.queue == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return
	}
	close(p.quit)
	p.cond.Signal()
	<-p.wait
	p.writer.Close()
}

// Flush implements Printer Flush method
func (p *printer) Flush() {
	wait := make(chan struct{})
	p.flush <- wait
	p.cond.Signal()
	<-wait
}

// GetFlags implements Printer GetFlags method
func (p *printer) GetFlags() int {
	return int(atomic.LoadInt32(&p.flags))
}

// SetFlags implements Printer SetFlags method
func (p *printer) SetFlags(flags int) {
	atomic.StoreInt32(&p.flags, int32(flags))
}

// GetLevel implements Printer GetLevel method
func (p *printer) GetLevel() Level {
	return Level(atomic.LoadInt32((*int32)(&p.level)))
}

// SetLevel implements Printer SetLevel method
func (p *printer) SetLevel(level Level) {
	atomic.StoreInt32((*int32)(&p.level), int32(level))
}

// SetPrefix implements Printer SetPrefix method, SetPrefix is not concurrent-safe
func (p *printer) SetPrefix(prefix string) {
	p.prefix = prefix
}

// Printf implements Printer Printf method
func (p *printer) Printf(file string, line int, level Level, prefix, format string, args ...interface{}) {
	flags := p.GetFlags()
	p.output(flags, level, file, line, prefix, format, args...)
	if level == LvFATAL {
		p.Shutdown()
		os.Exit(1)
	}
}

func (p *printer) writeEntry(e *entry) {
	p.writer.Write(e.level, e.Bytes(), e.header)
	p.putEntry(e)
}

func (p *printer) getEntry() *entry {
	p.entryListLocker.Lock()
	if b := p.entryList; b != nil {
		p.entryList = b.next
		b.next = nil
		b.Reset()
		p.entryListLocker.Unlock()
		return b
	}
	p.entryListLocker.Unlock()
	return new(entry)
}

func (p *printer) putEntry(e *entry) {
	if e.Len() > 256 {
		return
	}
	p.entryListLocker.Lock()
	e.next = p.entryList
	p.entryList = e
	p.entryListLocker.Unlock()
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (p *printer) formatHeader(level Level, file string, line, flags int) *entry {
	if line < 0 {
		line = 0
	}
	var (
		e      = p.getEntry()
		caller = len(file) > 0
		off    int
	)
	e.tmp[0] = '['
	e.tmp[1] = getLevelByte(level)
	off = 2
	if flags&Ldatetime != 0 {
		now := time.Now()
		if flags&LUTC != 0 {
			now = now.In(time.UTC)
		}
		year, month, day := now.Date()
		hour, minute, second := now.Clock()
		millisecond := now.Nanosecond() / 1e6
		e.tmp[2] = ' '
		fourDigits(e, 3, year)
		e.tmp[7] = '/'
		twoDigits(e, 8, int(month))
		e.tmp[10] = '/'
		twoDigits(e, 11, day)
		e.tmp[13] = ' '
		twoDigits(e, 14, hour)
		e.tmp[16] = ':'
		twoDigits(e, 17, minute)
		e.tmp[19] = ':'
		twoDigits(e, 20, second)
		e.tmp[22] = '.'
		threeDigits(e, 23, millisecond)
		off = 26
	}
	if caller {
		e.tmp[off] = ' '
		e.Write(e.tmp[:off+1])
		e.WriteString(file)
		e.tmp[0] = ':'
		n := someDigits(e, 1, line)
		e.tmp[n+1] = ']'
		e.tmp[n+2] = ' '
		e.Write(e.tmp[:n+3])
	} else {
		e.tmp[off] = ']'
		e.tmp[off+1] = ' '
		e.Write(e.tmp[:off+2])
	}

	return e
}

func (p *printer) output(flags int, level Level, file string, line int, prefix, format string, args ...interface{}) {
	if flags&(Lshortfile|Llongfile) != 0 {
		if line <= 0 {
			file = "???"
			line = 0
		} else if flags&Lshortfile != 0 {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}
		}
	}
	e := p.formatHeader(level, file, line, flags)
	e.header = e.Len()
	if len(p.prefix) > 0 {
		e.WriteByte('(')
		e.WriteString(p.prefix)
		if len(prefix) > 0 {
			e.WriteByte('/')
			e.WriteString(prefix)
		}
		e.WriteString(") ")
	} else if len(prefix) > 0 {
		e.WriteByte('(')
		e.WriteString(prefix)
		e.WriteString(") ")
	}
	if len(args) == 0 {
		e.WriteString(format)
	} else {
		fmt.Fprintf(e, format, args...)
	}
	if e.Len() == 0 {
		return
	}
	if e.Bytes()[e.Len()-1] != '\n' {
		e.WriteByte('\n')
	}
	if level == LvFATAL {
		stackBuf := stack(4)
		e.WriteString("========= BEGIN STACK TRACE =========\n")
		e.Write(stackBuf)
		e.WriteString("========== END STACK TRACE ==========\n")
	}
	e.level = level
	if p.queue != nil && atomic.LoadInt32(&p.running) != 0 {
		p.cond.L.Lock()
		if p.queue.push(e) == 1 {
			p.cond.Signal()
		}
		p.cond.L.Unlock()
	} else {
		p.writeLocker.Lock()
		p.writeEntry(e)
		p.writeLocker.Unlock()
	}
}

// emptyPrinter wraps golang standard log package
type emptyPrinter struct{}

func (emptyPrinter) Start()                                                         {}
func (emptyPrinter) Shutdown()                                                      {}
func (emptyPrinter) Flush()                                                         {}
func (emptyPrinter) SetPrefix(prefix string)                                        {}
func (emptyPrinter) GetFlags() int                                                  { return 0 }
func (emptyPrinter) SetFlags(flags int)                                             {}
func (emptyPrinter) GetLevel() Level                                                { return LvINFO }
func (emptyPrinter) SetLevel(level Level)                                           {}
func (emptyPrinter) Printf(_ string, _ int, _ Level, _, _ string, _ ...interface{}) {}

// global printer
var gprinter Printer = emptyPrinter{}

type startOptions struct {
	httpHandler bool
	flags       int
	sync        bool
	level       Level
	prefix      string
	printer     Printer
	writers     []Writer
	errors      []error
}

func defaultStartOptions() startOptions {
	return startOptions{
		flags: LdefaultFlags,
		level: LvINFO,
	}
}

func (opt *startOptions) apply(options []Option) {
	for i := range options {
		if options[i] != nil {
			options[i](opt)
		}
	}
}

// Option is option for Start
type Option func(*startOptions)

func errOption(err error) Option {
	return func(opt *startOptions) {
		opt.errors = append(opt.errors, err)
	}
}

// WithSync synchronize outputs log or not
func WithSync(yes bool) Option {
	return func(opt *startOptions) {
		opt.sync = yes
	}
}

// WithHTTPHandler enable or disable http handler for settting level
func WithHTTPHandler(yes bool) Option {
	return func(opt *startOptions) {
		opt.httpHandler = yes
	}
}

// WithFlags enable or disable flags information
func WithFlags(flags int) Option {
	return func(opt *startOptions) {
		opt.flags = flags
	}
}

// WithLevel sets log level
func WithLevel(level Level) Option {
	return func(opt *startOptions) {
		opt.level = level
	}
}

// WithPrefix set log prefix
func WithPrefix(prefix string) Option {
	return func(opt *startOptions) {
		opt.prefix = prefix
	}
}

// WithPrinter specify custom printer
func WithPrinter(printer Printer) Option {
	if printer == nil {
		panic("log: with a nil printer")
	}
	return func(opt *startOptions) {
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
	return func(opt *startOptions) {
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

// Start starts logging with options
func Start(options ...Option) error {
	var opt = defaultStartOptions()
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
			opt.printer = gprinter
			changed = false
		case 1:
			opt.printer = newPrinter(opt.writers[0], async)
		default:
			opt.printer = newPrinter(multiWriter{opt.writers}, async)
		}
	}
	if opt.level != 0 {
		opt.printer.SetLevel(opt.level)
	}
	opt.printer.SetPrefix(opt.prefix)
	opt.printer.SetFlags(opt.flags)

	if changed {
		gprinter.Shutdown()
		gprinter = opt.printer
		gprinter.Start()
	}
	if opt.httpHandler {
		registerHTTPHandlers()
	}
	return nil
}

// Shutdown shutdowns global printer
func Shutdown() {
	gprinter.Shutdown()
}

// GetFlags returns the flags
func GetFlags() {
	gprinter.GetFlags()
}

// SetFlags sets the flags
func SetFlags(flags int) {
	gprinter.SetFlags(flags)
}

// GetLevel returns current log level
func GetLevel() Level {
	return gprinter.GetLevel()
}

// SetLevel sets current log level
func SetLevel(level Level) {
	gprinter.SetLevel(level)
}

// Trace creates a context fields with level trace
//loglint: Trace
func Trace() *Fields { return getFields(LvTRACE, "") }

// Debug creates a context fields with level debug
//loglint: Debug
func Debug() *Fields { return getFields(LvDEBUG, "") }

// Info creates a context fields with level info
//loglint: Info
func Info() *Fields { return getFields(LvINFO, "") }

// Warn creates a context fields with level warn
//loglint: Warn
func Warn() *Fields { return getFields(LvWARN, "") }

// Error creates a context fields with level error
//loglint: Error
func Error() *Fields { return getFields(LvERROR, "") }

// Fatal creates a context fields with level fatal
//loglint: Fatal
func Fatal() *Fields { return getFields(LvFATAL, "") }

// Printf wraps the global printer Printf method
func Printf(level Level, format string, args ...interface{}) {
	if gprinter.GetLevel() < level {
		return
	}
	var (
		file string
		line int
	)
	if gprinter.GetFlags()&(Lshortfile|Llongfile) != 0 {
		_, file, line, _ = runtime.Caller(1)
	}
	gprinter.Printf(file, line, level, "", format, args...)
}

// Log is a low-level API to print log.
func Log(calldepth int, level Level, prefix, format string, args ...interface{}) {
	if gprinter.GetLevel() < level {
		return
	}
	var (
		file string
		line int
	)
	if gprinter.GetFlags()&(Lshortfile|Llongfile) != 0 {
		_, file, line, _ = runtime.Caller(calldepth)
	}
	gprinter.Printf(file, line, level, prefix, format, args...)
}

// Prefix wraps a string as a prefixed logger
type Prefix string

// Trace creates a context fields with level trace
//loglint:method Prefix.Trace
func (p Prefix) Trace() *Fields { return getFields(LvTRACE, p) }

// Debug creates a context fields with level debug
//loglint:method Prefix.Debug
func (p Prefix) Debug() *Fields { return getFields(LvDEBUG, p) }

// Info creates a context fields with level info
//loglint:method Prefix.Info
func (p Prefix) Info() *Fields { return getFields(LvINFO, p) }

// Warn creates a context fields with level warn
//loglint:method Prefix.Warn
func (p Prefix) Warn() *Fields { return getFields(LvWARN, p) }

// Error creates a context fields with level error
//loglint:method Prefix.Error
func (p Prefix) Error() *Fields { return getFields(LvERROR, p) }

// Fatal creates a context fields with level fatal
//loglint:method Prefix.Fatal
func (p Prefix) Fatal() *Fields { return getFields(LvFATAL, p) }

// Printf wraps the global printer Printf method
func (p Prefix) Printf(level Level, format string, args ...interface{}) {
	if gprinter.GetLevel() < level {
		return
	}
	var (
		file string
		line int
	)
	if gprinter.GetFlags()&(Lshortfile|Llongfile) != 0 {
		_, file, line, _ = runtime.Caller(1)
	}
	gprinter.Printf(file, line, level, string(p), format, args...)
}

// Prefix appends a prefix to current prefix
func (p Prefix) Prefix(prefix string) Prefix {
	return p + "/" + Prefix(prefix)
}
