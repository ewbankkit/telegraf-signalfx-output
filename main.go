package main

import (
	"flag"
	"fmt"
	"os"

	"errors"

	"log"

	"github.com/signalfx/golib/sfxclient"
)

type SignalFx struct {
	AuthToken string `toml:"auth_token"`
	UserAgent string `toml:"user_agent"`
	Endpoint  string `toml:"endpoint"`

	sink *sfxclient.HTTPDatapointSink
}

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

func loadConfig(path string) error {
	if path == "" {
		return errors.New("No configuration file specified")
	}
	return nil
}

func main() {
	flag.Usage = func() { usageExit(0) }
	fConfig := flag.String("config", "", "configuration file to load")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "version":
			fmt.Printf("Telegraf SignalFx Output v%s (git: %s %s)\n", version, branch, commit)
			return
		}
	}

	err := loadConfig(*fConfig)
	if err != nil {
		log.Fatal(err)
	}
}
