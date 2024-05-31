// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package benchmarks

import (
	"context"
	"io"

	"github.com/blugnu/ulog"
)

// func newUlog(fields map[string]any) (ulog.Logger, ulog.CloseFunc) {
// 	logger, _ := ulog.NewLogger(
// 		context.Background(),
// 		ulog.Transport(&ulog.StdioWriter{Writer: io.Discard}),
// 	)
// 	return logger.WithFields(fields)
// }

func newUlogLogFmt(level ulog.Level, opt ...ulog.LoggerOption) (ulog.Logger, ulog.CloseFn) {
	opt = append(opt,
		ulog.LoggerLevel(level),
		ulog.LoggerOutput(io.Discard),
		ulog.LoggerFormat(ulog.Logfmt()),
	)

	logger, cfn, _ := ulog.NewLogger(context.Background(), opt...)

	return logger, cfn
}

func newUlogJson(level ulog.Level) (ulog.Logger, ulog.CloseFn) {
	logger, cfn, _ := ulog.NewLogger(context.Background(),
		ulog.LoggerLevel(level),
		ulog.LoggerFormat(ulog.NewJSONFormatter()),
		ulog.LoggerOutput(io.Discard),
	)

	return logger, cfn
}

func newUlogMsgpack(level ulog.Level) (ulog.Logger, ulog.CloseFn) {
	logger, cfn, _ := ulog.NewLogger(context.Background(),
		ulog.LoggerLevel(level),
		ulog.LoggerFormat(ulog.MsgpackFormatter()),
		ulog.LoggerOutput(io.Discard),
	)

	return logger, cfn
}

func newUlogMux(level ulog.Level) (ulog.Logger, ulog.CloseFn) {
	logger, cfn, _ := ulog.NewLogger(
		context.Background(),
		ulog.Mux(
			ulog.Target(
				ulog.TargetLevel(level),
				//				ulog.TargetFormat(&ulog.LogfmtFormat{Levels: ulog.DefaultLogfmtLevels}),
				// 				ulog.TargetTransport(&ulog.StdioWriter{Writer: io.Discard}),
				ulog.TargetFormat(ulog.Logfmt()),
				ulog.TargetTransport(ulog.StdioTransport(
					ulog.StdioOutput(io.Discard),
				)),
			),
		),
	)
	return logger, cfn
}

// func newDisabledUlog(fields map[string]any) ulog.Logger {
// 	logger := ulog.NewLogger(context.Background(),
// 		ulog.SetLevel(ulog.ErrorLevel),
// 		ulog.Transport(&ulog.StdioWriter{Writer: io.Discard}),
// 	)
// 	return logger.WithFields(fields)
// }

func fakeUlogFields() map[string]any {
	return map[string]any{
		"int":     _tenInts[0],
		"ints":    _tenInts,
		"string":  _tenStrings[0],
		"strings": _tenStrings,
		"time":    _tenTimes[0],
		"times":   _tenTimes,
		// "user1":   _oneUser,
		// "user2":   _oneUser,
		// "users":   _tenUsers,
		"error": errExample,
	}
}

func fakeUlogArgs() []any {
	return []any{
		"int", _tenInts[0],
		"ints", _tenInts,
		"string", _tenStrings[0],
		"strings", _tenStrings,
		"time", _tenTimes[0],
		"times", _tenTimes,
		// "user1", _oneUser,
		// "user2", _oneUser,
		// "users", _tenUsers,
		"error", errExample,
	}
}
