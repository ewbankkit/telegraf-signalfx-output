package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

// https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#json
type Metric struct {
	Fields    map[string]interface{}
	Tags      map[string]string
	Name      string
	Timestamp int64
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

func (m *Metric) unmarshal(line string) error {
	err := json.Unmarshal([]byte(line), m)
	if err != nil {
		return err
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

	var s SignalFx
	err := s.loadConfig(*fConfig)
	if err != nil {
		log.Fatal(err)
	}
	err = s.connect()
	if err != nil {
		log.Fatal(err)
	}
	defer s.close()

	var metrics []*Metric
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var m Metric
		err = m.unmarshal(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
		metrics = append(metrics, &m)
	}

	err = s.write(metrics)
	if err != nil {
		log.Fatal(err)
	}
}
