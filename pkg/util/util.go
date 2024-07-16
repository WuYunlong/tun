package util

import (
	"crypto/rand"
	"fmt"
)

func RandID() (id string, err error) {
	return RandIDWithLen(16)
}

func RandIDWithLen(idLen int) (id string, err error) {
	if idLen <= 0 {
		return "", nil
	}
	b := make([]byte, idLen/2+1)
	_, err = rand.Read(b)
	if err != nil {
		return
	}

	id = fmt.Sprintf("%x", b)
	return id[:idLen], nil
}

func EmptyOr[T comparable](v T, fallback T) T {
	var zero T
	if zero == v {
		return fallback
	}
	return v
}
