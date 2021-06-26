package log

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

func parseSize(s string) (int64, error) {
	n := len(s)
	if n == 0 {
		return 0, errors.New("invalid size")
	}
	u := s[n-1]
	x := int64(1)
	switch u {
	case 'k', 'K':
		x = KB
	case 'm', 'M':
		x = MB
	case 'g', 'G':
		x = GB
	}
	if x > 1 {
		s = s[:n-1]
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("invalid size")
	}
	return i * x, nil
}

var (
	errNilWriter = errors.New("log: nil writer")
)

// Writer represents a writer for logging
type Writer interface {
	Write(level Level, data []byte, headerLen int) error
	Close() error
}

type WriterCreator func(source string) (Writer, error)

var (
	writerCreatorsMu sync.RWMutex
	writerCreators   = make(map[string]WriterCreator)
)

func init() {
	Register("console", openConsole)
	Register("file", openFile)
	Register("multifile", openMultiFile)
}

func Register(name string, creator WriterCreator) {
	if creator == nil {
		panic("log: Register creator is nil")
	}
	writerCreatorsMu.Lock()
	defer writerCreatorsMu.Unlock()
	if _, dup := writerCreators[name]; dup {
		panic("log: Register called twice for " + name)
	}
	writerCreators[name] = creator
}

func Open(url string) (Writer, error) {
	var (
		name   string
		source string
	)
	i := strings.Index(url, ":")
	if i < 0 {
		name = url
	} else {
		name = url[:i]
		source = url[i+1:]
	}
	if name == "" {
		return nil, errors.New("log: writer name is empty, url format: `name[:source]`")
	}

	writerCreatorsMu.RLock()
	creator, ok := writerCreators[name]
	writerCreatorsMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("log: unknown writer %q (forgotten import?)", name)
	}
	return creator(source)
}

// multiWriter merges multi-writers
type multiWriter struct {
	writers []Writer
}

// Write writes log to all inner writers
func (w multiWriter) Write(level Level, data []byte, headerLen int) error {
	var lastErr error
	for i := range w.writers {
		if err := w.writers[i].Write(level, data, headerLen); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Close closes all inner writers
func (w multiWriter) Close() error {
	var lastErr error
	for i := range w.writers {
		if err := w.writers[i].Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// console is a writer that writes logs to console
type console struct {
	w io.Writer
}

// newConsole creates a console writer
func newConsole(w io.Writer) *console {
	return &console{
		w: w,
	}
}

func openConsole(source string) (Writer, error) {
	switch source {
	case "stdout":
		return newConsole(os.Stdout), nil
	case "", "stderr":
		return newConsole(os.Stderr), nil
	default:
		return nil, errors.New("log: invalid source for console: " + source)
	}
}

// Write implements Writer Write method
func (w *console) Write(level Level, data []byte, _ int) error {
	_, err := w.w.Write(data)
	return err
}

// Close implements Writer Close method
func (w *console) Close() error { return nil }

// File contains the basic writable file operations for logging
type File interface {
	io.WriteCloser
	// Sync commits the current contents of the file to stable storage.
	// Typically, this means flushing the file system's in-memory copy
	// of recently written data to disk.
	Sync() error
}

// FS wraps the basic fs operations for logging
type FS interface {
	OpenFile(name string, flag int, perm os.FileMode) (File, error) // OpenFile opens the file
	Remove(name string) error                                       // Remove removes the file
	Symlink(oldname, newname string) error                          // Symlink creates file symlink
	MkdirAll(path string, perm os.FileMode) error                   // MkdirAll creates a directory
}

// stdFS wraps the standard filesystem
type stdFS struct{}

var defaultFS stdFS

// OpenFile implements FS OpenFile method
func (fs stdFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	f, err := os.OpenFile(name, flag, perm)
	return f, err
}

// Remove implements FS Remove method
func (fs stdFS) Remove(name string) error { return os.Remove(name) }

// Symlink implements FS Symlink method
func (fs stdFS) Symlink(oldname, newname string) error { return os.Symlink(oldname, newname) }

// MkdirAll implements FS MkdirAll method
func (fs stdFS) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }

// FileHeader represents header type of file
type FileHeader int

// FileHeader constants
const (
	NoHeader   FileHeader = 0 // no header in file
	HTMLHeader FileHeader = 1 // append html header in file
)

var fileHeaders = map[FileHeader]string{
	HTMLHeader: `<br/><head>
	<meta charset="UTF-8">
	<style>
		@media screen and (min-width: 1000px) {
			.item { width: 950px; padding-top: 6px; padding-bottom: 12px; padding-left: 24px; padding-right: 16px; }
			.metadata { font-size: 18px; }
			.content { font-size: 16px; }
			.datetime { font-size: 14px; }
		}
		@media screen and (max-width: 1000px) {
			.item { width: 95%; padding-top: 4px; padding-bottom: 8px; padding-left: 16px; padding-right: 12px; }
			.metadata { font-size: 14px; }
			.content { font-size: 13px; }
			.datetime { font-size: 12px; }
		}
		.item { max-width: 95%; box-shadow: rgba(60,64,67,.3) 0 1px 2px 0, rgba(60, 64, 67, .15) 0 1px 3px 1px; background: white; border-radius: 4px; margin: 20px auto; }
		.datetime { color: #00000080; display: block; }
		.metadata { color: #df005f; }
		pre {
			white-space: pre-wrap;       /* Since CSS 2.1 */
			white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
			white-space: -pre-wrap;      /* Opera 4-6 */
			white-space: -o-pre-wrap;    /* Opera 7 */
			word-wrap: break-word;       /* Internet Explorer 5.5+ */
		}
	</style>
</head>`,
}

// FileOptions represents options of file writer
//
// fullname of log file: $Filename.$date[.$rotateId]$Suffix
type FileOptions struct {
	Dir      string     `json:"dir"`      // log directory (default: .)
	Filename string     `json:"filename"` // log filename (default: <process name>)
	Symdir   string     `json:"symdir"`   // symlinked directory (default: "")
	Rotate   bool       `json:"rotate"`   // enable log rotate (default: false)
	MaxSize  int64      `json:"maxsize"`  // max number bytes of log file (default: 64M)
	Suffix   string     `json:"suffix"`   // filename suffix (default: .log)
	Header   FileHeader `json:"header"`   // header type of file (default: NoHeader)

	FS FS `json:"-"` // custom filesystem (default: stdFS)
}

func (opt *FileOptions) setDefaults() {
	if opt.Dir == "" {
		opt.Dir = "."
	}
	if opt.MaxSize == 0 {
		opt.MaxSize = 64 * MB
	}
	if opt.Suffix == "" {
		opt.Suffix = ".log"
	} else if opt.Suffix[0] != '.' {
		opt.Suffix = "." + opt.Suffix
	}
	if opt.Filename == "" {
		name := filepath.Base(os.Args[0])
		if strings.HasSuffix(name, ".exe") {
			name = strings.TrimSuffix(name, ".exe")
		}
		opt.Filename = name
	}
	if opt.FS == nil {
		opt.FS = defaultFS
	}
}

// file is a writer which writes logs to file
type file struct {
	options          FileOptions
	written          int64
	createdAt        time.Time
	rotateId         int
	onceCreateLogDir sync.Once

	mu     sync.Mutex
	writer *bufio.Writer
	file   File
	quit   chan struct{}
}

func newFile(options FileOptions) (*file, error) {
	options.setDefaults()
	w := &file{
		options:  options,
		rotateId: -1,
		quit:     make(chan struct{}),
	}
	if err := w.rotate(time.Now()); err != nil {
		return nil, err
	}
	go func(f *file) {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f.mu.Lock()
				if f.written > 0 {
					f.writer.Flush()
					f.file.Sync()
					f.written = 0
				}
				f.mu.Unlock()
			case <-f.quit:
				return
			}
		}
	}(w)
	return w, nil
}

func parseFileSource(opt *FileOptions, source string) (url.Values, error) {
	i := strings.Index(source, "?")
	if i > 0 {
		opt.Dir, opt.Filename = filepath.Split(source[:i])
	} else {
		opt.Dir, opt.Filename = filepath.Split(source)
	}
	opt.Dir = filepath.Clean(opt.Dir)
	q, err := url.ParseQuery(source[i+1:])
	if err != nil {
		return nil, errors.New("log: invalid source for file: " + source)
	}
	opt.Symdir = q.Get("symdir")
	opt.MaxSize, _ = parseSize(q.Get("maxsize"))
	opt.Rotate, _ = strconv.ParseBool(q.Get("rotate"))
	opt.Suffix = q.Get("suffix")
	header, _ := strconv.Atoi(q.Get("header"))
	opt.Header = FileHeader(header)
	opt.setDefaults()
	return q, nil
}

// source format: path/to/file?k1=v1&...&kn=vn
func openFile(source string) (Writer, error) {
	var opt FileOptions
	_, err := parseFileSource(&opt, source)
	if err != nil {
		return nil, err
	}
	return newFile(opt)
}

// Write writes log to file
func (w *file) Write(level Level, data []byte, _ int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return errNilWriter
	}
	now := time.Now()
	if !isSameDay(now, w.createdAt) {
		if err := w.rotate(now); err != nil {
			return err
		}
	}
	n, err := w.writer.Write(data)
	w.written += int64(n)
	if w.written >= w.options.MaxSize {
		w.rotate(now)
	}
	return err
}

func (w *file) clear() error {
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return err
		}
		if err := w.file.Sync(); err != nil {
			return err
		}
		if err := w.file.Close(); err != nil {
			return err
		}
	}
	w.written = 0
	return nil
}

// Close closes current log file
func (w *file) Close() error {
	close(w.quit)
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.clear()
}

func (w *file) rotate(now time.Time) error {
	w.clear()
	if isSameDay(now, w.createdAt) {
		w.rotateId = (w.rotateId + 1) % 1000
	} else {
		w.rotateId = 0
	}
	w.createdAt = now

	var err error
	w.file, err = w.create()
	if err != nil {
		return err
	}

	w.writer = bufio.NewWriterSize(w.file, 1<<14) // 16k
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "File opened at: %s.\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Built with %s %s for %s/%s.\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	if header, ok := fileHeaders[w.options.Header]; ok {
		fmt.Fprintln(&buf, header)
	}
	n, err := w.file.Write(buf.Bytes())
	w.written += int64(n)
	w.writer.Flush()
	w.file.Sync()
	return err
}

func (w *file) create() (File, error) {
	w.onceCreateLogDir.Do(w.createDir)

	// make filename
	var (
		prefix = w.options.Filename
		date   = w.createdAt.Format("20060102")
		name   = fmt.Sprintf("%s.%s", prefix, date)
	)
	if w.rotateId > 0 {
		name = fmt.Sprintf("%s.%03d", name, w.rotateId)
	}
	name += w.options.Suffix

	// create file
	var (
		fullname = filepath.Join(w.options.Dir, name)
		f        File
		err      error
	)
	if w.options.Symdir != "" {
		fullname = filepath.Join(w.options.Dir, w.options.Symdir, name)
	}
	if w.options.Rotate {
		f, err = w.options.FS.OpenFile(fullname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		f, err = w.options.FS.OpenFile(fullname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	}
	if err == nil && w.options.Symdir != "" {
		symlink := filepath.Join(w.options.Dir, w.options.Filename+w.options.Suffix)
		w.options.FS.Remove(symlink)
		w.options.FS.Symlink(filepath.Join(w.options.Symdir, name), symlink)
	}
	return f, err
}

func (w *file) createDir() {
	dir := w.options.Dir
	if w.options.Symdir != "" {
		dir = filepath.Join(w.options.Dir, w.options.Symdir)
	}
	w.options.FS.MkdirAll(dir, 0755)
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// MultiFileOptions represents options for multi file writer
type MultiFileOptions struct {
	FileOptions
	FatalDir string `json:"fataldir"` // fatal subdirectory (default: fatal)
	ErrorDir string `json:"errordir"` // error subdirectory (default: error)
	WarnDir  string `json:"warndir"`  // warn subdirectory (default: warn)
	InfoDir  string `json:"infodir"`  // info subdirectory (default: info)
	DebugDir string `json:"debugdir"` // debug subdirectory (default: debug)
	TraceDir string `json:"tracedir"` // trace subdirectory (default: trace)
}

func (opt *MultiFileOptions) setDefaults() {
	opt.FileOptions.setDefaults()
	if opt.FatalDir == "" {
		opt.FatalDir = "fatal"
	}
	if opt.ErrorDir == "" {
		opt.ErrorDir = "error"
	}
	if opt.WarnDir == "" {
		opt.WarnDir = "warn"
	}
	if opt.InfoDir == "" {
		opt.InfoDir = "info"
	}
	if opt.DebugDir == "" {
		opt.DebugDir = "debug"
	}
	if opt.TraceDir == "" {
		opt.TraceDir = "trace"
	}
}

type multiFile struct {
	options MultiFileOptions
	files   [numLevel]*file
	group   map[string][]Level
}

func absPath(path string) string {
	s, _ := filepath.Abs(path)
	return s
}

func newMultiFile(options MultiFileOptions) *multiFile {
	options.setDefaults()
	w := new(multiFile)
	w.options = options
	w.group = map[string][]Level{}
	for level := LevelFatal; level <= LevelTrace; level++ {
		dir := w.levelDir(level)
		if levels, ok := w.group[dir]; ok {
			w.group[dir] = append(levels, level)
		} else {
			w.group[dir] = []Level{level}
		}
	}
	return w
}

// source format: path/to/file?k1=v1&...&kn=vn
func openMultiFile(source string) (Writer, error) {
	var opt MultiFileOptions
	q, err := parseFileSource(&opt.FileOptions, source)
	if err != nil {
		return nil, err
	}
	opt.FatalDir = q.Get("fataldir")
	opt.ErrorDir = q.Get("errordir")
	opt.WarnDir = q.Get("warndir")
	opt.InfoDir = q.Get("infodir")
	opt.DebugDir = q.Get("debugdir")
	opt.TraceDir = q.Get("tracedir")
	return newMultiFile(opt), nil
}

func (w *multiFile) Write(level Level, data []byte, headerLen int) error {
	if w.files[level.index()] == nil {
		if err := w.initForLevel(level); err != nil {
			return err
		}
	}
	return w.files[level.index()].Write(level, data, headerLen)
}

func (w *multiFile) Close() error {
	var lastErr error
	for i := range w.files {
		if w.files[i] != nil {
			if err := w.files[i].Close(); err != nil {
				lastErr = err
			}
			w.files[i] = nil
		}
	}
	return lastErr
}

func (w *multiFile) initForLevel(level Level) error {
	index := level.index()
	if index < 0 || index >= len(w.files) {
		return errUnrecognizedLevel
	}
	f, err := newFile(w.optionsOfLevel(level))
	if err != nil {
		return err
	}
	w.files[index] = f
	if levels, ok := w.group[absPath(f.options.Dir)]; ok {
		for _, lv := range levels {
			if w.files[lv.index()] == nil {
				w.files[lv.index()] = f
			}
		}
	}
	return nil
}

func (w *multiFile) optionsOfLevel(level Level) FileOptions {
	options := w.options.FileOptions
	options.Dir = filepath.Join(options.Dir, w.levelDir(level))
	return options
}

func (w *multiFile) levelDir(level Level) string {
	switch level {
	case LevelFatal:
		return w.options.FatalDir
	case LevelError:
		return w.options.ErrorDir
	case LevelWarn:
		return w.options.WarnDir
	case LevelInfo:
		return w.options.InfoDir
	case LevelDebug:
		return w.options.DebugDir
	default:
		return w.options.TraceDir
	}
}
