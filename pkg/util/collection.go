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

func Duplicates[T any, K comparable](values []T, fn func(v T) K) *K {
	seen := map[K]struct{}{}
	for _, v := range values {
		key := fn(v)
		_, ok := seen[key]
		if ok {
			return &key
		}
		seen[key] = struct{}{}
	}
	return nil
}

func ExistMap[T comparable, K comparable](values []T, fn func(v T) K) map[K]struct{} {
	results := map[K]struct{}{}
	for _, e := range values {
		results[fn(e)] = struct{}{}
	}
	return results
}
