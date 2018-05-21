package main

import (
	"provider"
	"fmt"
	"os"
	"strings"
	"consumer"
)

func main() {
	fmt.Print("Args: ")
	fmt.Println(os.Args)
	opts := make(map[string]string)
	for _, arg := range os.Args[1:] {
		i := strings.IndexByte(arg, '=')
		if i >= 0 {
			opts[arg[1:i]] = arg[i+1:]
		}
	}
	switch opts["type"] {
	case "provider":
		provider.Start(opts)
	case "consumer":
		consumer.Start(opts)
	default:
		fmt.Println("Invalid options, exiting.")
	}
}
