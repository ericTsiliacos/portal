package slices

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func All(xxs [][]string, f func(xs []string) bool) bool {
	for _, xs := range xxs {
		if f(xs) == false {
			return false
		}
	}
	return true
}

func Many(xxs [][]string, f func(xs []string) bool) bool {
	seen := false
	for _, xs := range xxs {
		if seen && f(xs) {
			return true
		} else if f(xs) {
			seen = true
		} else {
			continue
		}
	}
	return false
}

func FindFirst(xxs [][]string, f func(xs []string) bool) []string {
	for i, xs := range xxs {
		if f(xs) == true {
			return xxs[i]
		}
	}

	return nil
}

func Empty(xs []string) bool {
	return len(xs) == 0
}

func NonEmpty(xs []string) bool {
	return len(xs) != 0
}
