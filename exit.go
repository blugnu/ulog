package ulog

import "os"

// ExitFn is a func var allowing `ulog` code paths that reach an `os.Exit` call
// to be replaced by a non-exiting behaviour.
//
// This is primarily used in `ulog` unit tests but also allows (e.g.) a
// microservice to intercept exit calls and perform a controlled exit,
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
