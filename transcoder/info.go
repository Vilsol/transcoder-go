package transcoder

import (
	"encoding/json"
	"github.com/Vilsol/transcoder-go/models"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strings"
)

func ReadFileMetadata(file string) (*models.FileMetadata, error) {
	params := []string{"-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", file}

	log.Tracef("Executing ffprobe %s", strings.Join(params, " "))

	var outerErr error
	for i := 0; i < 3; i++ {
		c := exec.Command("ffprobe", params...)

		pipe, err := c.StdoutPipe()
		if err != nil {
			outerErr = errors.Wrap(err, "failed hooking ffprobe stdout")
			continue
		}

		err = c.Start()
		if err != nil {
			outerErr = errors.Wrap(err, "failed running ffprobe")
			continue
		}

		stdoutData, err := ioutil.ReadAll(pipe)
		if err != nil {
			outerErr = errors.Wrap(err, "failed reading ffprobe response")
			continue
		}

		err = c.Wait()
		if err != nil {
			outerErr = errors.Wrap(err, "ffprobe exited")
			continue
		}

		var metadata models.FileMetadata
		err = json.Unmarshal(stdoutData, &metadata)
		if err != nil {
			outerErr = errors.Wrap(err, "failed parsing ffprobe output")
			continue
		}

		return &metadata, nil
	}

	return nil, outerErr
}
