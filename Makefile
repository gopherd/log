all: genslice genlintfuncs loglint

genslice:
	go run genslice.go > slice.go

genlintfuncs:
	go run genlintfuncs.go > cmd/loglint/funcs.go

loglint:
	cd cmd/loglint && go install
