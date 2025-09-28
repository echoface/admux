package utils

func Ignore(...any) {}

func IgnoreErr(err error, format string, vs ...any) {
	if err == nil {
		return
	}
	// TODO: log here
}
