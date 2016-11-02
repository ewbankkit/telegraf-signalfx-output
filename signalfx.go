package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"time"

	"github.com/influxdata/toml"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"
)

type signalFx struct {
	Config struct {
		AuthToken string `toml:"auth_token"`
		UserAgent string `toml:"user_agent"`
		Endpoint  string `toml:"endpoint"`
	} `toml:"signalfx"`

	sink *sfxclient.HTTPDatapointSink
}

var (
	invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
	envVarRE          = regexp.MustCompile(`\$\w+`)
)

// Lifted from internal/config/config.go:parseFile.
func (s *signalFx) loadConfig(path string) error {
	if path == "" {
		return errors.New("No configuration file specified")
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// ugh windows why
	contents = bytes.TrimPrefix(contents, []byte("\xef\xbb\xbf"))

	envVars := envVarRE.FindAll(contents, -1)
	for _, envVar := range envVars {
		envVal := os.Getenv(strings.TrimPrefix(string(envVar), "$"))
		if envVal != "" {
			contents = bytes.Replace(contents, envVar, []byte(envVal), 1)
		}
	}

	if err := toml.Unmarshal(contents, s); err != nil {
		return err
	}

	return nil
}

func (s *signalFx) connect() error {
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

func (s *signalFx) close() error {
	return nil
}

func (s *signalFx) write(metrics []*metric) error {
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
