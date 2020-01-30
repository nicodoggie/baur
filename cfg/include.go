package cfg

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Include struct {
	Inputs  []*InputInclude
	Outputs []*OutputInclude
	Tasks   []*TaskInclude

	filePath string
}

// InputInclude is a reusable Input definition
type InputInclude struct {
	ID string `toml:"id" comment:"identifier of the include"`

	Files         FileInputs    `comment:"Inputs specified by file glob paths"`
	GitFiles      GitFileInputs `comment:"Inputs specified by path, matching only Git tracked files"`
	GolangSources GolangSources `comment:"Inputs specified by directories containing Golang applications"`
}

// OutputInclude is a reusable Output definition
type OutputInclude struct {
	ID string `toml:"id" comment:"identifier of the include"`

	DockerImage []*DockerImageOutput `comment:"Docker images that are produced by the [Task.command]"`
	File        []*FileOutput        `comment:"Files that are produces by the [Task.command]"`
}

// TaskInclude is a reusable Tasks definition
type TaskInclude struct {
	ID string `toml:"id" comment:"identifier of the include"`

	Name     string   `toml:"name" comment:"Identifies the task, currently the name must be 'build'."`
	Command  string   `toml:"command" comment:"Command that the task executes"`
	Includes []string `toml:"includes" comment:"IDs of input or output includes that the task inherits."`
	Input    *Input   `toml:"Input" comment:"Specification of task inputs like source files, Makefiles, etc"`
	Output   *Output  `toml:"Output" comment:"Specification of task outputs produced by the Task.command"`
}

// ExampleInclude returns an Include struct with exemplary values.
func ExampleInclude(id string) *Include {
	return &Include{
		Inputs: []*InputInclude{
			{
				ID: id + "_input",
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
			},
		},
		Outputs: []*OutputInclude{
			{
				ID: id + "_output",
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
			},
		},
		Tasks: []*TaskInclude{
			{
				ID:      id + "_task_cbuild",
				Name:    "cbuild",
				Command: "make",
				Input: &Input{
					GitFiles: GitFileInputs{
						Paths: []string{"*.c", "*.h", "Makefile"},
					},
				},
				Output: &Output{
					File: []*FileOutput{
						{
							Path: "a.out",
							FileCopy: FileCopy{
								Path: "/artifacts",
							},
						},
					},
				},
			},
		},
	}
}

// IncludeToFile serializes the Include struct to TOML and writes it to filepath.
func (incl *Include) IncludeToFile(filepath string) error {
	return toFile(incl, filepath, false)
}

// IncludeFromFile deserializes an Include struct from a file.
func IncludeFromFile(path string) (*Include, error) {
	config := Include{}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	config.filePath = path

	/*
		TODO: IS THIS NEEDED?

		if config.Output != nil {
			config.Output.removeEmptySections()
		}
	*/

	return &config, err
}

// Validate validates an Include configuration struct.
func (incl *Include) Validate() error {
	for _, in := range incl.Inputs {
		if err := in.Validate(); err != nil {
			if in.ID != "" {
				return NewFieldError(err, "Inputs", in.ID)
			}

			return NewFieldError(err, "Inputs")
		}
	}

	for _, out := range incl.Outputs {
		if err := out.Validate(); err != nil {
			if out.ID != "" {
				return NewFieldError(err, "Outputs", out.ID)
			}

			return NewFieldError(err, "Outputs")
		}
	}

	// TODO: we first have to merge the task itself with the includes that it references and then validate, otherwise validation will fail
	/*
		for _, tasks := range incl.Tasks {
				if err := tasks.Validate(); err != nil {
					if tasks.ID != "" {
						return NewFieldError(err, "Tasks", tasks.ID)
					}

					return NewFieldError(err, "Tasks")
				}

			if len(incl.Inputs) == 0 && len(incl.Outputs) == 0 && len(incl.Tasks) == 0 {
				return errors.New("the include does not contain any definition, either an Input, Output or Task must be defined")
			}
		}
	*/

	return nil
}

func (in *InputInclude) FileInputs() *FileInputs {
	return &in.Files
}

func (in *InputInclude) GitFileInputs() *GitFileInputs {
	return &in.GitFiles
}

func (in *InputInclude) GolangSourcesInputs() *GolangSources {
	return &in.GolangSources
}

func (in *InputInclude) Validate() error {
	if in.ID == "" {
		return NewFieldError(
			errors.New("can not be empty"),
			"id",
		)
	}

	if inputsIsEmpty(in) {
		return NewFieldError(
			errors.New("no input is defined"),
			"Input",
		)
	}

	if err := InputValidate(in); err != nil {
		return NewFieldError(err, "Input")
	}

	return nil
}

func (out *OutputInclude) DockerImageOutputs() []*DockerImageOutput {
	return out.DockerImage
}

func (out *OutputInclude) FileOutputs() []*FileOutput {
	return out.File
}

func (out *OutputInclude) Validate() error {
	if out.ID == "" {
		return NewFieldError(
			errors.New("can not be empty"),
			"id",
		)
	}

	if len(out.DockerImage) == 0 && len(out.File) == 0 {
		return NewFieldError(
			errors.New("no output is defined"),
			"Output",
		)
	}

	if err := OutputValidate(out); err != nil {
		return NewFieldError(err, "Output")
	}

	return nil
}

func (t *TaskInclude) GetCommand() string {
	return t.Command
}

func (t *TaskInclude) GetName() string {
	return t.Name
}

func (t *TaskInclude) GetIncludes() []string {
	return t.Includes
}

func (t *TaskInclude) GetInput() *Input {
	return t.Input
}

func (t *TaskInclude) GetOutput() *Output {
	return t.Output
}

func (t *TaskInclude) SetOutput(o *Output) {
	t.Output = o
}

func (t *TaskInclude) SetInput(in *Input) {
	t.Input = in
}

func (t *TaskInclude) Validate() error {
	if t.ID == "" {
		return NewFieldError(
			errors.New("can not be empty"),
			"id",
		)
	}

	if err := TaskValidate(t); err != nil {
		return err
	}

	return nil
}
