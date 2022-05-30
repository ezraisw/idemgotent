package util

func MCopyFilter[K comparable, V any](src map[K]V, dst map[K]V, filterFn func(K, V) bool) {
	for name, value := range src {
		if filterFn != nil && !filterFn(name, value) {
			continue
		}
		dst[name] = value
	}
}

func Contains[T comparable](arr []T, el T) bool {
	for _, x := range arr {
		if x == el {
			return true
		}
	}
	return false
}
