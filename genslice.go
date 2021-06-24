// +build gopherd_log_genslice

// command to generate slice.go
//
//	go run genslice.go > slice.go
//
package main

import (
	"fmt"
	"strings"
)

type scalar struct {
	name string
}

func (s scalar) sliceName() string { return strings.Title(s.name) + "s" }
func (s scalar) isInt() bool       { return strings.HasPrefix(s.name, "int") }
func (s scalar) isUint() bool      { return strings.HasPrefix(s.name, "uint") }
func (s scalar) isFloat() bool     { return strings.HasPrefix(s.name, "float") }
func (s scalar) isComplex() bool   { return strings.HasPrefix(s.name, "complex") }
func (s scalar) bits() int {
	switch s.name {
	case "float32", "complex64":
		return 32
	case "float64", "complex128":
		return 64
	default:
		return 0
	}
}

func t(name string) scalar {
	return scalar{
		name: name,
	}
}

func p(a ...interface{}) {
	for i := range a {
		fmt.Print(a[i])
	}
	fmt.Println("")
}

func main() {
	types := []scalar{
		t("int"), t("int8"), t("int16"), t("int32"), t("int64"),
		t("uint"), t("uint8"), t("uint16"), t("uint32"), t("uint64"),
		t("float32"), t("float64"),
		t("complex64"), t("complex128"),
		t("bool"), t("string"), t("byte"),
	}
	p("// Auto-generated by genslice.go, DON'T EDIT IT!")
	p("package log")
	for _, t := range types {
		p()
		p("//loglint:method ", t.sliceName())
		p("func (fields *Fields) ", t.sliceName(), "(key string, value []", t.name, ") *Fields {")
		p("	if fields != nil {")
		p("		fields.encoder.encodeKey(key)")
		p("		fields.encoder.encode", t.sliceName(), "(value)")
		p("	}")
		p("	return fields")
		p("}")
	}

	for _, t := range types {
		p()
		p("func (b *jsonx) encode", t.sliceName(), "(s []", t.name, ") {")
		if t.name != "byte" {
			p("\tb.writeByte('[')")
		} else {
			p("\tb.writeString(\"0x\")")
		}
		p("\tfor i := range s {")
		if t.name != "byte" {
			p("\t\tif i > 0 {")
			p("\t\t\tb.writeByte(',')")
			p("\t\t}")
		}
		if t.isInt() {
			p("\t\tb.encodeInt(int64(s[i]))")
		} else if t.isUint() {
			p("\t\tb.encodeUint(uint64(s[i]))")
		} else if t.isFloat() {
			p("\t\tb.encodeFloat(float64(s[i]), ", t.bits(), ")")
		} else if t.isComplex() {
			p("\t\tb.encodeComplex(float64(real(s[i])), float64(imag(s[i])), ", t.bits(), ")")
		} else if t.name == "bool" {
			p("\t\tb.encodeBool(s[i])")
		} else if t.name == "string" {
			p("\t\tb.encodeString(s[i])")
		} else if t.name == "byte" {
			p("\t\th, l := s[i]>>4, s[i]&0xF")
			p("\t\tb.writeByte(hex[h])")
			p("\t\tb.writeByte(hex[l])")
		}
		p("\t}")
		if t.name != "byte" {
			p("\tb.writeByte(']')")
		}
		p("}")
	}
}