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

func EnableTrace() {
	traceFn = func(a ...any) {
		a = append([]any{"[ulog trace]"}, a...)
		log.Println(a...)
	}
}
