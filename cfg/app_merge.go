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

		a.Tasks = append(a.Tasks, include.Tasks...)

		// TODO: store the repository relative cfg path in the include somehow
		//task.Input.Files.Paths = append(task.Input.Files.Paths, include.RelCfgPath)
		// TODO: use FieldError for errors
	}

	return a.Tasks.Merge(includedb)
}

func (t Tasks) Merge(includedb *IncludeDB) error {
	for _, task := range t {
		if err := task.Merge(includedb); err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) Merge(includeDB *IncludeDB) error {
	for _, includeID := range t.Includes {
		if include, exist := includeDB.Inputs[includeID]; exist {
			if t.Input == nil {
				t.Input = &Input{}
			}

			t.Input.Merge(include.Input)
			continue
		}

		if include, exist := includeDB.Outputs[includeID]; exist {
			if t.Output == nil {
				t.Output = &Output{}
			}

			t.Output.Merge(include.Output)

			continue
		}

		return fmt.Errorf("could not find include with id '%s'", includeID)
	}

	return nil

}

// Merge merges the Input with another one.
func (i *Input) Merge(other *Input) {
	i.Files.Merge(&other.Files)
	i.GitFiles.Merge(&other.GitFiles)
	i.GolangSources.Merge(&other.GolangSources)
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

// Merge merges 2 FileInputs structs
func (f *FileInputs) Merge(other *FileInputs) {
	f.Paths = append(f.Paths, other.Paths...)
}

// Merge merges two GitFileInputs structs
func (g *GitFileInputs) Merge(other *GitFileInputs) {
	g.Paths = append(g.Paths, other.Paths...)
}
