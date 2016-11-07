package main

import (
	"fmt"
	"os"
)

func main() {
	baseImage := os.Args[len(os.Args)-2]

	if baseImage == "fail-this" {
		fmt.Println("fake grootfs failed")
		os.Exit(1)
	}

	fmt.Println("/var/lib/btrfs/image")
}
