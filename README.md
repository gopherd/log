LOG
===

`log` is a structured logger library. We have following key features:

+ lightweight: no any dependencies.
+ fast: formmating log message very fast.
+ zero-memory: zero memory allocation while formmating log message.
+ customizable: customize `Writer` and `Printer`.

Getting started
---------------

Install:

```
go get github.com/gopherd/log
```

```go
package main

import (
	"errors"
	"time"

	"github.com/gopherd/log"
)

func main() {
	// Start and defer Shutdown
	if err := log.Start(log.WithConsole()); err != nil {
		panic(err)
	}
	defer log.Shutdown()

	// Default log level is log.LvINFO, you can change the level as following:
	//
	//	log.SetLevel(log.LvTRACE)
	// 	log.SetLevel(log.LvDEBUG)
	// 	log.SetLevel(log.LvINFO)
	// 	log.SetLevel(log.LvWARN)
	// 	log.SetLevel(log.LvERROR)
	// 	log.SetLevel(log.LvFATAL)
	log.SetLevel(log.LvTRACE)

	log.Trace().Print("verbose message")
	log.Debug().
		Int("id", 123).
		String("name", "gopherd").
		Print("debug message")

	log.Info().
		Int32("i32", -12).
		Print("important message")
	log.Warn().
		Duration("duration", time.Second).
		Print("warning: cost to much time")
	log.Error().
		Error("error", errors.New("EOF")).
		Print("something is wrong")

	// Set header flags, all supported flags: Ldatetime, Llongfile, Lshortfile, LUTC
	log.SetFlags(log.Ldatetime | log.Lshortfile | log.LUTC)

	log.Fatal().Print("should be printed and exit program with status code 1")
	log.Info().Print("You cannot see me")
}
```

## Linter: [loglint](https://github.com/gopherd/log/tree/main/cmd/loglint)

install loglint:

```
go install github.com/gopherd/log/cmd/loglint
```

loglint used to check unfinished chain calls, e.g.

```go
// wrong
log.Debug()                              // warning: result of github.com/gopherd/log.Debug call not used
log.Prefix("pre").Trace()                // warning: result of (github.com/gopherd/log.Prefix).Trace call not used
log.Debug().String("k", "v")             // warning: result of (*github.com/gopherd/log.Fields).String call not used
log.Debug().Int("i", 1).String("k", "v") // warning: result of (*github.com/gopherd/log.Fields).String call not used

// right
log.Debug().String("k", "v").Print("message")
log.Debug().Print("message")
```
