package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"golang.org/x/net/context"

	"time"

	"github.com/influxdata/toml"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"
)

type SignalFx struct {
	Config struct {
		AuthToken string `toml:"auth_token"`
		UserAgent string `toml:"user_agent"`
		Endpoint  string `toml:"endpoint"`
	} `toml:"signalfx"`

	sink *sfxclient.HTTPDatapointSink
}

var (
	version string
	commit  string
	branch  string
)

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

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

func (s *SignalFx) loadConfig(path string) error {
	if path == "" {
		return errors.New("No configuration file specified")
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(buf, s); err != nil {
		return err
	}

	return nil
}

func (s *SignalFx) connect() error {
	s.sink = sfxclient.NewHTTPDatapointSink()
	s.sink.AuthToken = s.Config.AuthToken
	if len(s.Config.UserAgent) > 0 {
		s.sink.UserAgent = s.Config.UserAgent
	}
	if len(s.Config.Endpoint) > 0 {
		s.sink.Endpoint = s.Config.Endpoint
	}

	return nil
}

func (s *SignalFx) close() error {
	return nil
}

func (s *SignalFx) write(metrics []*Metric) error {
	var datapoints []*datapoint.Datapoint
	for _, metric := range metrics {
		// Sanitize metric name.
		metricName := metric.Name
		metricName = invalidNameCharRE.ReplaceAllString(metricName, "_")

		// Get a type if it's available, defaulting to Gauge.
		sfMetricType := datapoint.Gauge

		// One SignalFx metric per field.
		for fieldName, fieldValue := range metric.Fields {
			var sfValue datapoint.Value
			switch fieldValue.(type) {
			case float64:
				sfValue = datapoint.NewFloatValue(fieldValue.(float64))
			case int64:
				sfValue = datapoint.NewIntValue(fieldValue.(int64))
			default:
				log.Printf("Unhandled type %T for field %s\n", fieldValue, fieldName)
				continue
			}

			// Sanitize field name.
			fieldName = invalidNameCharRE.ReplaceAllString(fieldName, "_")

			var sfMetricName string
			if fieldName == "value" {
				sfMetricName = metricName
			} else {
				sfMetricName = fmt.Sprintf("%s.%s", metricName, fieldName)
			}

			timestamp := time.Unix(metric.Timestamp, 0)
			datapoint := datapoint.New(sfMetricName, metric.Tags, sfValue, sfMetricType, timestamp)
			datapoints = append(datapoints, datapoint)
		}
	}

	ctx := context.Background()
	err := s.sink.AddDatapoints(ctx, datapoints)
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

// https://github.com/influxdata/telegraf/blob/master/plugins/serializers/json/json.go
// https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#json
type Metric struct {
	Fields    map[string]interface{}
	Tags      map[string]string
	Name      string
	Timestamp int64
}

func (m *Metric) unmarshal(line string) error {
	err := json.Unmarshal([]byte(line), m)
	if err != nil {
		return err
	}
	return nil
}
