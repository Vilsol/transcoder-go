package notifications

import (
	"github.com/Vilsol/transcoder-go/models"
	"path/filepath"
	"strconv"
	"time"
)

type Initialize func()
type Start func(*models.NotificationData)
type ProgressStatus func(*models.NotificationData)
type End func(*models.NotificationData, models.Result)

var initialize []Initialize
var start []Start
var progressStatus []ProgressStatus
var end []End

var started time.Time
var currentFileMetadata *models.FileMetadata

func InitializeNotifications() {
	for _, f := range initialize {
		f()
	}
}

func NotifyStart(metadata *models.FileMetadata) {
	currentFileMetadata = metadata
	started = time.Now()

	notificationData := generateUpdatedNotificationData(nil)

	for _, f := range start {
		f(notificationData)
	}
}

func NotifyProgressStatus(report *models.ProgressReport) {
	notificationData := generateUpdatedNotificationData(report)
	for _, f := range progressStatus {
		f(notificationData)
	}
}

func NotifyEnd(finalMeta *models.FileMetadata, lastReport *models.ProgressReport, result models.Result) {
	notificationData := generateUpdatedNotificationData(lastReport)

	if finalMeta != nil {
		notificationData.CurrentSize, _ = strconv.Atoi(finalMeta.Format.Size)
		framerate := float64(0)

		for _, stream := range finalMeta.Streams {
			if stream.CodecType == "video" {
				notificationData.CurrentFrame, _ = strconv.Atoi(stream.NumberFrames)
				framerate = stream.FrameRate()
				break
			}
		}

		if notificationData.CurrentFrame == 0 && framerate > 0 {
			duration, _ := strconv.ParseFloat(finalMeta.Format.Duration, 64)
			notificationData.CurrentFrame = int(framerate * duration)
		}
	}

	for _, f := range end {
		f(notificationData, result)
	}
}

func generateUpdatedNotificationData(report *models.ProgressReport) *models.NotificationData {
	data := models.NotificationData{
		Started:  started,
		Filename: filepath.Base(currentFileMetadata.Format.Filename),
	}

	data.OriginalSize, _ = strconv.Atoi(currentFileMetadata.Format.Size)
	framerate := float64(0)

	for _, stream := range currentFileMetadata.Streams {
		if stream.CodecType == "video" {
			data.OriginalFrames, _ = strconv.Atoi(stream.NumberFrames)
			framerate = stream.FrameRate()
			break
		}
	}

	if data.OriginalFrames == 0 && framerate > 0 {
		duration, _ := strconv.ParseFloat(currentFileMetadata.Format.Duration, 64)
		data.OriginalFrames = int(framerate * duration)
	}

	if report != nil {
		data.Speed = report.Speed
		data.Bitrate = report.Bitrate
		data.FPS = report.FPS
		data.CurrentFrame = report.Frame
		data.CurrentSize = report.TotalSize
	}

	return &data
}
