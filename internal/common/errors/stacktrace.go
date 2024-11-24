package errors

import (
	"fmt"

	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/pkg/errors"
)

type Trace []string

func StackTraces(err error) []Trace {
	var traces []Trace
	recursiveTraces(err, &traces)
	return traces
}

func recursiveTraces(err error, traces *[]Trace) {
	if err == nil {
		return
	}

	if c, ok := err.(causer); ok {
		recursiveTraces(c.Cause(), traces)
	}

	if join, ok := err.(unwrapper); ok {
		for _, err2 := range join.Unwrap() {
			recursiveTraces(err2, traces)
		}
	}
	if tr, ok := err.(stackTracer); ok {
		*traces = append(
			*traces,
			slices.Map(tr.StackTrace(), func(f errors.Frame) string {
				return fmt.Sprintf("%+v", f)
			}),
		)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

type unwrapper interface {
	Unwrap() []error
}
