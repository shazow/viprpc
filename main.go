package main

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

// Version of the binary, assigned during build.
var Version = "dev"

// Options contains the flag options
type Options struct {
	Version bool `long:"version" description:"Print version and exit."`
}

func exit(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

func main() {
	options := Options{}
	p, err := flags.NewParser(&options, flags.Default).ParseArgs(os.Args[1:])
	if err != nil {
		if p == nil {
			fmt.Println(err)
		}
		return
	}

	if options.Version {
		fmt.Println(Version)
		os.Exit(0)
	}

	// TODO: ...
}
