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

A configurable, structured logging library for Go that does not sacrifice (too much)
efficiency in exchange for convenience and flexibility.

# Features

_Features unique to `ulog`_:

- [x] **Automatic Context Enrichment** - logs may be enriched with context from the context
  provided by the caller or from context carried by an error (via support for the
  [blugnu/errorcontext](https://github.com/blugnu/errorcontext) module); automatic contextual
  enrichment can be embedded by registering `ulog.ContextEnricher` functions with the logger

- [x] **Multiplexing** - (_optionally_) send logs to multiple destinations simultaneously with
  independently configured formatting and leveling; for example, full logs could be sent to
  an aggregator in `msgpack` or `JSON` format while logs at error level and above are sent
  to `os.Stderr` in in human-readable `logfmt` format

- [x] **Testable** - use the provided mock logger to verify that the logs expected by your
  observability alerts are actually produced!

_Features you'd expect of any logger_:

- [x] **Highly Configurable** - configure your logger to emit logs in the format you want,
  with the fields you want, at the level you want, sending them to the destination (or
  destination**s**) you want; the default configuration is designed to be sensible and
  useful out-of-the-box, but can be easily customised to suit your needs

- [x] **Structured Logs** - logs are emitted in a structured format (`logfmt` by default)
  that can be easily parsed by log aggregators and other tools; a JSON formatter is also
  provided

- [x] **Efficient** - allocations are kept to a minimum and conditional flows eliminated
  where-ever possible to ensure that the overhead of logging is kept to a minimum

- [x] **Intuitive** - a consistent API using Golang idioms that is easy and intuitive to use

# Installation

```bash
go get github.com/blugnu/ulog
```

# Tech Stack

`blugnu/ulog` is built on the following main stack:

<!-- markdownlint-disable MD013 // line length-->
- <img width='25' height='25' src='https://img.stackshare.io/service/1005/O6AczwfV_400x400.png' alt='Golang'/> [Golang](http://golang.org/) – Languages
- <img width='25' height='25' src='https://img.stackshare.io/service/11563/actions.png' alt='GitHub Actions'/> [GitHub Actions](https://github.com/features/actions) – Continuous Integration
<!-- markdownlint-enable MD013 -->

Full tech stack [here](/techstack.md)

# Usage

A logger is created using the `ulog.NewLogger` factory function.

This function returns a logger, a function to close the logger and an error if the logger
could not be created. The close function _must_ be called when the logger is no longer
required to release any resources it may have acquired and any batched log entries are
flushed.

The logger may be configured using a set of configuration functions that are passed as arguments to the factory function.

A minimal example of using `ulog`:

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

This example uses a logger with default configuration that writes log entries at
`InfoLevel` and above to `os.Stdout` in [`logfmt`](https://brandur.org/logfmt) format.

Running this code would produce output similar to the following:

```bash
time=2023-11-23T12:35:04.789346Z level=INFO  msg="hello world"
time=2023-11-23T12:35:04.789347Z level=ERROR msg="oops!"
```

<!-- FUTURE: More complete examples of different use cases are 
             provided in the [examples](examples) directory of this repo. -->

# Configuration

> **_NOTE:_** _This section deals with configuration of a simple logger sending output to a
  single `io.Writer` such as `os.Stdout`.
  [Configuration of a multiplexing logger](#mux-configuration) is described
  separately._

The default configuration is designed to be sensible and useful out-of-the-box but can be
customised to suit your needs.

The [Functional Options Pattern](https://www.sohamkamani.com/golang/options-pattern/) is
employed for configuration.  To configure the logger pass the required option functions
as additional arguments to the `NewLogger()` factory.  All options have sensible defaults.

<!-- markdownlint-disable MD013 // line length -->
| Option | Description | Options | Default (_if not configured_) |
|--------|-------------|---------|-------------------------------|
| `LogCallSite` | Whether to include the file name and line number of the call-site in the log message | `true`<br>`false` | `false` |
| `LoggerLevel` | The minimum log level to emit | see: [Log Levels](#log-levels) | `ulog.InfoLevel` |
| `LoggerFormat` | The formatter to be used when writing logs | `ulog.NewLogfmtFormatter`<br>`ulog.NewJSONFormatter`<br>`ulog.NewMsgpackFormatter` | `ulog.NewLogfmtFormatter()` |
| `LoggerOutput` | The destination to which logs are written | any `io.Writer` | `os.Stdout` |
| `Mux` | Use a multiplexer to send logs to multiple destinations simultaneously | `ulog.Mux`. See: [Mux Configuration](#mux-configuration) | - |
<!-- markdownlint-enable MD013 -->

## LoggerLevel

### Log Levels

The following levels are supported (in ascending order of significance):

| Level | Description |  |
|-------|-------------|--|
| `ulog.TraceLevel` | low-level diagnostic logs | |
| `ulog.DebugLevel` | debug messages | |
| `ulog.InfoLevel` | informational messages | _default_ |
| `ulog.WarnLevel` | warning messages | |
| `ulog.ErrorLevel` | error messages | |
| `ulog.FatalLevel` | fatal errors | |

The default log level may be overridden using the `LoggerLevel` configuration function
to `NewLogger`, e.g:

```go
ulog.NewLogger(
    ulog.LoggerLevel(ulog.DebugLevel)
)
```

In this example, the logger is configured to emit logs at `DebugLevel` and above.  `TraceLevel` logs would not be emitted.

`LoggerLevel` establishes the minimum log level for the logger.

If the logger is configured with a multiplexer the level for each mux target may be set
independently but any target level that is lower than the logger level is effectively
ignored.

| Mux Target Level  | Logger Level      | _Effective_ Mux Target Level |
|-------------------|-------------------|------------------------------|
| `ulog.InfoLevel`  | `ulog.InfoLevel`  | `ulog.InfoLevel`             |
| `ulog.DebugLevel` | `ulog.InfoLevel`  | `ulog.InfoLevel`             |
| `ulog.DebugLevel` | `ulog.TraceLevel` | `ulog.DebugLevel`            |
| `ulog.InfoLevel`  | `ulog.WarnLevel`  | `ulog.WarnLevel`             |

## LoggerFormat

The following formatters are supported:

| Format | LoggerFormat Option | Description |
|--------|---------------------|-------------|
| [logfmt](https://brandur.org/logfmt) | `ulog.NewLogfmtFormatter` | a simple, structured format that is easy for both humans and machines to read<br><br>_This is the default formatter if none is configured_ |
| [JSON](https://www.json.org/) | `ulog.NewJSONFormatter` | a structured format that is easy for machines to parse but can be noisy for humans |
| [msgpack](https://msgpack.org/) | `ulog.NewMsgpackFormatter` | an efficient structured, binary format for machine use, unintelligible to (most) humans |

Formatters offer a number of configuration options that can be set via functions supplied
to their factory function. For example, to configure the name used for the `time` field by a
`logfmt` formatter:

```go
log, closelog, _ := ulog.NewLogger(ctx,
   ulog.LoggerFormat(ulog.LogfmtFormatter(
      ulog.LogfmtFieldNames(map[ulog.FieldId]string{
         ulog.TimeField: "timestamp",
      }),
   )),
)
```

## LoggerOutput

Logger output may be written to any `io.Writer`.  The default output is `os.Stdout`.

The output may be configured  via the `LoggerOutput` configuration function when configuring a new logger:

```go
logger, closefn, err := ulog.NewLogger(ctx, ulog.LoggerOutput(io.Stderr))
```

## LogCallsite

If desired, the logger may be configured to emit the file name and line number of the
call-site that produced the log message.  This is _disabled_ by default.

Call-site logging may be enabled via the `LogCallSite` configuration function when configuring a new logger:

```go
logger, closelog, err := ulog.NewLogger(ctx, ulog.LogCallSite(true))
```

If `LogCallsite` is enabled call site information is added to all log entries, including
multiplexed output where configured, regardless of the level.

# Mux Configuration

Logs may be sent to _multiple_ destinations simultaneously using a `ulog.Mux`.  For example,
full logs could be sent to an aggregator using `msgpack` format, while `logfmt` formatted
logs at error level and above are also sent to `os.Stderr`.
