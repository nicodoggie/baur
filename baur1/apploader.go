package baur1

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
		return "", err
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
		return nil, err
	}

	if err := repoCfg.Validate(); err != nil {
		return nil, err
	}

	return repoCfg, nil
}

// AppLoader discovers, and and loads application configuration files in a repository.
type AppLoader struct {
	appCfgLoader   *cfg.AppLoader
	repositoryRoot string

	configPaths []string
}

func NewAppLoader(repoCfg *cfg.Repository) (*AppLoader, error) {
	repositoryRootDir := filepath.Dir(repoCfg.FilePath())

	includeDb, err := cfg.LoadIncludes(fs.AbsPaths(repositoryRootDir, repoCfg.IncludeDirs)...)
	if err != nil {
		return nil, err
	}

	configPaths, err := findAppConfigs(fs.AbsPaths(repositoryRootDir, repoCfg.Discover.Dirs), repoCfg.Discover.SearchDepth)
	if err != nil {
		return nil, fmt.Errorf("discovering application configs failed: %w", err)
	}

	return &AppLoader{
		appCfgLoader:   &cfg.AppLoader{IncludeDB: includeDb},
		configPaths:    configPaths,
		repositoryRoot: repositoryRootDir,
	}, nil
}

func (db *AppLoader) Path(path string) (*App, error) {
	return db.load(path)
}

func (db *AppLoader) Name(name string) (*App, error) {
	for _, path := range db.configPaths {
		app, err := db.Path(path)
		if err != nil {
			return nil, err
		}

		if app.cfg.Name == name {
			return app, nil
		}
	}

	return nil, errors.New("application not found")
}

func (db *AppLoader) All() ([]*App, error) {
	result := make([]*App, 0, len(db.configPaths))

	for _, path := range db.configPaths {
		cfg, err := db.Path(path)
		if err != nil {
			return nil, err
		}

		result = append(result, cfg)
	}

	return result, nil
}

func (db *AppLoader) load(path string) (*App, error) {
	cfg, err := db.appCfgLoader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("loading config file %q failed: %w", path, err)
	}

	return &App{
		repositoryRoot: db.repositoryRoot,
		cfg:            cfg,
		directory:      filepath.Dir(path),
	}, nil
}

func findAppConfigs(searchDirs []string, searchDepth int) ([]string, error) {
	var result []string

	for _, searchDir := range searchDirs {
		if err := fs.DirsExist(searchDir); err != nil {
			return nil, fmt.Errorf("application search directory: %w", err)
		}

		cfgPaths, err := fs.FindFilesInSubDir(searchDir, AppCfgFile, searchDepth)
		if err != nil {
			return nil, fmt.Errorf("discovering application configs failed: %w", err)
		}

		result = append(result, cfgPaths...)
	}

	return result, nil
}
