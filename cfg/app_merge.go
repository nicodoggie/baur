package cfg

import "fmt"

// Merge for each ID in the Includes slice a TasksInclude in the includedb is looked up.
// The tasks of the found TasksInclude are appended to the Apps Tasks slice.
func (a *App) Merge(includedb *IncludeDB) error {
	for _, includeID := range a.Includes {
		include, exist := includedb.Tasks[includeID]
		if !exist {
			return fmt.Errorf("could not find include with id '%s'", includeID)
		}

		a.Tasks = append(a.Tasks, &Task{
			Name:     include.Name,
			Command:  include.Command,
			Includes: include.Includes,
			Input:    include.Input,
			Output:   include.Output,
		})

		// TODO: store the repository relative cfg path in the include somehow
		//task.Input.Files.Paths = append(task.Input.Files.Paths, include.RelCfgPath)
		// TODO: use FieldError for errors
	}

	return a.Tasks.Merge(includedb)
}

func (t Tasks) Merge(includedb *IncludeDB) error {
	for _, task := range t {
		if err := TaskMerge(task, includedb); err != nil {
			return err
		}
	}

	return nil
}

func TaskMerge(t TaskDef, includeDB *IncludeDB) error {
	for _, includeID := range t.GetIncludes() {
		if include, exist := includeDB.Inputs[includeID]; exist {
			if t.GetInput() == nil {
				t.SetInput(&Input{})
			}

			t.GetInput().Files.Merge(&include.Files)
			t.GetInput().GitFiles.Merge(&include.GitFiles)
			t.GetInput().GolangSources.Merge(&include.GolangSources)

			continue
		}

		if include, exist := includeDB.Outputs[includeID]; exist {
			if t.GetOutput() == nil {
				t.SetOutput(&Output{})
			}

			t.GetOutput().Merge(include)

			continue
		}

		return fmt.Errorf("could not find include with id '%s'", includeID)
	}

	return nil

}

// Merge merges the Input with another one.
func (i *Input) Merge(other InputDef) {
	i.Files.Merge(other.FileInputs())
	i.GitFiles.Merge(other.GitFileInputs())
	i.GolangSources.Merge(other.GolangSourcesInputs())
}

// Merge merges the two GolangSources structs
func (g *GolangSources) Merge(other *GolangSources) {
	g.Paths = append(g.Paths, other.Paths...)
	g.Environment = append(g.Environment, other.Environment...)
}

// Merge merges the two Output structs
func (o *Output) Merge(other OutputDef) {
	o.DockerImage = append(o.DockerImage, other.DockerImageOutputs()...)
	o.File = append(o.File, other.FileOutputs()...)
}

// Merge merges 2 FileInputs structs
func (f *FileInputs) Merge(other *FileInputs) {
	f.Paths = append(f.Paths, other.Paths...)
}

// Merge merges two GitFileInputs structs
func (g *GitFileInputs) Merge(other *GitFileInputs) {
	g.Paths = append(g.Paths, other.Paths...)
}
