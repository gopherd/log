// +build gopherd_log_genlintfuncs

// command to generate cmd/loglint/funcs.go
//
//	go run genlintfuncs.go > funcs.go
//
package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"strings"
	"unicode"
)

//go:embed *.go
var dir embed.FS

func main() {
	entries, err := dir.ReadDir(".")
	if err != nil {
		panic(err)
	}
	const (
		prefix = "//loglint:"
		method = "method"
	)
	p("// Auto-generated by genlintfuncs.go, DON'T EDIT IT!")
	p("package main")
	var (
		methods []string
		funcs   []string
	)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "gen") {
			continue
		}
		content, err := dir.ReadFile(e.Name())
		if err != nil {
			panic(e.Name() + ": " + err.Error())
		}
		scanner := bufio.NewScanner(bytes.NewReader(content))
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			lint := strings.TrimPrefix(line, prefix)
			var isMethod bool
			if strings.HasPrefix(lint, method) {
				isMethod = true
				lint = strings.TrimPrefix(lint, method)
			}
			if len(lint) == 0 || lint[0] != ' ' || !isFuncname(strings.TrimSpace(lint[1:]), isMethod) {
				panic("invalid loglint comment: " + line)
			}
			name := strings.TrimSpace(lint[1:])
			if isMethod {
				methods = append(methods, name)
			} else {
				funcs = append(funcs, name)
			}
		}
	}
	p()
	if len(funcs) > 0 {
		p("var allFuncs = []string{")
		for _, name := range funcs {
			p("\t", `"`, name, `",`)
		}
		p("}")
	} else {
		p("var allFuncs = []string{}")
	}
	p()
	if len(methods) > 0 {
		p("var allMethods = []string{")
		for _, name := range methods {
			p("\t", `"`, name, `",`)
		}
		p("}")
	} else {
		p("var allMethods = []string{}")
	}
}

func p(a ...interface{}) {
	for i := range a {
		fmt.Print(a[i])
	}
	fmt.Println("")
}

func isFuncname(s string, isMethod bool) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		isDigit := unicode.IsDigit(c)
		if isDigit && i == 0 {
			return false
		}
		if !isDigit && c != '_' && !unicode.IsLetter(c) && !(isMethod && c == '.') {
			return false
		}
	}
	return true
}