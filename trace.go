package ulog

import (
	"fmt"
	"log"
)

var (
	traceFn = func(...any) { /* NO-OP */ }
	trace   = func(a ...any) { traceFn(a...) }
	tracef  = func(s string, a ...any) { traceFn(fmt.Sprintf(s, a...)) }
)

func EnableTraceLogs(fn func(...any)) {
	if traceFn = fn; fn == nil {
		traceFn = func(a ...any) {
			a = append([]any{"ULOG:TRACE"}, a...)
			log.Println(a...)
		}
	}
}
