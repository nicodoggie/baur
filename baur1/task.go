package baur1

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"

	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/resolve"
)

type Task struct {
	Directory string
	resolver  *resolve.Resolver

	AppName          string
	Command          string
	Name             string
	Outputs          []Output
	UnresolvedInputs *cfg.Input
}

func NewTask(appName string, cfg *cfg.Task, workingDir string) (*Task, error) {
	outputs, err := outputs(cfg, workingDir)
	if err != nil {
		return nil, err
	}

	return &Task{
		Directory:        workingDir,
		Outputs:          outputs,
		Command:          cfg.Command,
		Name:             cfg.Name,
		AppName:          appName,
		UnresolvedInputs: cfg.Input,
	}, nil
}

// ID returns <APP-NAME>.<TASK-NAME>
func (t *Task) ID() string {
	return fmt.Sprintf("%s.%s", t.AppName, t.Name)
}

func (t *Task) String() string {
	return t.ID()
}

func outputs(cfg *cfg.Task, taskDir string) ([]Output, error) {
	var result []Output

	for _, outputFile := range cfg.Output.File {
		// TODO: use pointers in the outputfile struct for filecopy and S3 instead of having to provide and use IsEmpty)
		if !outputFile.S3Upload.IsEmpty() {
			var err error

			uploadDestination, err := url.Parse("s3://" + outputFile.S3Upload.Bucket + "/" + outputFile.S3Upload.DestFile)
			if err != nil {
				return nil, err
			}

			f := NewOutputFile(outputFile.Path, filepath.Join(taskDir, outputFile.Path), uploadDestination)
			result = append(result, f)

			continue
		}

		if !outputFile.FileCopy.IsEmpty() {
			var err error

			uploadDestination, err := url.Parse("file://" + outputFile.FileCopy.Path)
			if err != nil {
				return nil, err
			}

			f := NewOutputFile(outputFile.Path, filepath.Join(taskDir, outputFile.Path), uploadDestination)
			result = append(result, f)

			continue
		}

		return nil, fmt.Errorf("no upload method for output %q is specified", outputFile.Path)
	}

	for _, dockerOutput := range cfg.Output.DockerImage {
		strURL := fmt.Sprintf("docker://%s:%s", dockerOutput.RegistryUpload.Repository, dockerOutput.RegistryUpload.Tag)

		url, err := url.Parse(strURL)
		if err != nil {
			return nil, err
		}

		d := NewOutputDockerImage(
			fmt.Sprintf("%s:%s", dockerOutput.RegistryUpload.Repository, dockerOutput.RegistryUpload.Tag),
			filepath.Join(taskDir, dockerOutput.IDFile),
			url,
		)

		result = append(result, d)
	}

	return result, nil
}

func SortTasksByID(tasks []*Task) {
	sort.Slice(tasks, func(i int, j int) bool {
		return tasks[i].ID() < tasks[i].ID()
	})
}
