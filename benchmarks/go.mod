module github.com/blugnu/ulog/benchmarks

go 1.21

replace github.com/blugnu/ulog => ../

require (
	github.com/blugnu/ulog v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/blugnu/errorcontext v0.2.1 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
