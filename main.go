package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version string
	commit  string
	branch  string
)

const usage = `Telegraf SignalFx Output plugin.

Usage:

  telegraf-signalfx-output [commands|flags]

The commands & flags are:
  version            print the version to stdout

  --config <file>     configuration file to load
`

func usageExit(rc int) {
	fmt.Println(usage)
	os.Exit(rc)
}

func main() {
	flag.Usage = func() { usageExit(0) }
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "version":
			fmt.Printf("Telegraf SignalFx Output v%s (git: %s %s)\n", version, branch, commit)
			return
		}
	}
}
