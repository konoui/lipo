package util

func Map[T, V any](values []T, fn func(T) V) []V {
	results := make([]V, len(values), cap(values))
	for i, elm := range values {
		results[i] = fn(elm)
	}
	return results
}

func Filter[T any](values []T, fn func(T) bool) []T {
	results := make([]T, 0, len(values))
	for _, e := range values {
		if fn(e) {
			results = append(results, e)
		}
	}
	return results
}

func Contains[T comparable](values []T, target T) bool {
	for _, e := range values {
		if e == target {
			return true
		}
	}
	return false
}
