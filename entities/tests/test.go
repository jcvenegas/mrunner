package tests

import (
	"time"
)

type TestEnv struct {
	WorkDir string
	Runtime string
}

type Result struct {
	Error    error
	Duration time.Duration
	TestID   string
}

func (t *Result) SetError(err error) {
	if err == nil {
		t.Error = err
	}
}

type Test interface {
	Setup() error
	Run(TestEnv) (Result, error)
	TearDown() error
	ID() string
}
