<div align="center" style="margin-bottom:20px">
  <img src=".assets/banner.png" alt="ulog" />
  <div align="center">
    <a href="https://github.com/blugnu/ulog/actions/workflows/release.yml"><img alt="build-status" src="https://github.com/blugnu/ulog/actions/workflows/pipeline.yml/badge.svg?branch=master&style=flat-square"/></a>
    <a href="https://goreportcard.com/report/github.com/blugnu/ulog" ><img alt="go report" src="https://goreportcard.com/badge/github.com/blugnu/ulog"/></a>
    <a><img alt="go version >= 1.14" src="https://img.shields.io/github/go-mod/go-version/blugnu/ulog?style=flat-square"/></a>
    <a href="https://github.com/blugnu/ulog/blob/master/LICENSE"><img alt="MIT License" src="https://img.shields.io/github/license/blugnu/ulog?color=%234275f5&style=flat-square"/></a>
    <a href="https://coveralls.io/github/blugnu/ulog?branch=master"><img alt="coverage" src="https://img.shields.io/coveralls/github/blugnu/ulog?style=flat-square"/></a>
    <a href="https://pkg.go.dev/github.com/blugnu/ulog"><img alt="docs" src="https://pkg.go.dev/badge/github.com/blugnu/ulog"/></a>
  </div>
</div>

<br>

# blugnu/ulog

A configurable, structured logging library for Go that does not sacrifice (too much) efficiency in exchange for convenience and flexibility.

## Features

- [x] **Highly configurable** - configure your logger to emit logs in the format you want, with the fields you want, at the level you want, sending them to the destination (or destination**s**) you want; the default configuration is designed to be sensible and useful out-of-the-box, but can be easily customised to suit your needs

- [x] **Structured logging** - logs are emitted in a structured format (`logfmt` by default) that can be easily parsed by log aggregators and other tools; a JSON formatter is also provided 

- [x] **Contextual logging** - logs may be enriched with context from the context provided by the caller or from context carried by an error (via support for the [blugnu/errorcontext](https://github.com/blugnu/errorcontext) module); automatic contextual enrichment can be embedded by registering a `ulog.ContextEnricher` with the logger

- [x] **Efficient** - allocations are kept to a minimum and conditional flows eliminated where-ever possible to ensure that the overhead of logging is kept to a minimum

- [x] **Flexible** - support for different logger backends, including a mux to send simultaneously to multiple, independently configured transports

- [x] **Convenient** - a simple, consistent and familiar API that is easy and intuitive to use

- [x] **Testable** - use the provided mock logger to verify that the logs expected by your observability alerts are actually produced!


## Installation

```bash
go get github.com/blugnu/ulog
```

## Tech Stack
blugnu/ulog is built on the following main stack:

- <img width='25' height='25' src='https://img.stackshare.io/service/1005/O6AczwfV_400x400.png' alt='Golang'/> [Golang](http://golang.org/) – Languages
- <img width='25' height='25' src='https://img.stackshare.io/service/11563/actions.png' alt='GitHub Actions'/> [GitHub Actions](https://github.com/features/actions) – Continuous Integration

Full tech stack [here](/techstack.md)

## Usage

A minimal example of using `ulog` to produce a `logfmt` formatted message to `os.Stdout`:

```go
package main

func main() {
    ctx := context.Background()

    // create a new logger
    logger, closelog, err := ulog.NewLogger(ctx)
    if err != nil {
      log.Fatalf("error initialising logger: %s", err)
    }
    defer closelog()

    // log a message
    logger.Info("hello world")
    logger.Error("oops!")
}
```
Will produce output similar to the following:

```bash
time=2023-11-23T12:35:04.789346Z level=INFO  msg="hello world"
time=2023-11-23T12:35:04.789347Z level=ERROR msg="oops!"
```

For more detailed examples, see the [examples](examples) directory.

## Configuration

The default configuration is designed to be sensible and useful out-of-the-box, but can be easily customised to suit your needs.

### Levels

The following log levels are supported:

* `TRACE` - for low-level diagnostic purposes
* `DEBUG` - for debugging purposes
* `INFO` - for informational messages
* `WARN` - for warning messages
* `ERROR` - for error messages
* `FATAL` - for fatal errors

The default log level is `INFO` which may be overridden via the `LoggerLevel` configuration function when configuring a new logger:

```go
logger := ulog.New(ulog.LoggerLevel(ulog.DEBUG))
```

### Formatter

The following formatters are supported:

* `logfmt` - a simple, structured format that is easy for both humans and machines to read and parse
* `json` - a structured format that is easy for machines to parse but can be noisy for humans
* `msgpack` - an efficient structured, binary format for machine use, unintelligible to humans

The default formatter is `logfmt` but may be overridden using the `LoggerFormat` configuration function when configuring a new logger:

```go
logger, closelog, err := ulog.NewLogger(ulog.LoggerFormat(ulog.NewJSONFormatter()))
```

Formatters offer a number of configuration options that can be set via functions supplied to their factory function.  For example, to configure the name used for the `time` field by a `logfmt` formatter:

```go
	log, closelog, _ := ulog.NewLogger(ctx,
		ulog.LoggerFormat(ulog.LogfmtFormatter(
			ulog.LogfmtFieldNames(map[ulog.FieldId]string{
				ulog.TimeField: "timestamp",
			}),
		)),
	)
```

### Output

On a standard logger, the output may be written to any `io.Writer`.  The default output is `os.Stdout`.

The output may be configured  via the `LoggerOutput` configuration function when configuring a new logger:

```go
logger, closefn, err := ulog.NewLogger(ctx, ulog.LoggerOutput(io.Stderr))
```

Logs may be sent to _multiple_ destinations simultaneously using a `ulog.Mux`.  For example, full logs could be sent to an aggregator using `msgpack` format, while `logfmt` formatted logs at error level and above are also sent to `os.Stderr`.

See the [examples](examples) directory for more details.

### Call-site logging

If desired, the logger may be configured to emit the file name and line number of the call-site that generated the log message.  This is _disabled_ by default.

Call-site logging may be enabled via the `LogCallSite` configuration function when configuring a new logger:

```go
logger, closelog, err := ulog.NewLogger(ctx, ulog.LogCallSite(true))
```
