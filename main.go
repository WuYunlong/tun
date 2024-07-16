package main

import "tun/internal/pkg/file"

func main() {
	client := file.NewClient("")
	file.GetDB().NewClient(client)
}
