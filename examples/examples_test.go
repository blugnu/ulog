package ulog

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/blugnu/ulog"
)

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
		ulog.LoggerFormat(ulog.Logfmt(
			ulog.LogfmtLevels(map[ulog.Level]string{
				ulog.InfoLevel: "FYI  ",
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
	logger, cfn, err := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.Target(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat(ulog.Logfmt()),
				ulog.TargetTransport(ulog.StdioTransport()),
			),
		),
	)
	defer cfn()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger.Info("this should be logged")
}

func TestMux(t *testing.T) {
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.Format("logfmt", ulog.Logfmt()),
			ulog.Target(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat("logfmt"),
				ulog.TargetTransport(ulog.StdioTransport()),
			),
			ulog.Target(
				ulog.TargetLevel(ulog.ErrorLevel),
				ulog.TargetFormat("logfmt"),
				ulog.TargetTransport(ulog.StdioTransport(
					ulog.StdioOutput(os.Stderr),
				)),
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
