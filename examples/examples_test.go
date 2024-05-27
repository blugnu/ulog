package examples

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/blugnu/ulog"
)

func TestDefault(t *testing.T) {
	ctx := context.Background()
	log, closelog, _ := ulog.NewLogger(ctx)
	defer closelog()

	log.Info("hello world")
	log.Error("oops!")
}

func TestCustomTimeField(t *testing.T) {
	ctx := context.Background()
	log, closelog, _ := ulog.NewLogger(ctx,
		ulog.LoggerFormat(ulog.LogfmtFormatter(
			ulog.LogfmtFieldNames(map[ulog.FieldId]string{
				ulog.TimeField: "timestamp",
			}),
		)),
	)
	defer closelog()

	log.Info("hello world")
	log.Error("oops!")
}

func TestStdioConfiguration(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.LogCallsite(true),
		ulog.LoggerLevel(ulog.InfoLevel),
		ulog.LoggerOutput(os.Stdout),
		ulog.LoggerFormat(ulog.LogfmtFormatter()),
	)
	defer closelog()

	info := logger.AtLevel(ulog.InfoLevel)
	info.WithFields(map[string]any{"key": "value"}).Log("this should not be logged")

	logger.Infof("this should %s be logged", "not")
	logger.Trace("this should not be logged")
	logger.Error("this should be logged")

	t.Run("with fields", func(t *testing.T) {
		logger := logger.WithFields(map[string]any{
			"key1": "value1",
			"key2": 42,
		})

		logger.Info("info with fields")
		logger.Debug("debug with fields")
	})
}

func TestDisabled(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.LogCallsite(true),
		ulog.LoggerLevel(ulog.ErrorLevel),
	)
	defer closelog()

	info := logger.AtLevel(ulog.InfoLevel)
	info.WithFields(map[string]any{"key": "value"}).Log("this should not be logged")

	logger.Infof("this should %s be logged", "not")
	logger.Trace("this should not be logged")
	logger.Error("this should be logged")

	t.Run("with fields", func(t *testing.T) {
		logger := logger.WithFields(map[string]any{
			"key1": "value1",
			"key2": 42,
		})

		logger.Info("info with fields")
		logger.Debug("debug with fields")
	})
}

func TestJsonFormatter(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.LoggerLevel(ulog.DebugLevel),
		ulog.LoggerFormat(ulog.NewJSONFormatter()),
	)
	defer closelog()

	logger.Info("this is a test")
	logger.Trace("this should not be logged")

	t.Run("with fields", func(t *testing.T) {
		logger := logger.WithFields(map[string]any{
			"key1": "value1",
			"key2": 42,
		})

		logger.Info("info with fields")
		logger.Debug("debug with fields")
	})
}

func TestLogfmt(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.LoggerLevel(ulog.DebugLevel),
		ulog.LoggerFormat(ulog.LogfmtFormatter(
			ulog.LogfmtLevelLabels(map[ulog.Level]string{
				ulog.InfoLevel: "FYI",
			}),
		)),
	)
	defer closelog()

	logger.Info("this is a test")
	logger.Trace("this should not be logged")

	t.Run("with fields", func(t *testing.T) {
		logger := logger.WithFields(map[string]any{
			"key1": "value1",
			"key2": 42,
		})

		logger.Info("info with fields")
		logger.Debug("debug with fields")
	})
}

func TestStress(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.LoggerOutput(io.Discard),
	)
	defer closelog()

	for i := 1; i < 10000; i++ {
		logger.Info("this is a test entry of a meaningful length, or so I should think")
	}
}

func TestSimpleMux(t *testing.T) {
	logger, cfn, _ := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat(ulog.LogfmtFormatter()),
				ulog.TargetTransport(ulog.StdioTransport(os.Stdout)),
			),
		),
	)
	defer cfn()

	logger.Info("this should be logged")
}

func TestMux(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.MuxFormat("logfmt", ulog.LogfmtFormatter()),
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat("logfmt"),
				ulog.TargetTransport(
					ulog.StdioTransport(os.Stdout),
				),
			),
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.ErrorLevel),
				ulog.TargetFormat("logfmt"),
				ulog.TargetTransport(
					ulog.StdioTransport(os.Stderr),
				),
			),
		),
	)
	defer closelog()

	logger.Info("-- new run ---------------------------------------------")
	logger.Info("test log message, info level")
	logger.Error("test log message, error level")

	for i := 1; i <= 10; i++ {
		logger.Infof(fmt.Sprintf("this is muxed test log #%d", i))
	}

	logger = logger.WithFields(map[string]any{
		"key": "value",
	})
	logger.Info("this message should have a field")
}

func TestLogtailMux(t *testing.T) {
	ulog.EnableTraceLogs(nil)
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.MuxFormat("logfmt", ulog.LogfmtFormatter()),
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat("logfmt"),
				ulog.TargetTransport(
					ulog.StdioTransport(os.Stdout),
				),
			),
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat(
					ulog.MsgpackFormatter(
						ulog.MsgpackKeys(map[ulog.FieldId]string{
							ulog.TimeField: "dt",
						}),
					),
				),
				ulog.TargetTransport(
					ulog.LogtailTransport(
						ulog.LogtailSourceToken("../.logtail/token"),
					),
				),
			),
		),
	)
	defer closelog()

	logger.Info("-- new run ---------------------------------------------")
	logger.Info("test log message, info level")
	logger.Error("test log message, error level")

	for i := 1; i <= 10; i++ {
		logger.Infof(fmt.Sprintf("this is muxed test log #%d", i))
	}

	logger = logger.WithFields(map[string]any{
		"key": "value",
	})
	logger.Info("this message should have a field")
}
