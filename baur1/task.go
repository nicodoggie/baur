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
	workingDir string
	Cfg        *cfg.Task

	appName string

	resolver *resolve.Resolver

	outputs []Output
}

// TODO: instead of making Cfg public, make it private and add a getter or do
// not make the cfg accessible and have getters for the required information
// that must be retrieved

func NewTask(appName string, cfg *cfg.Task, workingDir string) (*Task, error) {
	outputs, err := outputs(cfg, workingDir)
	if err != nil {
		return nil, err
	}

	return &Task{
		appName:    appName,
		workingDir: workingDir,
		Cfg:        cfg,
		outputs:    outputs,
	}, nil
}

// Name returns the name of the task
func (t *Task) Name() string {
	return t.Cfg.Name
}

func (t *Task) AppName() string {
	return t.appName
}

// ID returns <APP-NAME>.<TASK-NAME>
func (t *Task) ID() string {
	return fmt.Sprintf("%s.%s", t.appName, t.Cfg.Name)
}

func (t *Task) String() string {
	return t.ID()
}

func (t *Task) Directory() string {
	return t.workingDir
}

func (t *Task) Outputs() []Output {
	return t.outputs
}

func (t *Task) Command() string {
	return t.Cfg.Command
}

func outputs(cfg *cfg.Task, taskDir string) ([]Output, error) {
	var result []Output

	for _, outputFile := range cfg.Output.File {
		f := &OutputFile{
			localPath: outputFile.Path,
			absPath:   filepath.Join(taskDir, outputFile.Path),
		}

		// TODO: use pointers in the outputfile struct for filecopy and S3 instead of having to provide and use IsEmpty)

		if !outputFile.S3Upload.IsEmpty() {
			var err error

			f.uploadDestination, err = url.Parse("s3://" + outputFile.S3Upload.Bucket + "/" + outputFile.S3Upload.DestFile)
			if err != nil {
				return nil, err
			}

			continue
		}
		if !outputFile.FileCopy.IsEmpty() {
			var err error

			f.uploadDestination, err = url.Parse("file://" + outputFile.FileCopy.Path)
			if err != nil {
				return nil, err
			}
		}

		result = append(result, f)
	}

	for _, dockerOutput := range cfg.Output.DockerImage {
		strURL := fmt.Sprintf("docker://%s:%s", dockerOutput.RegistryUpload.Repository, dockerOutput.RegistryUpload.Tag)
		url, err := url.Parse(strURL)
		if err != nil {
			return nil, err
		}

		result = append(result, &OutputDockerImage{
			localPath:         dockerOutput.IDFile,
			absPath:           filepath.Join(taskDir, dockerOutput.IDFile),
			uploadDestination: url,
		})
	}

	return result, nil
}

func SortTasksByID(tasks []*Task) {
	sort.Slice(tasks, func(i int, j int) bool {
		return tasks[i].ID() < tasks[i].ID()
	})
}
