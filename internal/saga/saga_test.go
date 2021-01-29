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
			Name:       "addOne",
			Run:        func() (string, error) { globalState = globalState + 1; return "", nil },
			Compensate: func(input string) (string, error) { globalState = globalState - 1; return "", nil },
		},
		{
			Name: "addOne",
			Run:  func() (string, error) { globalState = globalState + 1; return "", nil },
		},
	}

	saga := Saga{Steps: steps}
	saga.Run()

	assert.Equal(t, globalState, 2)
}

func TestSagaWithFailure(t *testing.T) {
	globalState := 0
	steps := []Step{
		{
			Name:       "addOne",
			Run:        func() (string, error) { globalState = globalState + 1; return "", nil },
			Compensate: func(input string) (string, error) { globalState = globalState - 1; return "", nil },
		},
		{
			Name:       "addTwo",
			Run:        func() (string, error) { globalState = globalState + 2; return "", nil },
			Compensate: func(input string) (string, error) { globalState = globalState - 2; return "", nil },
		},
		{
			Name: "boom!",
			Run:  func() (string, error) { return "", errors.New("uh oh!") },
		},
	}

	saga := Saga{Steps: steps}
	saga.Run()

	assert.Equal(t, globalState, 0)
}

func TestSagaWithMissingCompensate(t *testing.T) {
	globalState := 0
	compensationInput := ""
	steps := []Step{
		{
			Name: "addOne",
			Run:  func() (string, error) { globalState = globalState + 1; return "compensate", nil },
			Compensate: func(input string) (string, error) {
				globalState = globalState - 1
				compensationInput = input
				return "", nil
			},
		},
		{
			Name: "addTwo",
			Run:  func() (string, error) { globalState = globalState + 2; return "", nil },
		},
		{
			Name: "boom!",
			Run:  func() (string, error) { return "", errors.New("uh oh!") },
		},
	}

	saga := Saga{Steps: steps}
	saga.Run()

	assert.Equal(t, globalState, 2)
	assert.Equal(t, compensationInput, "compensate")
}

func TestSagaWithRetries(t *testing.T) {
	globalState := 0
	totalRetries := 2
	steps := []Step{
		{
			Name:       "addOne",
			Run:        func() (string, error) { globalState = globalState + 1; return "", nil },
			Compensate: func(input string) (string, error) { globalState = globalState - 2; return "", nil },
		},
		{
			Name: "retries",
			Run: func() (string, error) {
				if totalRetries == 0 {
					globalState = globalState + 1
					return "", nil
				} else {
					totalRetries = totalRetries - 1
					return "", errors.New("boom!")
				}
			},
			Retries: 2,
		},
	}

	saga := Saga{Steps: steps}
	saga.Run()

	assert.Equal(t, globalState, 2)
}
