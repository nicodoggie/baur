package baur

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/fs"
	"github.com/simplesurance/baur/git"
)

// Repository represents an repository containing applications
type Repository struct {
	Path               string
	CfgPath            string
	AppSearchDirs      []string
	SearchDepth        int
	gitCommitID        string
	gitWorktreeIsDirty *bool
	PSQLURL            string
	AppLoader          cfg.AppLoader
}

// FindRepository searches for a repository config file. The search starts in
// the passed directory and traverses the parent directory down to '/'. The first found repository
// configuration file is returned.
func FindRepository(dir string) (*Repository, error) {
	rootPath, err := fs.FindFileInParentDirs(dir, RepositoryCfgFile)
	if err != nil {
		return nil, err
	}

	return NewRepository(rootPath)
}

// FindRepositoryCwd searches for a repository config file in the current directory
// and all it's parents. If a repository config file is found it returns a
// Repository
func FindRepositoryCwd() (*Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return FindRepository(cwd)
}

// NewRepository reads the configuration file and returns a Repository
func NewRepository(cfgPath string) (*Repository, error) {
	repoCfg, err := cfg.RepositoryFromFile(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err,
			"reading repository config %s failed", cfgPath)
	}

	err = repoCfg.Validate()
	if err != nil {
		return nil, errors.Wrapf(err,
			"validating repository config %q failed", cfgPath)
	}

	r := Repository{
		CfgPath:       cfgPath,
		Path:          path.Dir(cfgPath),
		AppSearchDirs: fs.PathsJoin(path.Dir(cfgPath), repoCfg.Discover.Dirs),
		SearchDepth:   repoCfg.Discover.SearchDepth,
		PSQLURL:       repoCfg.Database.PGSQLURL,
	}

	err = fs.DirsExist(r.AppSearchDirs...)
	if err != nil {
		return nil, errors.Wrapf(err, "validating repository config %q failed, "+
			"application_dirs parameter is invalid", cfgPath)
	}

	includeDB := cfg.NewIncludeDB()
	err = includeDB.Load(fs.AbsPaths(r.Path, repoCfg.IncludeDirs)...)
	if err != nil {
		return nil, errors.Wrap(err, "loading includes failed")
	}

	// TODO: we should release the memory when it's not used anymore,
	// either empty the db when it's not used anymore or store it somewhere
	// else then in the Repository struct
	r.AppLoader = cfg.AppLoader{IncludeDB: includeDB}

	return &r, nil
}

// FindApps searches for application config files in the AppSearchDirs of the
// repository and returns all found apps
func (r *Repository) FindApps() ([]*App, error) {
	var result []*App

	for _, searchDir := range r.AppSearchDirs {
		appsCfgPaths, err := fs.FindFilesInSubDir(searchDir, AppCfgFile, r.SearchDepth)
		if err != nil {
			return nil, errors.Wrap(err, "finding application configs failed")
		}

		for _, appCfgPath := range appsCfgPaths {
			a, err := NewApp(r, appCfgPath)
			if err != nil {
				return nil, err
			}

			result = append(result, a)
		}
	}

	return result, nil
}

// AppByDir reads an application config file from the direcory and returns an
// App
func (r *Repository) AppByDir(appDir string) (*App, error) {
	cfgPath := path.Join(appDir, AppCfgFile)

	cfgPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return nil, err
	}

	return NewApp(r, cfgPath)
}

// AppByName searches for an App with the given name in the repository and
// returns it. If none is found os.ErrNotExist is returned.
func (r *Repository) AppByName(name string) (*App, error) {
	for _, searchDir := range r.AppSearchDirs {
		appsCfgPaths, err := fs.FindFilesInSubDir(searchDir, AppCfgFile, r.SearchDepth)
		if err != nil {
			return nil, errors.Wrap(err, "finding application failed")
		}

		for _, appCfgPath := range appsCfgPaths {
			a, err := NewApp(r, appCfgPath)
			if err != nil {
				return nil, err
			}
			if a.Name == name {
				return a, nil
			}
		}
	}

	return nil, os.ErrNotExist
}

// GitCommitID returns the Git commit ID in the baur repository root
func (r *Repository) GitCommitID() (string, error) {
	if len(r.gitCommitID) != 0 {
		return r.gitCommitID, nil
	}

	commit, err := git.CommitID(r.Path)
	if err != nil {
		return "", errors.Wrap(err, "determining Git commit ID failed, "+
			"ensure that the git command is in a directory in $PATH and "+
			"that the .baur.toml file is part of a git repository")
	}

	r.gitCommitID = commit

	return commit, nil
}

// GitWorkTreeIsDirty returns true if the git repository contains untracked
// changes
func (r *Repository) GitWorkTreeIsDirty() (bool, error) {
	if r.gitWorktreeIsDirty != nil {
		return *r.gitWorktreeIsDirty, nil
	}

	isDirty, err := git.WorkTreeIsDirty(r.Path)
	if err != nil {
		return false, errors.Wrap(err, "determining Git worktree state failed, "+
			"ensure that the git command is in a directory in $PATH and "+
			"that the .baur.toml file is part of a git repository")
	}

	r.gitWorktreeIsDirty = &isDirty

	return isDirty, nil
}
