package log

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Caller struct {
	Filename string
	Line     int
}

// Printer represents the printer for logging
type Printer interface {
	// Start starts the printer
	Start()
	// Shutdown shutdowns the printer
	Shutdown()
	// Print outputs leveled logs with file, line and extra prefix.
	// If line <= 0, then file and line both are invalid.
	Print(level Level, flags int, caller Caller, prefix, msg string)
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
	writer Writer

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

// Print implements Printer Print method
func (p *printer) Print(level Level, flags int, caller Caller, prefix, msg string) {
	p.output(level, flags, caller, prefix, msg)
	if level == LevelFatal {
		p.Shutdown()
		os.Exit(1)
	}
}

func (p *printer) writeEntry(e *entry) {
	p.writer.Write(e.level, e.buf.Bytes(), e.header)
	p.putEntry(e)
}

func (p *printer) getEntry() *entry {
	p.entryListLocker.Lock()
	if b := p.entryList; b != nil {
		p.entryList = b.next
		b.next = nil
		b.reset()
		p.entryListLocker.Unlock()
		return b
	}
	p.entryListLocker.Unlock()
	return new(entry)
}

func (p *printer) putEntry(e *entry) {
	if e.buf.Len() > 256 {
		return
	}
	p.entryListLocker.Lock()
	e.next = p.entryList
	p.entryList = e
	p.entryListLocker.Unlock()
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (p *printer) formatHeader(level Level, caller Caller, flags int) *entry {
	var (
		e   = p.getEntry()
		off int
	)
	e.tmp[0] = '['
	e.tmp[1] = getLevelByte(level)
	off = 2
	if flags&Ltimestamp != 0 {
		now := time.Now()
		if flags&LUTC != 0 {
			now = now.In(time.UTC)
		}
		year, month, day := now.Date()
		hour, minute, second := now.Clock()
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
		off = 22
		if flags&Lmicroseconds != 0 {
			e.tmp[off] = '.'
			off++
			sixDigits(e, off, now.Nanosecond()/1e3)
			off += 6
		}
	}
	if caller.Line > 0 {
		e.tmp[off] = ' '
		e.buf.Write(e.tmp[:off+1])
		e.buf.WriteString(caller.Filename)
		e.tmp[0] = ':'
		n := someDigits(e, 1, caller.Line)
		e.tmp[n+1] = ']'
		e.tmp[n+2] = ' '
		e.buf.Write(e.tmp[:n+3])
	} else {
		e.tmp[off] = ']'
		e.tmp[off+1] = ' '
		e.buf.Write(e.tmp[:off+2])
	}

	return e
}

func (p *printer) output(level Level, flags int, caller Caller, prefix, msg string) {
	if flags&(Lshortfile|Llongfile) != 0 {
		if caller.Line <= 0 {
			caller.Filename = "???"
			caller.Line = 0
		} else if flags&Lshortfile != 0 {
			slash := strings.LastIndex(caller.Filename, "/")
			if slash >= 0 {
				caller.Filename = caller.Filename[slash+1:]
			}
		}
	}
	e := p.formatHeader(level, caller, flags)
	e.header = e.buf.Len()
	if len(prefix) > 0 {
		e.buf.WriteByte('(')
		e.buf.WriteString(prefix)
		e.buf.WriteString(") ")
	}
	e.buf.WriteString(msg)
	if e.buf.Len() == 0 {
		return
	}
	if e.buf.Bytes()[e.buf.Len()-1] != '\n' {
		e.buf.WriteByte('\n')
	}
	if level == LevelFatal {
		stackBuf := stack(4)
		e.buf.WriteString("========= BEGIN STACK TRACE =========\n")
		e.buf.Write(stackBuf)
		e.buf.WriteString("========== END STACK TRACE ==========\n")
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

type emptyPrinter struct{}

var empty Printer = emptyPrinter{}

func (emptyPrinter) Start()                                                          {}
func (emptyPrinter) Shutdown()                                                       {}
func (emptyPrinter) Print(level Level, flags int, caller Caller, prefix, msg string) {}
