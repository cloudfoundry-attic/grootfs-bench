package main

import (
	"fmt"
	"os"
)

func main() {
	image := os.Args[len(os.Args)-2]

	if image == "fail-this" {
		fmt.Println("fake grootfs failed")
		os.Exit(1)
	}

	fmt.Println("/var/lib/btrfs/bundle")
}
