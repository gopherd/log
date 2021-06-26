all: genslice genlintfuncs loglint benchmark

genslice:
	go run genslice.go > slice.go

genlintfuncs:
	./genlintfuncs.sh > cmd/loglint/funcs.go

loglint:
	cd cmd/loglint && go install

test:
	go test

benchmark:
	go test -bench=. -benchmem
