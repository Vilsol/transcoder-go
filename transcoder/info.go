package transcoder

import (
	"encoding/json"
	"github.com/Vilsol/transcoder-go/models"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strings"
)

func ReadFileMetadata(file string) *models.FileMetadata {
	params := []string{"-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", file}

	log.Tracef("Executing ffprobe %s", strings.Join(params, " "))

	c := exec.Command("ffprobe", params...)

	pipe, err := c.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed hooking ffprobe stdout: %s", err)
	}

	err = c.Start()
	if err != nil {
		log.Fatalf("Failed running ffprobe: %s", err)
	}

	stdoutData, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Fatalf("Failed reading ffprobe response: %s", err)
	}

	err = c.Wait()
	if err != nil {
		log.Fatalf("ffprobe exited: %s", err)
	}

	var metadata models.FileMetadata
	err = json.Unmarshal(stdoutData, &metadata)
	if err != nil {
		log.Fatalf("Failed parsing ffprobe output: %s", err)
	}

	return &metadata
}
