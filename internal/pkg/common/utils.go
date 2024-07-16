package common

import "os"

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
