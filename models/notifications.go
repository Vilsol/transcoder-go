package models

import "time"

type NotificationData struct {
	Started time.Time

	Filename       string
	OriginalFrames int
	OriginalSize   int

	CurrentFrame int
	CurrentSize  int
	FPS          float64
	Bitrate      float64
	Speed        float64
}
