package cfg

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Include struct {
	Inputs  []*InputInclude
	Outputs []*OutputInclude
	Tasks   []*TasksInclude

	filePath string
}

// InputInclude is a reusable Input definition
type InputInclude struct {
	ID    string `toml:"id" comment:"identifier for the include"`
	Input *Input `commented:"true"`
}

// OutputInclude is a reusable Output definition
type OutputInclude struct {
	ID     string  `toml:"id" comment:"identifier for the include"`
	Output *Output `commented:"true"`
}

// TasksInclude is a reusable Tasks definition
type TasksInclude struct {
	ID    string `toml:"id" comment:"identifier for the include"`
	Tasks Tasks  `toml:"Task" commented:"true"`
}

// ExampleInclude returns an Include struct with exemplary values.
func ExampleInclude() *Include {
	return &Include{
		Inputs: []*InputInclude{
			{
				ID:    "input.go",
				Input: exampleInput(),
			},
		},
		Outputs: []*OutputInclude{
			{
				ID:     "output.go",
				Output: exampleOutput(),
			},
		},
		Tasks: []*TasksInclude{
			{
				ID: "task.cbuild",
				Tasks: Tasks{
					&Task{
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
				return PrependValidationErrorPath(err, fmt.Sprintf("Inputs(ID:%s)", in.ID))
			}

			return PrependValidationErrorPath(err, "Inputs")
		}
	}

	for _, out := range incl.Outputs {
		if err := out.Validate(); err != nil {

			if out.ID != "" {
				return PrependValidationErrorPath(err, fmt.Sprintf("Outputs(ID:%s)", out.ID))
			}

			return PrependValidationErrorPath(err, "Outputs")
		}
	}

	for _, tasks := range incl.Tasks {
		if err := tasks.Validate(); err != nil {
			if tasks.ID != "" {
				return PrependValidationErrorPath(err, fmt.Sprintf("Tasks(ID:%s)", tasks.ID))
			}

			return PrependValidationErrorPath(err, "Tasks")
		}

		if len(incl.Inputs) == 0 && len(incl.Outputs) == 0 && len(incl.Tasks) == 0 {
			return &ValidationError{
				Message: "the include does not contain any definition, either an Input, Output or Task must be defined",
			}
		}
	}

	return nil
}

func (in *InputInclude) Validate() error {
	if in.ID == "" {
		return &ValidationError{
			ElementPath: []string{"id"},
			Message:     "can not be empty",
		}
	}

	if in.Input == nil {
		return &ValidationError{
			ElementPath: []string{"Input"},
			Message:     "no input is defined",
		}
	}

	if err := in.Input.Validate(); err != nil {
		return PrependValidationErrorPath(err, "Input")
	}

	return nil
}

func (out *OutputInclude) Validate() error {
	if out.ID == "" {
		return &ValidationError{
			ElementPath: []string{"id"},
			Message:     "can not be empty",
		}
	}

	if out.Output == nil {
		return &ValidationError{
			ElementPath: []string{"Output"},
			Message:     "no output is defined",
		}
	}

	if err := out.Validate(); err != nil {
		return PrependValidationErrorPath(err, "Output")
	}

	return nil
}

func (tasks *TasksInclude) Validate() error {
	if tasks.ID == "" {
		return &ValidationError{
			ElementPath: []string{"id"},
			Message:     "can not be empty",
		}
	}

	if len(tasks.Tasks) == 0 {
		return &ValidationError{
			ElementPath: []string{fmt.Sprintf("Task(id:%s)", tasks.ID)},
			Message:     "no output is defined",
		}
	}

	for _, task := range tasks.Tasks {
		if err := task.Validate(); err != nil {
			if task.Name == "" {
				return PrependValidationErrorPath(err, "Task")
			}

			return PrependValidationErrorPath(err, fmt.Sprintf("Task(name: %s)", task.Name))
		}
	}

	return nil
}
