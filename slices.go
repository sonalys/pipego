package pipego

import (
	"context"

	"github.com/samber/lo"
)

// Group takes any amount of steps and return a single step that bounds them all.
func Group(steps ...StepFunc) StepFunc {
	return func(ctx context.Context) (err error) {
		return runSteps(ctx, steps...)
	}
}

// DivideSliceInSize receives a slice `s` and divide it into groups with `n` elements each,
// then it uses a step factory to generate steps for each group.
func DivideSliceInSize[T any](
	s []T, n int, stepFactory func(T) StepFunc) (steps []StepFunc) {
	for _, chunk := range lo.Chunk(s, n) {
		batch := make([]StepFunc, len(chunk))
		for _, v := range chunk {
			batch = append(batch, stepFactory(v))
		}
		steps = append(steps, Group(batch...))
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
func DivideSliceInGroups[T any](
	s []T, n int, stepFactory func(T) StepFunc) (steps []StepFunc) {
	for _, chunk := range divideSliceInGroups(s, n) {
		batch := make([]StepFunc, len(chunk))
		for _, v := range chunk {
			batch = append(batch, stepFactory(v))
		}
		steps = append(steps, Group(batch...))
	}
	return steps
}

// ForEach takes a slice `s` and a stepFactory, and creates a step for each element inside.
func ForEach[T any](s []T, stepFactory func(T, int) StepFunc) StepFunc {
	batch := []StepFunc{}
	for i := range s {
		batch = append(batch, stepFactory(s[i], i))
	}
	return Group(batch...)
}
