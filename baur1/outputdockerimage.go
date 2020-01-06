package baur1

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/simplesurance/baur/fs"
)

// OutputDockerImage is a docker container artifact
type OutputDockerImage struct {
	localPath         string
	absPath           string
	imageIDFile       string
	uploadDestination *url.URL
}

// Exists returns true if the ImageIDFile exists
func (d *OutputDockerImage) Exists() bool {
	return fs.FileExists(d.imageIDFile)
}

// ImageID reads the image from ImageIDFile
func (d *OutputDockerImage) ImageID() (string, error) {
	id, err := fs.FileReadLine(d.imageIDFile)
	if err != nil {
		return "", err
	}

	if len(id) == 0 {
		return "", errors.New("file is empty")
	}

	return id, nil
}

// String returns LocalPath()
func (d *OutputDockerImage) String() string {
	return d.LocalPath()
}

// LocalPath returns the local path to the artifact
func (d *OutputDockerImage) LocalPath() string {
	return d.localPath
}

func (d *OutputDockerImage) Path() string {
	return d.absPath
}

// UploadDestination returns the upload destination
func (d *OutputDockerImage) UploadDestination() *url.URL {
	return d.uploadDestination
}

// Type returns "docker"
func (d *OutputDockerImage) Type() string {
	return "docker"
}
