package log

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Caller holds caller information
type Caller struct {
	Filename string
	Line     int
}

// Provider represents the provider for logging
type Provider interface {
	// Start starts the provider
	Start() error
	// Shutdown shutdowns the provider
	Shutdown() error
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

// provider implements Provider
type provider struct {
	writer Writer

	entryListLocker sync.Mutex
	entryList       *entry

	async bool

	// used for async==false
	writeLocker sync.Mutex

	// used for async==true
	running int32
	queue   *queue
	queueMu sync.Mutex
	cond    *sync.Cond
	flush   chan chan struct{}
	quit    chan struct{}
	wait    chan struct{}
}

// newProvider creates built in provider
func newProvider(writer Writer, async bool) Provider {
	p := &provider{
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

// Start implements Provider Start method
func (p *provider) Start() error {
	if p.queue == nil {
		return errors.New("queue is nil")
	}
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return errors.New("provider already running")
	}
	go p.run()
	return nil
}

func (p *provider) run() {
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

func (p *provider) consumeSignals() bool {
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

func (p *provider) flushAll() {
	p.cond.L.Lock()
	entries := p.queue.popAll()
	p.cond.L.Unlock()
	p.writeEntries(entries)
}

func (p *provider) writeEntries(entries []*entry) {
	for _, e := range entries {
		p.writeEntry(e)
	}
}

// Shutdown implements Provider Shutdown method
func (p *provider) Shutdown() error {
	if p.queue == nil {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return nil
	}
	close(p.quit)
	p.cond.Signal()
	<-p.wait
	p.writer.Close()
	return nil
}

// Print implements Provider Print method
func (p *provider) Print(level Level, flags int, caller Caller, prefix, msg string) {
	p.output(level, flags, caller, prefix, msg)
	if level == LevelFatal {
		p.Shutdown()
		os.Exit(1)
	}
}

func (p *provider) writeEntry(e *entry) {
	p.writer.Write(e.level, e.buf.Bytes(), e.header)
	p.putEntry(e)
}

func (p *provider) getEntry() *entry {
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

func (p *provider) putEntry(e *entry) {
	if e.buf.Len() > 256 {
		return
	}
	p.entryListLocker.Lock()
	e.next = p.entryList
	p.entryList = e
	p.entryListLocker.Unlock()
}

// [L yyyy/MM/dd hh:mm:ss.uuu file:line]
func (p *provider) formatHeader(level Level, caller Caller, flags int) *entry {
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

func (p *provider) output(level Level, flags int, caller Caller, prefix, msg string) {
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

type emptyProvider struct{}

var empty Provider = emptyProvider{}

func (emptyProvider) Start() error                                { return nil }
func (emptyProvider) Shutdown() error                             { return nil }
func (emptyProvider) Print(_ Level, _ int, _ Caller, _, _ string) {}
