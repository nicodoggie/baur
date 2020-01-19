package baur1

import (
	"fmt"
	"net/url"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/fs"
)

type DockerInfoClient interface {
	Size(imageID string) (int64, error)
	Exists(imageID string) (bool, error)
}

// OutputDockerImage is a docker container artifact
type OutputDockerImage struct {
	imageID           string
	uploadDestination *url.URL
	dockerClient      DockerInfoClient
	digest            *digest.Digest
}

func NewOutputDockerImageFromIIDFile(dockerClient DockerInfoClient, iidfile string, uploadDestination *url.URL) (*OutputDockerImage, error) {
	id, err := fs.FileReadLine(iidfile)
	if err != nil {
		return nil, fmt.Errorf("reading %s failed: %w", iidfile, err)
	}

	digest, err := digest.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("image id %q read from %q has an invalid format: %w", id, iidfile, err)
	}

	return &OutputDockerImage{
		dockerClient:      dockerClient,
		imageID:           id,
		uploadDestination: uploadDestination,
		digest:            digest,
	}, nil
}

func (d *OutputDockerImage) String() string {
	return fmt.Sprintf("docker image: %s", d.imageID)
}

func (d *OutputDockerImage) Path() string {
	return d.imageID
}

func (d *OutputDockerImage) UploadDestination() *url.URL {
	return d.uploadDestination
}

func (d *OutputDockerImage) Type() OutputType {
	return DockerOutput
}

func (d *OutputDockerImage) Exists() (bool, error) {
	return d.dockerClient.Exists(d.imageID)
}

// Digest returns the imageID as a digest object. The method always returns a nil error.
func (d *OutputDockerImage) Digest() (*digest.Digest, error) {
	return d.digest, nil
}

func (d *OutputDockerImage) Size() (int64, error) {
	return d.dockerClient.Size(d.imageID)
}
