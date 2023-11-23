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
	"fmt"
	"log/slog"
	"testing"

	"github.com/blugnu/ulog"
)

func BenchmarkDisabledWithoutFields(b *testing.B) {
	if !runDisabled {
		return
	}

	b.Logf("Logging at a disabled level without any structured context.")

	if runAll || runLogrus {
		b.Run("sirupsen/logrus", func(b *testing.B) {
			logger := newDisabledLogrus()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
		b.Run("sirupsen/logrus-with-args", func(b *testing.B) {
			logger := newDisabledLogrus()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(fmt.Sprintf(getMessagef(0), "string"))
				}
			})
		})
	}

	if runAll || runSlog {
		b.Run("slog", func(b *testing.B) {
			logger := newDisabledSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
		b.Run("slog.LogAttrs", func(b *testing.B) {
			logger := newDisabledSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.LogAttrs(context.Background(), slog.LevelInfo, getMessage(0))
				}
			})
		})
	}

	b.Run("blugnu/ulog-json", func(b *testing.B) {
		logger, cfn := newUlogJson(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessagef(0))
			}
		})
	})
	b.Run("blugnu/ulog-json-with-args", func(b *testing.B) {
		logger, cfn := newUlogJson(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof(getMessagef(0), "string")
			}
		})
	})
	b.Run("blugnu/ulog-logfmt", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt-with-args", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof(getMessagef(0), "string")
			}
		})
	})
	b.Run("blugnu/ulog-mux", func(b *testing.B) {
		logger, cfn := newUlogMux(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-mux-with-args", func(b *testing.B) {
		logger, cfn := newUlogMux(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof(getMessagef(0), "string")
			}
		})
	})
}

func BenchmarkDisabledAccumulatedContext(b *testing.B) {
	if !runDisabled {
		return
	}

	b.Logf("Logging at a disabled level with some accumulated context.")
	if runAll || runLogrus {
		b.Run("sirupsen/logrus", func(b *testing.B) {
			logger := newDisabledLogrus().WithFields(fakeLogrusFields())
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
	}

	if runAll || runSlog {
		b.Run("slog", func(b *testing.B) {
			logger := newDisabledSlog(fakeSlogFields()...)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
		b.Run("slog.LogAttrs", func(b *testing.B) {
			logger := newDisabledSlog(fakeSlogFields()...)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.LogAttrs(context.Background(), slog.LevelInfo, getMessage(0))
				}
			})
		})
	}

	b.Run("blugnu/ulog-logfmt", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.ErrorLevel)
		defer cfn()

		logger = logger.WithFields(fakeUlogFields())

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt-levelled", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.ErrorLevel)
		defer cfn()

		info := logger.AtLevel(ulog.InfoLevel).WithFields(fakeUlogFields())

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				info.Log(getMessage(0))
			}
		})
	})
}

func BenchmarkDisabledAddingFields(b *testing.B) {
	if !runDisabled {
		return
	}

	b.Logf("Logging at a disabled level, adding context at each log site.")

	if runAll || runLogrus {
		b.Run("sirupsen/logrus", func(b *testing.B) {
			logger := newDisabledLogrus()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
				}
			})
		})
	}

	if runAll || runSlog {
		b.Run("slog", func(b *testing.B) {
			logger := newDisabledSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0), fakeSlogArgs()...)
				}
			})
		})
		b.Run("slog.LogAttrs", func(b *testing.B) {
			logger := newDisabledSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.LogAttrs(context.Background(), slog.LevelInfo, getMessage(0), fakeSlogFields()...)
				}
			})
		})
	}

	b.Run("blugnu/ulog-json", func(b *testing.B) {
		logger, cfn := newUlogJson(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeUlogFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeUlogFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt-levelled", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.AtLevel(ulog.InfoLevel).
					WithFields(fakeUlogFields()).
					Log(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-mux", func(b *testing.B) {
		logger, cfn := newUlogMux(ulog.ErrorLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeUlogFields()).Info(getMessage(0))
			}
		})
	})
}

func BenchmarkWithoutFields(b *testing.B) {
	b.Logf("Logging without any structured context.")

	if runAll || runLogrus {
		b.Run("sirupsen/logrus", func(b *testing.B) {
			logger := newLogrus()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
	}

	if runAll || runSlog {
		b.Run("slog", func(b *testing.B) {
			logger := newSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
		b.Run("slog.LogAttrs", func(b *testing.B) {
			logger := newSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.LogAttrs(context.Background(), slog.LevelInfo, getMessage(0))
				}
			})
		})
	}

	b.Run("blugnu/ulog-json", func(b *testing.B) {
		logger, cfn := newUlogJson(ulog.InfoLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.InfoLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-mux", func(b *testing.B) {
		logger, cfn := newUlogMux(ulog.InfoLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
}

func BenchmarkCallsiteOverhead(b *testing.B) {
	b.Logf("Logging with/without call-site reporting")

	if runAll || runLogrus {
		b.Run("sirupsen/logrus-no-callsite", func(b *testing.B) {
			logger := newLogrus()
			logger.SetReportCaller(false)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})

		b.Run("sirupsen/logrus-with-callsite", func(b *testing.B) {
			logger := newLogrus()
			logger.SetReportCaller(true)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
	}

	b.Run("blugnu/ulog-no-callsite", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.InfoLevel, ulog.LogCallsite(false))
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})

	b.Run("blugnu/ulog-with-callsite", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.InfoLevel, ulog.LogCallsite(true))
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
}

func BenchmarkAccumulatedContext(b *testing.B) {
	b.Logf("Logging with some accumulated context.")

	if runAll || runLogrus {
		b.Run("sirupsen/logrus", func(b *testing.B) {
			logger := newLogrus().WithFields(fakeLogrusFields())
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
	}

	if runAll || runSlog {
		b.Run("slog", func(b *testing.B) {
			logger := newSlog(fakeSlogFields()...)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0))
				}
			})
		})
		b.Run("slog.LogAttrs", func(b *testing.B) {
			logger := newSlog(fakeSlogFields()...)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.LogAttrs(context.Background(), slog.LevelInfo, getMessage(0))
				}
			})
		})
	}

	b.Run("blugnu/ulog-json", func(b *testing.B) {
		logger, cfn := newUlogJson(ulog.InfoLevel)
		defer cfn()

		logger = logger.WithFields(fakeUlogFields())

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.InfoLevel)
		defer cfn()

		logger = logger.WithFields(fakeUlogFields())

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-mux", func(b *testing.B) {
		logger, cfn := newUlogMux(ulog.InfoLevel)
		defer cfn()

		logger = logger.WithFields(fakeUlogFields())

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(getMessage(0))
			}
		})
	})
}

func BenchmarkAddingFields(b *testing.B) {
	b.Logf("Logging with additional context at each log site.")

	if runAll || runLogrus {
		b.Run("sirupsen/logrus", func(b *testing.B) {
			logger := newLogrus()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
				}
			})
		})
	}

	if runAll || runSlog {
		b.Run("slog", func(b *testing.B) {
			logger := newSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info(getMessage(0), fakeSlogArgs()...)
				}
			})
		})
		b.Run("slog.LogAttrs", func(b *testing.B) {
			logger := newSlog()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.LogAttrs(context.Background(), slog.LevelInfo, getMessage(0), fakeSlogFields()...)
				}
			})
		})
	}

	b.Run("blugnu/ulog-json", func(b *testing.B) {
		logger, cfn := newUlogJson(ulog.InfoLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeUlogFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-logfmt", func(b *testing.B) {
		logger, cfn := newUlogLogFmt(ulog.InfoLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeUlogFields()).Info(getMessage(0))
			}
		})
	})
	b.Run("blugnu/ulog-mux", func(b *testing.B) {
		logger, cfn := newUlogMux(ulog.InfoLevel)
		defer cfn()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.WithFields(fakeUlogFields()).Info(getMessage(0))
			}
		})
	})
}
