package saga

type Saga struct {
	steps []Step
}

type Step struct {
	Name string
	Run  func() error
	Undo func() error
}

func New(steps []Step) Saga {
	return Saga{steps: steps}
}

func (s *Saga) Run() (errors []string) {
	for i, step := range s.steps {

		if err := step.Run(); err != nil {
			errors = append(errors, err.Error())
			if latestError := undo(reverseSteps(s.steps[0:i])); latestError != nil {
				return append(errors, latestError.Error())
			} else {
				return errors
			}
		}

	}

	return errors
}

func undo(undoSteps []Step) (err error) {
	for _, undoStep := range undoSteps {
		if undoStep.Undo != nil {
			if err = undoStep.Undo(); err != nil {
				return
			}
		}
	}

	return
}

func reverseSteps(steps []Step) []Step {
	for i, j := 0, len(steps)-1; i < j; i, j = i+1, j-1 {
		steps[i], steps[j] = steps[j], steps[i]
	}
	return steps
}
