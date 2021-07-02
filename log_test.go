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
	log.Start(log.WithWriters(writer), log.WithLevel(log.LevelTrace), log.WithPrefix("testing"))
	log.Printf(log.LevelTrace, "hello %s", "log")
	logger := log.Prefix("prefix")
	logger.Debug().Print("hello world")
	log.Shutdown()
	got := writer.buf.String()
	want := "[TRACE] (testing) hello log\n[DEBUG] (testing/prefix) hello world\n"
	if got != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}

func ExampleRecorder() {
	writer := new(testingLogWriter)
	log.Start(log.WithWriters(writer), log.WithLevel(log.LevelInfo), log.WithPrefix("testing"))
	log.Info().Int("int", 123456).Print("recorder")
	log.Info().Int8("int8", -12).Print("recorder")
	log.Info().Int16("int16", 1234).Print("recorder")
	log.Info().Int32("int32", -12345678).Print("recorder")
	log.Info().Int64("int64", 1234567890).Print("recorder")
	log.Info().Uint("uint", 123456).Print("recorder")
	log.Info().Uint8("uint8", 120).Print("recorder")
	log.Info().Uint16("uint16", 12340).Print("recorder")
	log.Info().Uint32("uint32", 123456780).Print("recorder")
	log.Info().Uint64("uint64", 12345678900).Print("recorder")
	log.Info().Float32("float32", 1234.5678).Print("recorder")
	log.Info().Float64("float64", 0.123456789).Print("recorder")
	log.Info().Complex64("complex64", 1+2i).Print("recorder")
	log.Info().Complex128("complex128", 1).Print("recorder")
	log.Info().Complex128("complex128", 2i).Print("recorder")
	log.Info().Byte("byte", 'h').Print("recorder")
	log.Info().Rune("rune", 'Å').Print("recorder")
	log.Info().Bool("bool", true).Print("recorder")
	log.Info().Bool("bool", false).Print("recorder")
	log.Info().String("string", "hello").Print("recorder")
	log.Info().Error("error", nil).Print("recorder")
	log.Info().Error("error", errors.New("err")).Print("recorder")
	log.Info().Any("any", nil).Print("recorder")
	log.Info().Any("any", "nil").Print("recorder")
	log.Info().Any("any", struct {
		x int
		y string
	}{1, "hello"}).Print("recorder")
	log.Info().Type("type", nil).Print("recorder")
	log.Info().Type("type", "string").Print("recorder")
	log.Info().Type("type", new(int)).Print("recorder")
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
	log.Info().Date("date", t).Print("recorder")
	log.Info().Time("time", t).Print("recorder")
	log.Info().Duration("duration", time.Millisecond*1200).Print("recorder")
	log.Info().String("$name", "hello").Print("recorder")
	log.Info().String("name of", "hello").Print("recorder")
	log.Info().Int32s("int32s", []int32{1, 3, 5}).Print("recorder")
	log.Info().Strings("strings", []string{"x", "x y", "z"}).Print("recorder")
	log.Info().Bytes("bytes", []byte{'1', '3', 'x'}).Print("recorder")
	log.Prefix("prefix").Info().
		String("k1", "v1").
		Int("k2", 2).
		Print("prefix logging")
	log.Debug().String("key", "value").Print("not output")
	log.Shutdown()
	fmt.Print(writer.buf.String())
	// Output:
	// [INFO] (testing) {int:123456} recorder
	// [INFO] (testing) {int8:-12} recorder
	// [INFO] (testing) {int16:1234} recorder
	// [INFO] (testing) {int32:-12345678} recorder
	// [INFO] (testing) {int64:1234567890} recorder
	// [INFO] (testing) {uint:123456} recorder
	// [INFO] (testing) {uint8:120} recorder
	// [INFO] (testing) {uint16:12340} recorder
	// [INFO] (testing) {uint32:123456780} recorder
	// [INFO] (testing) {uint64:12345678900} recorder
	// [INFO] (testing) {float32:1234.5677} recorder
	// [INFO] (testing) {float64:0.123456789} recorder
	// [INFO] (testing) {complex64:1+2i} recorder
	// [INFO] (testing) {complex128:1} recorder
	// [INFO] (testing) {complex128:2i} recorder
	// [INFO] (testing) {byte:'h'} recorder
	// [INFO] (testing) {rune:'Å'} recorder
	// [INFO] (testing) {bool:true} recorder
	// [INFO] (testing) {bool:false} recorder
	// [INFO] (testing) {string:"hello"} recorder
	// [INFO] (testing) {error:nil} recorder
	// [INFO] (testing) {error:"err"} recorder
	// [INFO] (testing) {any:nil} recorder
	// [INFO] (testing) {any:"nil"} recorder
	// [INFO] (testing) {any:"{1 hello}"} recorder
	// [INFO] (testing) {type:"nil"} recorder
	// [INFO] (testing) {type:"string"} recorder
	// [INFO] (testing) {type:"*int"} recorder
	// [INFO] (testing) {date:"2020-05-01+08:00"} recorder
	// [INFO] (testing) {time:"2020-05-01T12:20:30.123456789+08:00"} recorder
	// [INFO] (testing) {duration:1.2s} recorder
	// [INFO] (testing) {$name:"hello"} recorder
	// [INFO] (testing) {"name of":"hello"} recorder
	// [INFO] (testing) {int32s:[1,3,5]} recorder
	// [INFO] (testing) {strings:["x","x y","z"]} recorder
	// [INFO] (testing) {bytes:0x313378} recorder
	// [INFO] (testing/prefix) {k1:"v1",k2:2} prefix logging
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

func benchmarkRecorder(b *testing.B, caller, on bool) {
	benchmarkSetup(b, caller, on)
	for i := 0; i < b.N; i++ {
		log.Debug().
			Int("int", 123456).
			Uint("uint", 123456).
			Float64("float64", 0.123456789).
			String("string", "hello").
			Duration("duration", time.Microsecond*1234567890).
			Print("benchmark recorder")
	}
	benchmarkTeardown(b)
}

func BenchmarkWithCaller(b *testing.B)    { benchmarkRecorder(b, true, false) }
func BenchmarkWithoutCaller(b *testing.B) { benchmarkRecorder(b, false, false) }
func BenchmarkOff(b *testing.B)           { benchmarkRecorder(b, true, true) }

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
