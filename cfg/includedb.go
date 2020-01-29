package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/simplesurance/baur/fs"
)

type Logger interface {
	Debugf(format string, v ...interface{})
}

// IncludeDB loads and stores include config files
type IncludeDB struct {
	// TODO: make fields private? They seem not to be accessed
	Inputs  map[string]*InputInclude
	Outputs map[string]*OutputInclude
	Tasks   map[string]*TasksInclude
	logger  Logger
}

func LoadIncludes(logger Logger, includeDirectory ...string) (*IncludeDB, error) {
	db := IncludeDB{
		Inputs:  map[string]*InputInclude{},
		Outputs: map[string]*OutputInclude{},
		Tasks:   map[string]*TasksInclude{},
		logger:  logger,
	}

	err := db.load(includeDirectory)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// load reads and validates all *.toml files in the passed includeDirectories
// as include config files and adds them to the database.
// Directories are searched recursively and symlinks are followed.
// Include directices in the loader files are not merged with their includes.
func (db *IncludeDB) load(includeDirectories []string) error {
	if err := db.loadIncludeFiles(includeDirectories); err != nil {
		return err
	}

	// TODO: use FieldError for validation errors
	for id, incl := range db.Inputs {
		if err := incl.Validate(); err != nil {
			return fmt.Errorf("validating input include %q failed: %w", id, err)
		}
	}

	for id, incl := range db.Outputs {
		if err := incl.Validate(); err != nil {
			return fmt.Errorf("validating output include %q failed: %w", id, err)
		}
	}

	for id, taskIncl := range db.Tasks {
		if err := taskIncl.Tasks.Merge(db); err != nil {
			return fmt.Errorf("merging task include %q with it's includes failed: %w", id, err)
		}

		if err := taskIncl.Validate(); err != nil {
			return fmt.Errorf("validating task include %q failed: %w", id, err)
		}
	}

	return nil
}

func (db *IncludeDB) loadIncludeFiles(includeDirectory []string) error {
	walkFunc := func(path string, _ os.FileInfo) error {
		if filepath.Ext(path) != ".toml" {
			return nil
		}

		db.logger.Debugf("includedb: loading includes from file %q", path)

		include, err := IncludeFromFile(path)
		if err != nil {
			return fmt.Errorf("loading include file %q failed: %w", path, err)
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

func (db *IncludeDB) InputIncludeExist(id string) bool {
	_, exist := db.Inputs[id]
	return exist
}

func (db *IncludeDB) OutputIncludeExist(id string) bool {
	_, exist := db.Outputs[id]
	return exist
}

func (db *IncludeDB) add(include *Include) error {
	for _, input := range include.Inputs {
		if db.InputIncludeExist(input.ID) || db.OutputIncludeExist(input.ID) {
			return fmt.Errorf("multiple input/output includes with id '%s' are defined, include/output ids must be unique", input.ID)
		}

		db.logger.Debugf("includedb: loaded include %q", input.ID)

		db.Inputs[input.ID] = input
	}

	for _, output := range include.Outputs {
		if db.InputIncludeExist(output.ID) || db.OutputIncludeExist(output.ID) {
			return fmt.Errorf("multiple input/output includes with id '%s' are defined, include/output ids must be unique", output.ID)
		}

		db.logger.Debugf("includedb: loaded include %q", output.ID)

		db.Outputs[output.ID] = output
	}

	for _, tasks := range include.Tasks {
		if _, exist := db.Tasks[tasks.ID]; exist {
			return fmt.Errorf("multiple tasks includes with id '%s' are defined, include ids must be unique", tasks.ID)
		}

		db.logger.Debugf("includedb: loaded include %q", tasks.ID)

		db.Tasks[tasks.ID] = tasks
	}

	return nil
}
