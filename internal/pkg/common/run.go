package common

import (
	"runtime"
)

func IsWin() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}
