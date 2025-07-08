package profiler

import (
	"errors"
	"os"
	"runtime/trace"
)

func StartTrace() error {
	if f, err := os.Create(traceFile); err != nil {
		return err
	} else {
		if err := trace.Start(f); err != nil {
			return err
		}
		return nil
	}
}

func StopTrace() error {
	trace.Stop()
	if s, err := traceReview(); err != nil {
		return errors.New(s)
	}
	return nil
}
