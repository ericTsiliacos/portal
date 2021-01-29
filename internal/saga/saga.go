package saga

import (
	"fmt"
)

type Saga struct {
	Steps   []Step
	results []string
}

func (s *Saga) Run() error {
	for i := 0; i < len(s.Steps); i++ {
		step := s.Steps[i]

		value, err := step.Run()

		if err != nil {
			err := retry(step.Retries, func() (err error) {
				_, err = s.Steps[i].Run()
				return err
			})

			if err != nil {
				compensate(s.Steps[0:i], s.results[0:i])
				return err
			}
		}

		s.results = append(s.results, value)
	}

	return nil
}

func compensate(steps []Step, results []string) error {
	compensateSteps := reverseSteps(steps)
	values := reverseResults(results)

	for i := 0; i < len(compensateSteps); i++ {
		compensationFn := compensateSteps[i].Compensate

		if compensationFn != nil {
			_, err := compensateSteps[i].Compensate(values[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func reverseSteps(steps []Step) []Step {
	for i, j := 0, len(steps)-1; i < j; i, j = i+1, j-1 {
		steps[i], steps[j] = steps[j], steps[i]
	}
	return steps
}

func reverseResults(results []string) []string {
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results
}

func retry(attempts int, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

	}

	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

type Step struct {
	Name       string
	Run        func() (string, error)
	Compensate func(string) (string, error)
	Retries    int
}
