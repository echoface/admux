package utils

import "fmt"

func Ignore(...any) {}

func IgnoreErr(err error, format string, vs ...any) {
	if err == nil {
		return
	}
	// TODO: log here
}

func PanicIf(cond bool, format string, vs ...any) {
	if !cond {
		return
	}
	panic(fmt.Errorf(format, vs...))
}

func PanicIfErr(err error, format string, vs ...any) {
	PanicIf(err != nil, format, vs...)
}
