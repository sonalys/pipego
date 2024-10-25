package pp

import (
	"context"
	"slices"
)

// DivideSliceInSize receives a slice `s` and divide it into groups with `n` elements each,
// then it uses a step factory to generate steps for each group.
// `n` must be greater than 0 or it will panic.
func DivideSliceInSize[T any](s []T, n int, stepFactory func(T) StepFunc) (steps Steps) {
	for chunk := range slices.Chunk(s, n) {
		batch := make(Steps, 0, len(chunk))
		for _, v := range chunk {
			batch = append(batch, stepFactory(v))
		}
		steps = append(steps, batch.Group())
	}
	return steps
}

// divideSliceInGroups receive a slice `s` and breaks it into `n` sub-slices.
func divideSliceInGroups[T any](s []T, n int) [][]T {
	length := float64(len(s))
	if length == 0 {
		return nil
	}
	var out [][]T
	for segment := 0; segment < n; segment++ {
		startIndex := int(float64(segment) / float64(n) * length)
		endIndex := int(float64(segment+1) / float64(n) * length)
		if startIndex == endIndex {
			continue
		}
		out = append(out, s[startIndex:endIndex])
	}
	return out
}

// DivideSliceInGroups receives a slice `s` and divide it into `n` groups,
// then it uses a step factory to generate steps for each group.
func DivideSliceInGroups[T any](s []T, n int, stepFactory func(T) StepFunc) (steps Steps) {
	for _, chunk := range divideSliceInGroups(s, n) {
		batch := make(Steps, 0, len(chunk))
		for _, v := range chunk {
			batch = append(batch, stepFactory(v))
		}
		steps = append(steps, batch.Group())
	}
	return steps
}

// ForEach takes a slice `s` and a stepFactory, and creates a step for each element inside.
func ForEach[T any](s []T, stepFactory func(T, int) StepFunc) StepFunc {
	batch := Steps{}
	for i := range s {
		batch = append(batch, stepFactory(s[i], i))
	}
	return batch.Group()
}
