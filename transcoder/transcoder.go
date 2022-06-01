package transcoder

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Vilsol/transcoder-go/models"
	"github.com/Vilsol/transcoder-go/notifications"
	"github.com/Vilsol/transcoder-go/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var lastReport *models.ProgressReport

func BuildFlags(fileName string, tempFileName string, metadata *models.FileMetadata) []string {
	finalFlags := make([]string, 0)

	if viper.GetBool("nice") && runtime.GOOS == "linux" {
		finalFlags = append(finalFlags, "ffmpeg")
	}

	// The input file
	finalFlags = append(finalFlags, "-y", "-i", fileName)

	if !viper.GetBool("stderr") {
		// Add quiet flag
		finalFlags = append(finalFlags, "-v", "quiet")
	}

	// Mandatory flags
	finalFlags = append(finalFlags, "-c", "copy", "-c:s", "srt", "-f", "matroska", "-progress", "-",)

	// Configurable flags
	finalFlags = append(finalFlags, strings.Split(viper.GetString("flags"), " ")...)

	// Add flags from original
	if metadata != nil {
		for _, stream := range metadata.Streams {
			if stream.CodecType == "video" {
				if stream.ColorPrimaries != nil {
					finalFlags = append(finalFlags, "-color_primaries", *stream.ColorPrimaries)
				}
				if stream.ColorRange != nil {
					finalFlags = append(finalFlags, "-color_range", *stream.ColorRange)
				}
				if stream.ColorSpace != nil {
					finalFlags = append(finalFlags, "-colorspace", *stream.ColorSpace)
				}
				if stream.ColorTransfer != nil {
					finalFlags = append(finalFlags, "-color_trc", *stream.ColorTransfer)
				}
				if stream.PixelFormat != nil {
					finalFlags = append(finalFlags, "-pix_fmt", *stream.PixelFormat)
				}
				break
			}
		}
	}

	// The output file
	finalFlags = append(finalFlags, tempFileName)

	return finalFlags
}

func TranscodeFile(fileName string, tempFileName string, metadata *models.FileMetadata, skip chan bool) (bool, *models.ProgressReport, bool) {
	flags := BuildFlags(fileName, tempFileName, metadata)

	notifications.NotifyStart(metadata)

	log.Tracef("Executing ffmpeg %s", strings.Join(flags, " "))

	var c *exec.Cmd
	if viper.GetBool("nice") && runtime.GOOS == "linux" {
		c = exec.Command("nice", flags...)
	} else {
		c = exec.Command("ffmpeg", flags...)
	}

	done := make(chan bool, 2)
	stopTranscoder := make(chan bool, 2)

	HookTermination(c, stopTranscoder, done, tempFileName)

	outPipe, err := c.StdoutPipe()
	defer outPipe.Close()
	if err != nil {
		log.Fatal(err)
	}

	errPipe, err := c.StderrPipe()
	defer errPipe.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = c.Start()
	if err != nil {
		log.Fatal(err)
	}

	if viper.GetBool("stderr") {
		go ReadError(errPipe)
	}

	go ReadOut(outPipe, fileName, metadata, stopTranscoder)

	skipping := false
	go func() {
		skipping = <-skip
		stopTranscoder <- true
	}()

	err = c.Wait()

	if err != nil {
		log.Errorf("ffmpeg: %s", err)
	}

	stopTranscoder <- false

	return <-done, lastReport, skipping
}

func ReadOut(pipe io.ReadCloser, filename string, metadata *models.FileMetadata, stopTranscoder chan bool) {
	lastLog := int64(0)
	lines := make([]string, 0)
	line := make([]byte, 0)
	for {
		buffer := make([]byte, 1)
		readCount, err := pipe.Read(buffer)

		if readCount == 0 {
			break
		}

		if buffer[0] != '\n' {
			line = append(line, buffer[0])
		} else {
			lines = append(lines, string(line))
			line = make([]byte, 0)

			// TODO Progress report based on value detection
			if len(lines) == 12 {
				report := OutputToReport(lines)
				lastReport = report

				if viper.GetBool("early-exit") && viper.GetBool("keep-old") {
					if int64(report.TotalSize) > metadata.Format.SizeInt() {
						stopTranscoder <- true
						return
					}

					if utils.SkipConfidenceMeta(metadata, report.Frame, report.TotalSize) > viper.GetFloat64("skip-confidence") {
						stopTranscoder <- true
						return
					}
				}

				notifications.NotifyProgressStatus(report)

				if time.Now().Unix()-lastLog > int64(viper.GetInt("interval")) {
					report.Log(filename)
					lastLog = time.Now().Unix()
				}

				lines = make([]string, 0)
			}
		}

		if err != nil && err != io.EOF && err != os.ErrClosed && !strings.HasSuffix(err.Error(), "file already closed") {
			log.Errorf("Error reading stdout: %s", err)
			return
		}
	}
}

func ReadError(pipe io.ReadCloser) {
	for {
		buffer := make([]byte, 1)
		readCount, err := pipe.Read(buffer)

		if readCount == 0 {
			break
		}

		if err != nil && err != io.EOF && err != os.ErrClosed && !strings.HasSuffix(err.Error(), "file already closed") {
			log.Errorf("Error reading stderr: %s", err)
			return
		}

		_, err = os.Stderr.Write(buffer)

		if err != nil {
			log.Errorf("Error writing to stderr: %s", err)
			return
		}
	}
}

var flatParseRegex = regexp.MustCompile(`\s*(-?[0-9.]+).*`)

func OutputToReport(lines []string) *models.ProgressReport {
	report := models.ProgressReport{}

	for _, line := range lines {
		split := strings.Split(line, "=")
		switch split[0] {
		case "frame":
			report.Frame, _ = strconv.Atoi(split[1])
			break
		case "fps":
			report.FPS, _ = strconv.ParseFloat(split[1], 64)
			break
		case "bitrate":
			matches := flatParseRegex.FindAllStringSubmatch(split[1], -1)
			if len(matches) > 0 {
				report.Bitrate, _ = strconv.ParseFloat(matches[0][1], 64)
			}
			break
		case "total_size":
			report.TotalSize, _ = strconv.Atoi(split[1])
			break
		case "speed":
			matches := flatParseRegex.FindAllStringSubmatch(split[1], -1)
			if len(matches) > 0 {
				report.Speed, _ = strconv.ParseFloat(matches[0][1], 64)
			}
			break
		case "progress":
			report.Progress = split[1]
			break
		}
	}

	return &report
}

func HookTermination(c *exec.Cmd, stopTranscoder chan bool, done chan bool, tempFileName string) {
	go func() {
		toTerminate := <-stopTranscoder

		if toTerminate {
			err := c.Process.Kill()

			if err != nil {
				log.Errorf("Error killing process: %s", err)
			}

			_, err = c.Process.Wait()

			if err != nil {
				log.Errorf("Error waiting for process exit: %s", err)
			}

			err = os.Remove(tempFileName)

			if err != nil {
				log.Errorf("Error deleting file %s: %s", tempFileName, err)
			}

			log.Warningf("ffmpeg killed")
		}

		done <- toTerminate
	}()

	terminate := make(chan os.Signal)

	go func() {
		toTerminate := <-terminate
		if toTerminate != nil {
			stopTranscoder <- true
		}
		signal.Stop(terminate)
	}()

	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
}
