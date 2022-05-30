package util

func Contains[T comparable](arr []T, el T) bool {
	for _, x := range arr {
		if x == el {
			return true
		}
	}
	return false
}
