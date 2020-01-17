package baur1

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/simplesurance/baur/digest"
)

type OutputType int

const (
	DockerOutput OutputType = iota
	FileOutput
)

type Output interface {
	fmt.Stringer

	Path() string
	Digest() (*digest.Digest, error)
	Exists() (bool, error)
	UploadDestination() *url.URL
	Type() OutputType
	Size() (int64, error)
	// UploadMethod
}

func OutputsFromTask(dockerClient DockerInfoClient, task *Task) ([]Output, error) {
	var result []Output

	// TODO create file outputs
	for _, dockerOutput := range task.Outputs.DockerImage {
		strURL := fmt.Sprintf("docker://%s:%s", dockerOutput.RegistryUpload.Repository, dockerOutput.RegistryUpload.Tag)
		url, err := url.Parse(strURL)
		if err != nil {
			return nil, err
		}

		d, err := NewOutputDockerImageFromIIDFile(
			dockerClient,
			filepath.Join(task.Directory, dockerOutput.IDFile),
			url,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, d)
	}

	return result, nil
}
