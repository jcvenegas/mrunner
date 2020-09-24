package tests

import (
	"time"
)

type TestEnv struct {
	WorkDir string
	Runtime string
}

type TestsResult struct {
	Error    error
	Duration time.Duration
	TestID   string
}

func (t *TestsResult) SetError(err error) {
	if err == nil {
		t.Error = err
	}
}

type Test interface {
	Setup() error
	Run(TestEnv) (TestsResult, error)
	TearDown() error
	ID() string
}
