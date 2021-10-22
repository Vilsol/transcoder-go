package utils

import (
	"math"
)

func SkipConfidence(originalSize int, currentSize int, completion float64) float64 {
	expectedSize := float64(currentSize*100) / completion
	sizeDiff := ((expectedSize / float64(originalSize)) - 1) * 100
	if sizeDiff > 10 {
		return math.Log(sizeDiff) / math.Log(3-math.Log10(completion))
	}
	return 0
}
