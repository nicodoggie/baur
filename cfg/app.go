package cfg

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

/* TODO:
- Update field documentation to list correctly which ones supporting which variables
- Ensure we do not call Validate or Merge on Nil structs
*/

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
	Environment []string `toml:"environment" comment:"Environment to use when discovering Golang source files\n This can be environment variables understood by the Golang tools, like GOPATH, GOFLAGS, etc.\n If empty the default Go environment is used.\n Valid variables: $ROOT" commented:"true"`
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

// IsEmpty returns true if FileCopy is empty
func (f *FileCopy) IsEmpty() bool {
	return len(f.Path) == 0
}

// IsEmpty returns true if S3Upload is empty
func (s *S3Upload) IsEmpty() bool {
	return len(s.Bucket) == 0 && len(s.DestFile) == 0
}

//IsEmpty returns true if the struct is empty
func (d *DockerImageRegistryUpload) IsEmpty() bool {
	return len(d.Repository) == 0 && len(d.Tag) == 0
}
