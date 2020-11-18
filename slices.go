package main

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func all(xxs [][]string, f func(xs []string) bool) bool {
	for _, xs := range xxs {
		if f(xs) == false {
			return false
		}
	}
	return true
}

func many(xxs [][]string, f func(xs []string) bool) bool {
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

func findFirst(xxs [][]string, f func(xs []string) bool) []string {
	for i, xs := range xxs {
		if f(xs) == true {
			return xxs[i]
		}
	}

	return nil
}

func empty(xs []string) bool {
	return len(xs) == 0
}

func nonEmpty(xs []string) bool {
	return len(xs) != 0
}
