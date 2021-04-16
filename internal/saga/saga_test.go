package saga

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSagaMultipleSuccessfulSteps(t *testing.T) {
	globalState := 0
	steps := []Step{
		{
			Name: "addOne",
			Run:  func() (err error) { globalState = globalState + 1; return },
			Undo: func() (err error) { globalState = globalState - 1; return },
		},
		{
			Name: "addOne",
			Run:  func() (err error) { globalState = globalState + 1; return },
		},
	}

	saga := New(steps)
	errs := saga.Run()

	assert.Equal(t, globalState, 2)
	assert.Empty(t, errs)
}

func TestSagaWithFailure(t *testing.T) {
	globalState := 0
	steps := []Step{
		{
			Name: "addOne",
			Run:  func() (err error) { globalState = globalState + 1; return },
			Undo: func() (err error) { globalState = globalState - 1; return },
		},
		{
			Name: "addTwo",
			Run:  func() (err error) { globalState = globalState + 2; return },
			Undo: func() (err error) { globalState = globalState - 2; return },
		},
		{
			Name: "boom!",
			Run:  func() (err error) { return errors.New("uh oh!") },
		},
	}

	saga := New(steps)
	errs := saga.Run()

	assert.Equal(t, globalState, 0)
	assert.Equal(t, errs, []string{"uh oh!"})
}

func TestSagaWithMissingUndo(t *testing.T) {
	globalState := 0
	steps := []Step{
		{
			Name: "addOne",
			Run:  func() (err error) { globalState = globalState + 1; return },
			Undo: func() (err error) {
				globalState = globalState - 1
				return
			},
		},
		{
			Name: "addTwo",
			Run:  func() (err error) { globalState = globalState + 2; return },
		},
		{
			Name: "boom!",
			Run:  func() (err error) { return errors.New("uh oh!") },
		},
	}

	saga := New(steps)
	errs := saga.Run()

	assert.Equal(t, globalState, 2)
	assert.Equal(t, errs, []string{"uh oh!"})
}

func TestSagaWithUndoFailure(t *testing.T) {
	globalState := 0
	steps := []Step{
		{
			Name: "addOne",
			Run:  func() (err error) { globalState = globalState + 1; return },
			Undo: func() (err error) {
				return errors.New("recovery error")
			},
		},
		{
			Name: "boom!",
			Run:  func() (err error) { return errors.New("uh oh!") },
		},
	}

	saga := New(steps)
	errs := saga.Run()

	assert.Equal(t, globalState, 1)
	assert.Equal(t, errs, []string{"uh oh!", "recovery error"})
}
