package util

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
