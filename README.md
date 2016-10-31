# Telegraf SignalFx Output

[SignalFx](https://signalfx.com/) external output plugin for [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/).

This plugin is run as a standalone application called by the [Telegraf **exec** output plugin](https://github.com/influxdata/telegraf/issues/1717).  
This plugin is experimental as there is no official Telegraf exec output plugin (yet).  
Telegraf metrics serialized in the [JSON data format](https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md#json) are read from stdin and written to SignalFx over HTTP.  
For each Telegraf metric a SignalFx gauge datapoint is written per field.  
The datapoint's metric name is a concatention of the Telegraf metric name and the field name.  
Tags are written as datapoint dimensions.

## Configuration:

```toml
# Send Telegraf metrics to SignalFx
[signalfx]
  ## Your organization's SignalFx API access token.
  auth_token = "SuperSecretToken"

  ## Optional HTTP User Agent value; Overrides the default.
  # user_agent = "Telegraf collector"

  ## Optional SignalFX API endpoint value; Overrides the default.
  # endpoint = "https://ingest.signalfx.com/v2/datapoint"
```

### Required parameters:

* `auth_token`: Your organization's SignalFx API access token.


### Optional parameters:

* `user_agent`: HTTP User Agent.
* `endpoint`: SignalFX API endpoint.

## Build:
Dependencies are managed via [gdm](https://github.com/sparrc/gdm),
which gets installed via the Makefile if you don't have it already.  
You also must build with golang version 1.7+.

1. [Install Go](https://golang.org/doc/install)
2. [Setup your GOPATH](https://golang.org/doc/code.html#GOPATH)
3. Run `go get github.com/ewbankkit/telegraf-signalfx-output`
4. Run `cd $GOPATH/src/github.com/ewbankkit/telegraf-signalfx-output`
5. Run `make`
