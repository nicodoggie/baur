package cfg

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pelletier/go-toml"
)

type Tasks []*Task

// App stores an application configuration.
type App struct {
	Name     string   `toml:"name" comment:"Name of the application"`
	Includes []string `toml:"includes" comment:"IDs of Tasks includes that the task inherits."`
	Tasks    Tasks    `toml:"Task"`
}

// Task is a task section
type Task struct {
	Name     string   `toml:"name" comment:"Identifies the task, currently the name must be 'build'."`
	Command  string   `toml:"command" comment:"Command that the task executes"`
	Includes []string `toml:"includes" comment:"IDs of input or output includes that the task inherits."`
	Input    *Input   `toml:"Input" comment:"Specification of task inputs like source files, Makefiles, etc"`
	Output   *Output  `toml:"Output" comment:"Specification of task outputs produced by the Task.command"`
}

// Input contains information about task inputs
type Input struct {
	Files         FileInputs    `comment:"Inputs specified by file glob paths"`
	GitFiles      GitFileInputs `comment:"Inputs specified by path, matching only Git tracked files"`
	GolangSources GolangSources `comment:"Inputs specified by directories containing Golang applications"`
}

// GolangSources specifies inputs for Golang Applications
type GolangSources struct {
	Environment []string `toml:"environment" comment:"Environment to use when discovering Golang source files\n This can be environment variables understood by the Golang tools, like GOPATH, GOFLAGS, etc.\n If empty the default Go environment is used.\n Valid variables: $ROOT " commented:"true"`
	Paths       []string `toml:"paths" comment:"Paths to directories containing Golang source files.\n All source files including imported packages are discovered,\n files from Go's stdlib package and testfiles are ignored." commented:"true"`
}

// FileInputs describes a file source
type FileInputs struct {
	Paths []string `toml:"paths" commented:"true" comment:"Relative path to source files,\n supports Golang's Glob syntax (https://golang.org/pkg/path/filepath/#Match) and\n ** to match files recursively\n Valid variables: $ROOT"`
}

// GitFileInputs describes source files that are in the git repository by git
// pathnames
type GitFileInputs struct {
	Paths []string `toml:"paths" commented:"true" comment:"Relative paths to source files.\n Only files tracked by Git that are not in the .gitignore file are matched.\n The same patterns that git ls-files supports can be used.\n Valid variables: $ROOT"`
}

// Output is the tasks output section
type Output struct {
	DockerImage []*DockerImageOutput `comment:"Docker images that are produced by the [Task.command]"`
	File        []*FileOutput        `comment:"Files that are produces by the [Task.command]"`
}

// FileOutput describes where a file artifact should be uploaded to
type FileOutput struct {
	Path     string   `toml:"path" comment:"Path relative to the application directory, valid variables: $APPNAME" commented:"true"`
	FileCopy FileCopy `comment:"Copy the file to a local directory"`
	S3Upload S3Upload `comment:"Upload the file to S3"`
}

// FileCopy describes where a file artifact should be copied to
type FileCopy struct {
	Path string `toml:"path" comment:"Destination directory" commented:"true"`
}

// DockerImageRegistryUpload holds information about where the docker image
// should be uploaded to
type DockerImageRegistryUpload struct {
	Repository string `toml:"repository" comment:"Repository path, format: [<server[:port]>/]<owner>/<repository>:<tag>, valid variables: $APPNAME" commented:"true"`
	Tag        string `toml:"tag" comment:"Tag that is applied to the image, valid variables: $APPNAME, $UUID, $GITCOMMIT" commented:"true"`
}

// S3Upload contains S3 upload information
type S3Upload struct {
	Bucket   string `toml:"bucket" comment:"Bucket name, valid variables: $APPNAME" commented:"true"`
	DestFile string `toml:"dest_file" comment:"Remote File Name, valid variables: $APPNAME, $UUID, $GITCOMMIT" commented:"true"`
}

// DockerImageOutput describes where a docker container is uploaded to
type DockerImageOutput struct {
	IDFile         string                    `toml:"idfile" comment:"Path to a file that is created by [Task.Command] and contains the image ID of the produced image (docker build --iidfile), valid variables: $APPNAME" commented:"true"`
	RegistryUpload DockerImageRegistryUpload `comment:"Registry repository the image is uploaded to"`
}

func exampleInput() *Input {
	return &Input{
		Files: FileInputs{
			Paths: []string{"dbmigrations/*.sql"},
		},
		GitFiles: GitFileInputs{
			Paths: []string{"Makefile"},
		},
		GolangSources: GolangSources{
			Paths:       []string{"."},
			Environment: []string{"GOFLAGS=-mod=vendor", "GO111MODULE=on"},
		},
	}
}

func exampleOutput() *Output {
	return &Output{
		File: []*FileOutput{
			{
				Path: "dist/$APPNAME.tar.xz",
				S3Upload: S3Upload{
					Bucket:   "go-artifacts/",
					DestFile: "$APPNAME-$GITCOMMIT.tar.xz",
				},
				FileCopy: FileCopy{
					Path: "/mnt/fileserver/build_artifacts/$APPNAME-$GITCOMMIT.tar.xz",
				},
			},
		},
		DockerImage: []*DockerImageOutput{
			{
				IDFile: fmt.Sprintf("$APPNAME-container.id"),
				RegistryUpload: DockerImageRegistryUpload{
					Repository: "my-company/$APPNAME",
					Tag:        "$GITCOMMIT",
				},
			},
		},
	}
}

// ExampleApp returns an exemplary app cfg struct with the name set to the given value
func ExampleApp(name string) *App {
	return &App{
		Name: name,

		Tasks: []*Task{
			&Task{
				Name:    "build",
				Command: "make dist",
				Input:   exampleInput(),
				Output:  exampleOutput(),
			},
		},
	}
}

// AppFromFile reads a application configuration file and returns it.
// If the buildCmd is not set in the App configuration it's set to
// defaultBuild.Command
func AppFromFile(path string) (*App, error) {
	config := App{}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}

// ToFile writes an exemplary Application configuration file to
// filepath. The name setting is set to appName
func (a *App) ToFile(filepath string) error {
	return toFile(a, filepath, false)
}

// Validate validates a App configuration
func (a *App) Validate() error {
	if len(a.Name) == 0 {
		return &ValidationError{
			ElementPath: []string{"name"},
			Message:     "can not be empty",
		}
	}

	if err := a.Tasks.Validate(); err != nil {
		return PrependValidationErrorPath(err, "Tasks")
	}

	return nil
}

func (tasks Tasks) Validate() error {
	if len(tasks) != 1 {
		return &ValidationError{
			Message: fmt.Sprintf("must contain exactly 1 Task definition, has %d", len(tasks)),
		}
	}

	duplMap := make(map[string]struct{}, len(tasks))

	for _, task := range tasks {
		_, exist := duplMap[task.Name]
		if exist {
			return &ValidationError{
				ElementPath: []string{"Task", "name"},
				Message:     fmt.Sprintf("multiple tasks with name '%s' exist, task names must be unique", task.Name),
			}
		}
		duplMap[task.Name] = struct{}{}

		err := task.Validate()
		if err != nil {
			if task.Name == "" {
				return PrependValidationErrorPath(err, "Task")
			}

			return PrependValidationErrorPath(err, fmt.Sprintf("Task(name: %s)", task.Name))
		}
	}

	return nil
}

// Validate validates the task section
func (t *Task) Validate() error {
	if len(t.Command) == 0 {
		return &ValidationError{
			ElementPath: []string{"command"},
			Message:     fmt.Sprintf("can not be empty"),
		}
	}

	// TODO: change it to check for an invalid name when we support multiple tasks
	if t.Name != "build" {
		return &ValidationError{
			ElementPath: []string{"name"},
			Message:     "name must be 'build'",
		}
	}

	if t.Input == nil {
		return &ValidationError{
			ElementPath: []string{"Input"},
			Message:     "section is empty",
		}
	}

	if err := t.Input.Validate(); err != nil {
		return PrependValidationErrorPath(err, "Input")
	}

	if t.Output == nil {
		return &ValidationError{
			ElementPath: []string{"Output"},
			Message:     "section is empty",
		}
	}

	if err := t.Output.Validate(); err != nil {
		return PrependValidationErrorPath(err, "Output")
	}

	return nil
}

func (t *Task) Merge(includeDB *IncludeDB) error {
	for _, includeID := range t.Includes {
		if include, exist := includeDB.Inputs[includeID]; exist {
			t.Input.Merge(include.Input)
			continue
		}

		if include, exist := includeDB.Outputs[includeID]; exist {
			t.Output.Merge(include.Output)
			continue
		}

		return fmt.Errorf("could not find include with id '%s'", includeID)
	}

	return nil

}

// Merge for each ID in the Includes slice a TasksInclude in the includedb is looked up.
// The tasks of the found TasksInclude are appended to the Apps Tasks slice.
func (a *App) Merge(includedb *IncludeDB) error {
	for _, includeID := range a.Includes {
		include, exist := includedb.Tasks[includeID]
		if !exist {
			return fmt.Errorf("could not find include with id '%s'", includeID)
		}

		a.Tasks = append(a.Tasks, include.Tasks...)

		// TODO: store the repository relative cfg path in the include somehow
		//task.Input.Files.Paths = append(task.Input.Files.Paths, include.RelCfgPath)
	}

	for _, task := range a.Tasks {
		if err := task.Merge(includedb); err != nil {
			return err
		}
	}

	return nil
}

// Merge merges the Input with another one.
func (i *Input) Merge(other *Input) {
	i.Files.Merge(&other.Files)
	i.GitFiles.Merge(&other.GitFiles)
	i.GolangSources.Merge(&other.GolangSources)
}

// Validate validates the Input section
func (i *Input) Validate() error {
	if err := i.Files.Validate(); err != nil {
		return PrependValidationErrorPath(err, "Files")
	}

	if err := i.GolangSources.Validate(); err != nil {
		return PrependValidationErrorPath(err, "GolangSources")
	}

	// TODO: add validation for gitfiles section

	return nil
}

// Validate validates the GolangSources section
func (g *GolangSources) Validate() error {
	if len(g.Environment) != 0 && len(g.Paths) == 0 {
		return &ValidationError{
			ElementPath: []string{"paths"},
			Message:     "must be set if environment is set",
		}
	}

	for _, p := range g.Paths {
		if len(p) == 0 {
			return &ValidationError{
				ElementPath: []string{"paths"},
				Message:     "empty string is an invalid path",
			}
		}
	}

	return nil
}

// Merge merges the two GolangSources structs
func (g *GolangSources) Merge(other *GolangSources) {
	g.Paths = append(g.Paths, other.Paths...)
	g.Environment = append(g.Environment, other.Environment...)
}

// Merge merges the two Output structs
func (o *Output) Merge(other *Output) {
	o.DockerImage = append(o.DockerImage, other.DockerImage...)
	o.File = append(o.File, other.File...)
}

// Validate validates the Output section
func (o *Output) Validate() error {
	for _, f := range o.File {
		if err := f.Validate(); err != nil {
			return PrependValidationErrorPath(err, "File")
		}
	}

	for _, d := range o.DockerImage {
		if err := d.Validate(); err != nil {
			return PrependValidationErrorPath(err, "DockerImage")
		}
	}

	return nil
}

// IsEmpty returns true if FileCopy is empty
func (f *FileCopy) IsEmpty() bool {
	return len(f.Path) == 0
}

// IsEmpty returns true if S3Upload is empty
func (s *S3Upload) IsEmpty() bool {
	return len(s.Bucket) == 0 && len(s.DestFile) == 0
}

// Validate validates a [[Task.Output.File]] section
func (f *FileOutput) Validate() error {
	if len(f.Path) == 0 {
		return &ValidationError{
			ElementPath: []string{"path"},
			Message:     "can not be empty",
		}
	}

	return f.S3Upload.Validate()
}

//IsEmpty returns true if the struct is empty
func (d *DockerImageRegistryUpload) IsEmpty() bool {
	return len(d.Repository) == 0 && len(d.Tag) == 0
}

// Validate validates a [[Task.Output.File]] section
func (s *S3Upload) Validate() error {
	if s.IsEmpty() {
		return nil
	}

	if len(s.DestFile) == 0 {
		return &ValidationError{
			ElementPath: []string{"destfile"},
			Message:     "can not be empty",
		}
	}

	if len(s.Bucket) == 0 {
		return &ValidationError{
			ElementPath: []string{"bucket"},
			Message:     "can not be empty",
		}
	}

	return nil
}

// Validate validates its content
func (d *DockerImageOutput) Validate() error {
	if len(d.IDFile) == 0 {
		return &ValidationError{
			ElementPath: []string{"idfile"},
			Message:     "can not be empty",
		}
	}

	if err := d.RegistryUpload.Validate(); err != nil {
		return PrependValidationErrorPath(err, "RegistryUpload")
	}

	return nil
}

// Validate validates its content
func (d *DockerImageRegistryUpload) Validate() error {
	if len(d.Repository) == 0 {
		return &ValidationError{
			ElementPath: []string{"repository"},
			Message:     "can not be empty",
		}
	}

	if len(d.Tag) == 0 {
		return &ValidationError{
			ElementPath: []string{"tag"},
			Message:     "can not be empty",
		}
	}

	return nil
}

// Merge merges 2 FileInputs structs
func (f *FileInputs) Merge(other *FileInputs) {
	f.Paths = append(f.Paths, other.Paths...)
}

// Validate validates a [[Sources.Files]] section
func (f *FileInputs) Validate() error {
	for _, path := range f.Paths {
		if len(path) == 0 {
			return &ValidationError{
				ElementPath: []string{"path"},
				Message:     "can not be empty",
			}

		}
		if strings.Count(path, "**") > 1 {
			return &ValidationError{
				ElementPath: []string{"path"},
				Message:     "'**' can only appear one time in a path",
			}
		}
	}

	return nil
}

// Merge merges two GitFileInputs structs
func (g *GitFileInputs) Merge(other *GitFileInputs) {
	g.Paths = append(g.Paths, other.Paths...)
}
