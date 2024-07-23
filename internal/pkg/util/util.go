package util

import "fmt"

func GenerateResponseErrorString(summary string, err error, detailed bool) string {
	if detailed {
		return err.Error()
	}
	return summary
}

var errorMap = map[string]string{
	"":              "未知错误",
	"invalid_token": "无效的 token",
}

func GenerateErrorString(err error) string {
	str := err.Error()
	if v, ok := errorMap[str]; ok {
		str = v
	}
	return str
}

func PanicToError(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic error: %v", r)
		}
	}()

	fn()
	return
}
