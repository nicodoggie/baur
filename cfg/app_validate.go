package cfg

import (
	"errors"
	"fmt"
	"strings"
)

// Validate validates a App configuration
func (a *App) Validate() error {
	if len(a.Name) == 0 {
		return NewFieldError(errors.New("can not be empty"), "name")
	}

	if err := a.Tasks.Validate(); err != nil {
		return NewFieldError(err, "Tasks")
	}

	return nil
}

func (tasks Tasks) Validate() error {
	if len(tasks) != 1 {
		// Is wrapped into PathError with path "Tasks" by the App.Validate() caller
		return fmt.Errorf("must contain exactly 1 Task definition, has %d", len(tasks))
	}

	duplMap := make(map[string]struct{}, len(tasks))

	for _, task := range tasks {
		_, exist := duplMap[task.Name]
		if exist {
			return NewFieldError(
				fmt.Errorf("multiple tasks with name '%s' exist, task names must be unique", task.Name),
				"Task", task.Name,
			)
		}
		duplMap[task.Name] = struct{}{}

		err := task.Validate()
		if err != nil {
			return NewFieldError(
				fmt.Errorf("multiple tasks with name '%s' exist, task names must be unique", task.Name),
				"Task", task.Name,
			)
		}
	}

	return nil
}

// Validate validates the task section
func (t *Task) Validate() error {
	if len(t.Command) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"command",
		)
	}

	// TODO: change it to check for an invalid name when we support multiple tasks
	if t.Name != "build" {
		return NewFieldError(
			errors.New("name must be 'build'"),
			"name",
		)
	}

	if t.Input == nil {
		return NewFieldError(
			errors.New("section is empty"),
			"Input",
		)
	}

	if err := t.Input.Validate(); err != nil {
		return NewFieldError(err, "Input")
	}

	if t.Output == nil {
		return NewFieldError(
			errors.New("section is empty"),
			"Output",
		)
	}

	if err := t.Output.Validate(); err != nil {
		return NewFieldError(err, "Output")
	}

	return nil
}

// Validate validates the Input section
func (i *Input) Validate() error {
	if err := i.Files.Validate(); err != nil {
		return NewFieldError(err, "Files")
	}

	if err := i.GolangSources.Validate(); err != nil {
		return NewFieldError(err, "GolangSources")
	}

	// TODO: add validation for gitfiles section

	return nil
}

// Validate validates the GolangSources section
func (g *GolangSources) Validate() error {
	if len(g.Environment) != 0 && len(g.Paths) == 0 {
		return NewFieldError(
			errors.New("must be set if environment is set"),
			"paths",
		)
	}

	for _, p := range g.Paths {
		if len(p) == 0 {
			return NewFieldError(
				errors.New("empty string is an invalid path"),
				"paths",
			)
		}
	}

	return nil
}

// Validate validates the Output section
func (o *Output) Validate() error {
	for _, f := range o.File {
		if err := f.Validate(); err != nil {
			return NewFieldError(err, "File")
		}
	}

	for _, d := range o.DockerImage {
		if err := d.Validate(); err != nil {
			return NewFieldError(err, "DockerImage")
		}
	}

	return nil
}

// Validate validates a [[Task.Output.File]] section
func (f *FileOutput) Validate() error {
	if len(f.Path) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"path",
		)
	}

	return f.S3Upload.Validate()
}

// Validate validates a [[Task.Output.File]] section
func (s *S3Upload) Validate() error {
	if s.IsEmpty() {
		return nil
	}

	if len(s.DestFile) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"destfile",
		)
	}

	if len(s.Bucket) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"bucket",
		)
	}

	return nil
}

// Validate validates its content
func (d *DockerImageOutput) Validate() error {
	if len(d.IDFile) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"idfile",
		)
	}

	if err := d.RegistryUpload.Validate(); err != nil {
		return NewFieldError(err, "RegistryUpload")
	}

	return nil
}

// Validate validates its content
func (d *DockerImageRegistryUpload) Validate() error {
	if len(d.Repository) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"repository",
		)
	}

	if len(d.Tag) == 0 {
		return NewFieldError(
			errors.New("can not be empty"),
			"tag",
		)
	}

	return nil
}

// Validate validates a [[Sources.Files]] section
func (f *FileInputs) Validate() error {
	for _, path := range f.Paths {
		if len(path) == 0 {
			return NewFieldError(
				errors.New("can not be empty"),
				"path",
			)

		}
		if strings.Count(path, "**") > 1 {
			return NewFieldError(
				errors.New("'**' can only appear one time in a path"),
				"path",
			)
		}
	}

	return nil
}
