package baur1

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/simplesurance/baur/fs"
)

// OutputDockerImage is a docker container artifact
type OutputDockerImage struct {
	imageIDFile       string
	name              string
	uploadDestination *url.URL
}

func NewOutputDockerImage(name, imageIDFilePath string, uploadDestination *url.URL) *OutputDockerImage {
	return &OutputDockerImage{
		name:              name,
		imageIDFile:       imageIDFilePath,
		uploadDestination: uploadDestination,
	}
}

// LocalPath reads the image ID from the imageIDFile and returns it.
func (d *OutputDockerImage) LocalPath() (string, error) {
	id, err := fs.FileReadLine(d.imageIDFile)
	if err != nil {
		return "", err
	}

	if len(id) == 0 {
		return "", errors.New("file is empty")
	}

	return id, nil
}

// String returns Name()
func (d *OutputDockerImage) String() string {
	return d.Name()
}

func (d *OutputDockerImage) Name() string {
	return d.name
}

// UploadDestination returns the upload destination
func (d *OutputDockerImage) UploadDestination() *url.URL {
	return d.uploadDestination
}

func (d *OutputDockerImage) Type() OutputType {
	return DockerOutput
}
