package pipego

// DivideSliceInSize receives a slice `s` and divide it into groups with `n` elements each,
// then it uses a step factory to generate steps for each group.
func DivideSliceInSize[T any](s []T, n int, stepFactory func([]T) StepFunc) []StepFunc {
	length := len(s)
	if length == 0 {
		return []StepFunc{}
	}
	segments := length / n
	steps := make([]StepFunc, 0, segments)
	for startIndex := 0; startIndex < length; startIndex += n {
		if startIndex > length-1 {
			startIndex = length - 1
		}
		endIndex := startIndex + n
		if endIndex > length {
			endIndex = length
		}
		steps = append(steps, stepFactory(s[startIndex:endIndex]))
	}
	return steps
}

// DivideSliceInGroups receives a slice `s` and divide it into `n` groups,
// then it uses a step factory to generate steps for each group.
func DivideSliceInGroups[T any](s []T, n int, stepFactory func([]T) StepFunc) []StepFunc {
	length := float64(len(s))
	if length == 0 {
		return []StepFunc{}
	}
	steps := make([]StepFunc, 0, n)
	for segment := 0; segment < n; segment++ {
		startIndex := int(float64(segment) / float64(n) * length)
		endIndex := int(float64(segment+1) / float64(n) * length)
		if startIndex == endIndex {
			continue
		}
		steps = append(steps, stepFactory(s[startIndex:endIndex]))
	}
	return steps
}
