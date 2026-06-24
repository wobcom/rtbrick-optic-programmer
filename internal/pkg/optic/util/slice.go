package util

func CopySlice[T any](old []T) []T {
	newSlice := make([]T, len(old))
	copy(newSlice, old)
	return newSlice
}
