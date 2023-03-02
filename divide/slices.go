package divide

import (
	"github.com/sonalys/pipego"
)

// DivideSlice receives a slice with L elements, and divides it into N segments,
// then it runs through a provided stepFactory to generate steps for each batch.
func DivideSize[T any](slice []T, size int, stepFactory func([]T) pipego.StepFunc) []pipego.StepFunc {
	length := len(slice)
	if length == 0 {
		return []pipego.StepFunc{}
	}
	segments := length / size
	steps := make([]pipego.StepFunc, 0, segments)
	for startIndex := 0; startIndex < length; startIndex += size {
		if startIndex > length-1 {
			startIndex = length - 1
		}
		endIndex := startIndex + size
		if endIndex > length {
			endIndex = length
		}
		steps = append(steps, stepFactory(slice[startIndex:endIndex]))
	}
	return steps
}

// DivideSlice receives a slice with L elements, and divides it into N segments,
// then it runs through a provided stepFactory to generate steps for each batch.
func DivideSegments[T any](slice []T, segments int, stepFactory func([]T) pipego.StepFunc) []pipego.StepFunc {
	length := float64(len(slice))
	if length == 0 {
		return []pipego.StepFunc{}
	}
	steps := make([]pipego.StepFunc, 0, segments)
	for segment := 0; segment < segments; segment++ {
		startIndex := int(float64(segment) / float64(segments) * length)
		endIndex := int(float64(segment+1) / float64(segments) * length)
		if startIndex == endIndex {
			continue
		}
		steps = append(steps, stepFactory(slice[startIndex:endIndex]))
	}
	return steps
}
