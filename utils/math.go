package utils

import (
	"github.com/Vilsol/transcoder-go/models"
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

func SkipConfidenceMeta(originalMetadata *models.FileMetadata, frame int, size int) float64 {
	complete := (float64(frame) / float64(originalMetadata.Frames())) * 100

	if (float64(frame)/float64(originalMetadata.Frames()))*100 > 25 {
		return SkipConfidence(int(originalMetadata.Format.SizeInt()), size, complete)
	}

	return 0
}
