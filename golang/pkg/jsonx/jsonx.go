package jsonx

import "github.com/bytedance/sonic"

func JSON(v any) []byte {
	data, _ := sonic.Marshal(v)
	return data
}

func JSONS(v any) string {
	return string(JSON(v))
}

func JSONE(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

func Pretty(v any) string {
	data, _ := sonic.MarshalIndent(v, "", " ")
	return string(data)
}

type Lz[T bool] struct {
	v      any
	pretty bool
}

func (lz Lz[bool]) String() string {
	if lz.pretty {
		return Pretty(lz.v)
	}
	return JSONS(lz.v)
}

func LzJSON(v any) Lz[bool] {
	return Lz[bool]{v: v, pretty: false}
}

func LzPretty(v any) Lz[bool] {
	return Lz[bool]{v: v, pretty: true}
}
