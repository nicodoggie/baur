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

func (o OutputType) String() string {
	switch o {
	case DockerOutput:
		return "docker"
	case FileOutput:
		return "file"
	default:
		return "invalid OutputType"
	}
}

type Output interface {
	fmt.Stringer

	Path() string
	Digest() (*digest.Digest, error)
	Exists() (bool, error)
	UploadDestination() *url.URL
	Type() OutputType
	Size() (int64, error)
}

func OutputsFromTask(dockerClient DockerInfoClient, task *Task) ([]Output, error) {
	var result []Output

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

	for _, fileOutput := range task.Outputs.File {
		// TODO: use pointers in the outputfile struct for filecopy and S3 instead of having to provide and use IsEmpty)
		if !fileOutput.S3Upload.IsEmpty() {
			var err error

			uploadDestination, err := url.Parse("s3://" + fileOutput.S3Upload.Bucket + "/" + fileOutput.S3Upload.DestFile)
			if err != nil {
				return nil, err
			}

			f := NewOutputFile(filepath.Join(task.Directory, fileOutput.Path), uploadDestination)
			result = append(result, f)

			continue
		}

		if !fileOutput.FileCopy.IsEmpty() {
			var err error

			uploadDestination, err := url.Parse("file://" + fileOutput.FileCopy.Path)
			if err != nil {
				return nil, err
			}

			f := NewOutputFile(filepath.Join(task.Directory, fileOutput.Path), uploadDestination)
			result = append(result, f)

			continue
		}

		return nil, fmt.Errorf("no upload method for output %q is specified", fileOutput.Path)
	}

	return result, nil
}
