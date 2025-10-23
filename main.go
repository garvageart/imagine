package main

import (
	"runtime"
	"strings"
)

func main() {
	pc, file, line, ok := runtime.Caller(0)
	if !ok {
		panic("could not get caller info")
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		panic("could not get function info")
	}

	println("Caller info:")
	println("File:", file)
	println("Line:", line)
	println("Function:", fn.Name())
	println("Package:", strings.Split(fn.Name(), ".")[0])
}
