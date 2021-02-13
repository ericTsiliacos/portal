package saga

import (
	"fmt"
)

type Saga struct {
	Steps   []Step
	results []string
	Verbose bool
}

type Step struct {
	Name       string
	Run        func() (string, error)
	Compensate func(string) (string, error)
	Retries    int
	Exclude    bool
}

func (s *Saga) Run() []string {
	var errors []string

	steps := filter(s.Steps, func(s Step) bool {
		return !s.Exclude
	})

	for i := 0; i < len(steps); i++ {
		step := steps[i]

		value, err := step.Run()

		if err != nil {
			err := retry(step, func() (err error) {
				_, err = steps[i].Run()
				return err
			})

			if s.Verbose {
				fmt.Printf("run step: %s", step.Name)
				fmt.Println()
				fmt.Println("output:")
				fmt.Println(err)
				fmt.Println()
			}

			if err != nil {
				latestError := compensate(steps[0:i], s.results[0:i], s.Verbose)
				return append(errors, err.Error(), latestError)
			}
		}

		s.results = append(s.results, value)

		if s.Verbose {
			fmt.Printf("run step: %s", step.Name)
			fmt.Println()
			fmt.Println("output:")
			fmt.Println(value)
			fmt.Println()
		}
	}

	return errors
}

func filter(steps []Step, test func(Step) bool) (ret []Step) {
	for _, step := range steps {
		if test(step) {
			ret = append(ret, step)
		}
	}
	return
}

func compensate(steps []Step, results []string, verbose bool) string {
	compensateSteps := reverseSteps(steps)
	values := reverseResults(results)

	for i := 0; i < len(compensateSteps); i++ {
		compensationStep := compensateSteps[i]
		compensationFn := compensationStep.Compensate

		if compensationFn != nil {
			value, err := compensateSteps[i].Compensate(values[i])

			if err != nil {
				return value
			}

			if verbose {
				fmt.Printf("undo step: %s", compensationStep.Name)
				fmt.Println()
				fmt.Println("output:")
				fmt.Println(value)
				fmt.Println()
			}
		}
	}

	return ""
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

func retry(step Step, f func() error) (err error) {
	attempts := step.Retries

	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

	}

	return fmt.Errorf("%s \"%s\" failed after (%d) attempts", err, step.Name, max(1, attempts))
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
