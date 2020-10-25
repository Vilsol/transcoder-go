package models

import (
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type FileMetadata struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	CodecName      string  `json:"codec_name"`
	CodecType      string  `json:"codec_type"`
	PixelFormat    *string `json:"pix_fmt"`
	Level          int     `json:"level"`
	ColorRange     *string `json:"color_range"`
	ColorSpace     *string `json:"color_space"`
	ColorTransfer  *string `json:"color_transfer"`
	ColorPrimaries *string `json:"color_primaries"`
	NumberFrames   string  `json:"nb_frames"`
	RFrameRate     *string `json:"r_frame_rate"`
	AvgFrameRate   *string `json:"avg_frame_rate"`
}

type Format struct {
	Filename   string `json:"filename"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
}

type ProgressReport struct {
	Frame     int
	FPS       float64
	Bitrate   float64
	TotalSize int
	Speed     float64
	Progress  string
}

type Result string

const (
	ResultKeepOriginal = Result("Kept original")
	ResultReplaced     = Result("Replaced with new")
	ResultError        = Result("Error")
	ResultSkipped      = Result("Skipped, kept original")
)

func (format Format) SizeInt() int64 {
	i, _ := strconv.Atoi(format.Size)
	return int64(i)
}

func (stream Stream) FrameRate() float64 {
	rate := ""

	if stream.RFrameRate != nil {
		rate = *stream.RFrameRate
	}

	if stream.AvgFrameRate != nil {
		rate = *stream.AvgFrameRate
	}

	if rate == "" {
		return 0
	}

	split := strings.Split(rate, "/")

	a, _ := strconv.Atoi(split[0])
	b, _ := strconv.Atoi(split[1])

	return float64(a) / float64(b)
}

func (report *ProgressReport) Log(filename string) {
	log.WithField("frame", report.Frame).
		WithField("fps", report.FPS).
		WithField("bitrate", report.Bitrate).
		WithField("total_size", report.TotalSize).
		WithField("speed", report.Speed).
		Infof("Progress: %s", filename)
}
