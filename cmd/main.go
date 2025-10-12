package main

import (
	"flag"
	"fmt"
)

func main() {
	// Define a string flag called "name"
	name := flag.String("name", "World", "name to greet")

	// Parse the flags passed in at runtime
	flag.Parse()

	// Use the flag value
	fmt.Printf("Hello, %s!\n", *name)
}
