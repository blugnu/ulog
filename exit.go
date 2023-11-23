package ulog

import "os"

type FatalExitOption int

const (
	ExitAlways FatalExitOption = iota
	ExitNever
	ExitWhenLogged
)

// ExitOnFatalLog determines the behaviour of `ulog` when a fatal log
// entry is written.
//
//   - if ExitAlways, then the logger will always call exit when Fatal()
//     or Fatalf() is called, even if a Nul() logger is used
//
//   - if ExitNever, then the logger will never call exit
//
//   - if ExitWhenLogged, then the logger will call exit when Fatal()
//     or Fatalf() are called on a logger that is not Nul()
//
// Note that a logger configured to write to a void destination such as
// io.Discard is not a Nul() logger.
//
// To completely prevent `ulog` causing a process to terminate, replace
// `ulog.ExitFn`.
var ExitOnFatalLog = ExitAlways

// ExitFn is a func var allowing `ulog` code paths that reach an `os.Exit` call
// to be replaced by a non-exiting behaviour.
//
// This is primarily used in `ulog` unit tests but also allows (e.g.) a
// microservice to intercept such exit calls and perform a controlled exit,
// even when a log call results in termination of the microservice.
var ExitFn func(int) = os.Exit

// exit calls the `ExitFn` with the specified exit code.  Code paths in `ulog`
// that require termination of the process (e.g. `log.FatalError()`) call this
// `exit` function which in turn calls the `ExitFn` func var.
//
// To prevent `ulog` causing a process to terminate, replace `ExitFn`.
func exit(code int) {
	ExitFn(code)
}
