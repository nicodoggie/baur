package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/simplesurance/baur/fs"
)

// IncludeDB loads and stores include config files
type IncludeDB struct {
	Inputs  map[string]*InputInclude
	Outputs map[string]*OutputInclude
	Tasks   map[string]*TasksInclude
}

func LoadIncludes(includeDirectory ...string) (*IncludeDB, error) {
	db := IncludeDB{
		Inputs:  map[string]*InputInclude{},
		Outputs: map[string]*OutputInclude{},
		Tasks:   map[string]*TasksInclude{},
	}

	err := db.load(includeDirectory...)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// load reads and validates all *.toml files in the passed includeDirectories
// as include config files and adds them to the database.
// Directories are searched recursively and symlinks are followed.
// Include directices in the loader files are not merged with their includes.
func (db IncludeDB) load(includeDirectory ...string) error {
	walkFunc := func(path string, _ os.FileInfo) error {
		if filepath.Ext(path) != ".toml" {
			return nil
		}

		include, err := IncludeFromFile(path)
		if err != nil {
			return errors.Wrapf(err, "loading include file %q failed", path)
		}

		err = include.Validate()
		if err != nil {
			return errors.Wrapf(err, "validating include config %q failed", path)
		}

		if err := db.add(include); err != nil {
			return err
		}

		return nil
	}

	for _, includeDir := range includeDirectory {
		err := fs.WalkFiles(includeDir, fs.SymlinksAreErrors, walkFunc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db IncludeDB) InputIncludeExist(id string) bool {
	_, exist := db.Inputs[id]
	return exist
}

func (db IncludeDB) OutputIncludeExist(id string) bool {
	_, exist := db.Outputs[id]
	return exist
}

func (db IncludeDB) add(include *Include) error {
	for _, input := range include.Inputs {
		if db.InputIncludeExist(input.ID) || db.OutputIncludeExist(input.ID) {
			return fmt.Errorf("multiple input/output includes with id '%s' are defined, include/output ids must be unique", input.ID)
		}

		db.Inputs[input.ID] = input
	}

	for _, output := range include.Outputs {
		if db.InputIncludeExist(output.ID) || db.OutputIncludeExist(output.ID) {
			return fmt.Errorf("multiple input/output includes with id '%s' are defined, include/output ids must be unique", output.ID)
		}

		db.Outputs[output.ID] = output
	}

	for _, tasks := range include.Tasks {
		if _, exist := db.Tasks[tasks.ID]; exist {
			return fmt.Errorf("multiple tasks includes with id '%s' are defined, include ids must be unique", tasks.ID)
		}

		db.Tasks[tasks.ID] = tasks
	}

	return nil
}
