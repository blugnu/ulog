package examples

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/blugnu/ulog"
)

func TestLogtail(t *testing.T) {
	if _, err := os.Stat(".token"); errors.Is(err, os.ErrNotExist) {
		log.Fatal("no token\n" +
			"> to run this example you must create a .token file in the same directory as the test;\n" +
			"> the file must contain a valid logtail source token\n" +
			"> (see: https://betterstack.com/docs/logs/logging-start/#step-2-test-the-pipes)")
	}

	ulog.EnableTraceLogs(nil)
	logger, closelog, _ := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat(ulog.LogfmtFormatter()),
				ulog.TargetTransport(ulog.StdioTransport(os.Stdout)),
			),
			ulog.MuxTarget(
				ulog.TargetLevel(ulog.InfoLevel),
				ulog.TargetFormat(ulog.MsgpackFormatter()),
				ulog.TargetTransport(
					ulog.LogtailTransport(
						ulog.LogtailSourceToken(".token"),
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
		"struct": struct {
			RandomLine   string
			RandomNumber int
		}{
			RandomLine: []string{
				"mary had a little lamb",
				"its fleecec was white as snow",
				"and everywhere that mary went",
				"that lamb was sure to go"}[rand.Intn(3)],
			RandomNumber: rand.Int(),
		},
	})
	logger.Info("this message should have a field")
}
