package log_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gopherd/log"
)

type testingLogWriter struct {
	discard bool
	buf     bytes.Buffer
}

func (w *testingLogWriter) Write(level log.Level, data []byte, headerLen int) error {
	if !w.discard {
		w.buf.WriteByte('[')
		w.buf.WriteString(level.String())
		w.buf.WriteByte(']')
		w.buf.WriteByte(' ')
		w.buf.Write(data[headerLen:])
	}
	return nil
}

func (w *testingLogWriter) Close() error { return nil }

func TestWriter(t *testing.T) {
	writer := new(testingLogWriter)
	log.Start(log.WithWriters(writer), log.WithLevel(log.LevelTrace))
	log.Log(log.LevelTrace).Print("hello log")
	log.Debug().Print("hello world")
	log.Shutdown()
	got := writer.buf.String()
	want := "[TRACE] hello log\n[DEBUG] hello world\n"
	if got != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}

func ExampleLog() {
	writer := new(testingLogWriter)
	logger := log.NewLogger("testing")
	logger.Start(log.WithWriters(writer), log.WithLevel(log.LevelInfo))
	logger.Info().Int("int", 123456).Print("ctx")
	logger.Info().Int8("int8", -12).Print("ctx")
	logger.Info().Int16("int16", 1234).Print("ctx")
	logger.Info().Int32("int32", -12345678).Print("ctx")
	logger.Info().Int64("int64", 1234567890).Print("ctx")
	logger.Info().Uint("uint", 123456).Print("ctx")
	logger.Info().Uint8("uint8", 120).Print("ctx")
	logger.Info().Uint16("uint16", 12340).Print("ctx")
	logger.Info().Uint32("uint32", 123456780).Print("ctx")
	logger.Info().Uint64("uint64", 12345678900).Print("ctx")
	logger.Info().Float32("float32", 1234.5678).Print("ctx")
	logger.Info().Float64("float64", 0.123456789).Print("ctx")
	logger.Info().Complex64("complex64", 1+2i).Print("ctx")
	logger.Info().Complex128("complex128", 1).Print("ctx")
	logger.Info().Complex128("complex128", 2i).Print("ctx")
	logger.Info().Byte("byte", 'h').Print("ctx")
	logger.Info().Rune("rune", 'Å').Print("ctx")
	logger.Info().Bool("bool", true).Print("ctx")
	logger.Info().Bool("bool", false).Print("ctx")
	logger.Info().String("string", "hello").Print("ctx")
	logger.Info().Error("error", nil).Print("ctx")
	logger.Info().Error("error", errors.New("err")).Print("ctx")
	logger.Info().Any("any", nil).Print("ctx")
	logger.Info().Any("any", "nil").Print("ctx")
	logger.Info().Any("any", struct {
		x int
		y string
	}{1, "hello"}).Print("ctx")
	logger.Info().Type("type", nil).Print("ctx")
	logger.Info().Type("type", "string").Print("ctx")
	logger.Info().Type("type", new(int)).Print("ctx")
	const (
		year  = 2020
		month = time.May
		day   = 1
		hour  = 12
		min   = 20
		sec   = 30
		nsec  = 123456789
	)
	t := time.Date(year, month, day, hour, min, sec, nsec, time.Local)
	logger.Info().Date("date", t).Print("ctx")
	logger.Info().Time("time", t).Print("ctx")
	logger.Info().Duration("duration", time.Millisecond*1200).Print("ctx")
	logger.Info().String("$name", "hello").Print("ctx")
	logger.Info().String("name of", "hello").Print("ctx")
	logger.Info().Int32s("int32s", []int32{1, 3, 5}).Print("ctx")
	logger.Info().Strings("strings", []string{"x", "x y", "z"}).Print("ctx")
	logger.Info().Bytes("bytes", []byte{'1', '3', 'x'}).Print("ctx")
	logger.Debug().String("key", "value").Print("not output")
	logger.If(true).Info().String("key", "value").Print("should be printed")
	logger.If(false).Info().String("key", "value").Print("should not be printed")
	logger.Shutdown()
	fmt.Print(writer.buf.String())
	// Output:
	// [INFO] (testing) {int:123456} ctx
	// [INFO] (testing) {int8:-12} ctx
	// [INFO] (testing) {int16:1234} ctx
	// [INFO] (testing) {int32:-12345678} ctx
	// [INFO] (testing) {int64:1234567890} ctx
	// [INFO] (testing) {uint:123456} ctx
	// [INFO] (testing) {uint8:120} ctx
	// [INFO] (testing) {uint16:12340} ctx
	// [INFO] (testing) {uint32:123456780} ctx
	// [INFO] (testing) {uint64:12345678900} ctx
	// [INFO] (testing) {float32:1234.5677} ctx
	// [INFO] (testing) {float64:0.123456789} ctx
	// [INFO] (testing) {complex64:1+2i} ctx
	// [INFO] (testing) {complex128:1} ctx
	// [INFO] (testing) {complex128:2i} ctx
	// [INFO] (testing) {byte:'h'} ctx
	// [INFO] (testing) {rune:'Å'} ctx
	// [INFO] (testing) {bool:true} ctx
	// [INFO] (testing) {bool:false} ctx
	// [INFO] (testing) {string:"hello"} ctx
	// [INFO] (testing) {error:nil} ctx
	// [INFO] (testing) {error:"err"} ctx
	// [INFO] (testing) {any:nil} ctx
	// [INFO] (testing) {any:"nil"} ctx
	// [INFO] (testing) {any:"{1 hello}"} ctx
	// [INFO] (testing) {type:"nil"} ctx
	// [INFO] (testing) {type:"string"} ctx
	// [INFO] (testing) {type:"*int"} ctx
	// [INFO] (testing) {date:"2020-05-01+08:00"} ctx
	// [INFO] (testing) {time:"2020-05-01T12:20:30.123456789+08:00"} ctx
	// [INFO] (testing) {duration:1.2s} ctx
	// [INFO] (testing) {$name:"hello"} ctx
	// [INFO] (testing) {"name of":"hello"} ctx
	// [INFO] (testing) {int32s:[1,3,5]} ctx
	// [INFO] (testing) {strings:["x","x y","z"]} ctx
	// [INFO] (testing) {bytes:0x313378} ctx
	// [INFO] (testing) {key:"value"} should be printed
}

func benchmarkSetup(b *testing.B, caller, off bool) {
	writer := new(testingLogWriter)
	writer.discard = true
	var options = make([]log.Option, 0, 4)
	options = append(options,
		log.WithWriters(writer),
		log.WithSync(true),
	)
	if off {
		options = append(options, log.WithLevel(log.LevelInfo))
	} else {
		options = append(options, log.WithLevel(log.LevelDebug))
	}
	if caller {
		options = append(options, log.WithFlags(log.Lshortfile|log.Ltimestamp|log.LUTC))
	} else {
		options = append(options, log.WithFlags(log.Ltimestamp|log.LUTC))
	}
	log.Start(options...)
	b.ResetTimer()
}

func benchmarkTeardown(b *testing.B) {
	b.StopTimer()
	log.Shutdown()
	b.StartTimer()
}

func benchmarkContext(b *testing.B, caller, on bool) {
	benchmarkSetup(b, caller, on)
	for i := 0; i < b.N; i++ {
		log.Debug().
			Int("int", 123456).
			Uint("uint", 123456).
			Float64("float64", 0.123456789).
			String("string", "hello").
			Duration("duration", time.Microsecond*1234567890).
			Print("benchmark ctx")
	}
	benchmarkTeardown(b)
}

func BenchmarkWithCaller(b *testing.B)    { benchmarkContext(b, true, false) }
func BenchmarkWithoutCaller(b *testing.B) { benchmarkContext(b, false, false) }
func BenchmarkOff(b *testing.B)           { benchmarkContext(b, true, true) }

// testFS implements File interface
type testFile struct {
	content bytes.Buffer
}

func (t *testFile) Write(p []byte) (int, error) { return t.content.Write(p) }
func (t *testFile) Close() error                { return nil }
func (t *testFile) Sync() error                 { return nil }

// testFS implements FS interface
type testFS struct {
	files map[string]*testFile
}

func newTestFS() *testFS {
	return &testFS{
		files: make(map[string]*testFile),
	}
}

// OpenFile implements FS OpenFile method
func (fs testFS) OpenFile(name string, flag int, perm os.FileMode) (log.File, error) {
	f, ok := fs.files[name]
	if ok {
		if flag&os.O_CREATE != 0 && flag&os.O_EXCL == 0 {
			return nil, os.ErrExist
		}
		if flag&os.O_TRUNC != 0 {
			f.content.Reset()
		}
	} else if flag&os.O_CREATE != 0 {
		f = &testFile{}
		fs.files[name] = f
	} else {
		return nil, os.ErrNotExist
	}
	return f, nil
}

// Remove implements FS Remove method
func (fs *testFS) Remove(name string) error {
	if _, ok := fs.files[name]; !ok {
		return os.ErrNotExist
	}
	delete(fs.files, name)
	return nil
}

// Symlink implements FS Symlink method
func (fs testFS) Symlink(oldname, newname string) error { return nil }

// MkdirAll implements FS MkdirAll method
func (fs testFS) MkdirAll(path string, perm os.FileMode) error { return nil }

func TestFile(t *testing.T) {
	// (TODO): test writer `file`
}
