package baur1

import (
	"fmt"
	"os"

	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/fs"
)

// AppCfgFile contains the name of application configuration files
const AppCfgFile = ".app.toml"

// RepositoryCfgFile contains the name of the repository configuration file.
const RepositoryCfgFile = ".baur.toml"

// FindRepositoryConfig searches for the RepositoryCfgFile. The search starts in
// the current working directory and traverses the parent directories down to '/'. The
// absolute path to the first found RepositoryCfgFile is returned.
// If the config file is not found os.ErrNotExist is returned.
func FindRepositoryConfigCwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	cfgPath, err := fs.FindFileInParentDirs(cwd, RepositoryCfgFile)
	if err != nil {
		return "", fmt.Errorf("could not find repository config file %q: %w", RepositoryCfgFile, err)
	}

	return cfgPath, nil
}

// FindAndLoadRepositoryConfigCwd locates the repository cfg file via
// FindRepositoryConfigCwd(), and then parses and validates it.
func FindAndLoadRepositoryConfigCwd() (*cfg.Repository, error) {
	path, err := FindRepositoryConfigCwd()
	if err != nil {
		return nil, err
	}

	repoCfg, err := cfg.RepositoryFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading repository config file %q failed: %w", path, err)
	}

	if err := repoCfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating repository config file %q failed: %w", path, err)
	}

	return repoCfg, nil
}
