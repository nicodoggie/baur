package baur1

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/fs"
	"github.com/simplesurance/baur/log"
)

type TaskLoader struct {
	logger         *log.Logger
	includeDB      *cfg.IncludeDB
	repositoryRoot string
	appConfigPaths []string
}

func NewTaskLoader(repoCfg *cfg.Repository) (*TaskLoader, error) {
	repositoryRootDir := filepath.Dir(repoCfg.FilePath())

	includeDB, err := cfg.LoadIncludes(fs.AbsPaths(repositoryRootDir, repoCfg.IncludeDirs)...)
	if err != nil {
		return nil, err
	}

	// TODO: log which appconfigpaths were found as debug msg
	appConfigPaths, err := findAppConfigs(fs.AbsPaths(repositoryRootDir, repoCfg.Discover.Dirs), repoCfg.Discover.SearchDepth)
	if err != nil {
		return nil, fmt.Errorf("discovering application config files failed: %w", err)
	}

	return &TaskLoader{
		logger:         log.StdLogger,
		repositoryRoot: repositoryRootDir,
		includeDB:      includeDB,
		appConfigPaths: appConfigPaths,
	}, nil
}

// Load loads the task that match the passed specifier.
// Valid specifiers are:
// - application-directory path
// - <APP-NAME>
// - <APP-NAME>.<TASK-NAME>
// - The '*' wildcard is supported for <APP-NAME> and <TASK-NAME>. '*' will
//   match all apps or tasks.
func (t *TaskLoader) Load(specifier string) ([]*Task, error) {
	if cfgPath, isAppDirectory := IsAppDirectory(specifier); isAppDirectory {
		t.logger.Debugf("taskloader: loading app from path %q", cfgPath)

		return t.Path(path.Join(specifier, cfgPath))
	}

	if !strings.Contains(specifier, ".") {
		t.logger.Debugf("taskloader: loading tasks by app name %q", specifier)

		return t.AppName(specifier)
	}

	spl := strings.Split(specifier, ".")
	if len(spl) != 2 {
		return nil, errors.New("invalid specifier")
	}

	appSpec := spl[0]
	taskSpec := spl[1]

	if appSpec == "" {
		return nil, errors.New("invalid specifier, app part is empty")
	}

	if taskSpec == "" {
		return nil, errors.New("invalid specifier, task part is empty")
	}

	if appSpec == "*" && taskSpec == "*" {
		return t.All()
	}

	if appSpec != "*" {
		if taskSpec == "*" {
			t.logger.Debugf("taskloader: loading tasks by app name %q", appSpec)

			return t.AppName(appSpec)
		}

		if taskSpec != "*" {
			t.logger.Debugf("taskloader: loading task %q of app %q", appSpec, taskSpec)
			task, err := t.AppTask(appSpec, taskSpec)
			if err != nil {
				return nil, err
			}

			return []*Task{task}, nil
		}
	}

	t.logger.Debugf("taskloader: loading %q tasks of all apps", taskSpec)
	return t.Task(taskSpec)
}

func (t *TaskLoader) All() ([]*Task, error) {
	result := make([]*Task, 0, len(t.appConfigPaths))

	for _, path := range t.appConfigPaths {
		tasks, err := t.Path(path)
		if err != nil {
			return nil, err
		}

		result = append(result, tasks...)
	}

	return result, nil
}

func (t *TaskLoader) Path(appConfigPath string) ([]*Task, error) {
	appCfg, err := cfg.AppFromFile(appConfigPath)
	if err != nil {
		return nil, err
	}

	return t.fromAppCfg(appCfg)
}

func (t *TaskLoader) fromAppCfg(appCfg *cfg.App) ([]*Task, error) {
	err := appCfg.ResolveIncludes(cfg.DefaultIncludePathResolvers(t.repositoryRoot))
	if err != nil {
		return nil, fmt.Errorf("resolving variables in include paths failed: %w", err)
	}

	err = appCfg.Merge(t.includeDB)
	if err != nil {
		return nil, fmt.Errorf("merging with includes failed: %w", err)
	}

	err = appCfg.Resolve(cfg.DefaultResolvers(t.repositoryRoot, appCfg.Name))
	if err != nil {
		return nil, fmt.Errorf("resolving variables in config failed: %w", err)
	}

	result := make([]*Task, 0, len(appCfg.Tasks))

	appDirectory := filepath.Dir(appCfg.FilePath())

	for _, task := range appCfg.Tasks {
		task, err := NewTask(appCfg.Name, task, appDirectory)
		if err != nil {
			return nil, err
		}

		result = append(result, task)
	}

	return result, nil
}

func (t *TaskLoader) Task(name string) ([]*Task, error) {
	tasks, err := t.All()
	if err != nil {
		return nil, err
	}

	var result []*Task

	for _, task := range tasks {
		if task.Name == "name" {
			result = append(result, task)
		}
	}

	if len(result) == 0 {
		return nil, errors.New("no task with the name exist")
	}

	return result, nil
}

func (t *TaskLoader) AppName(name string) ([]*Task, error) {
	for _, path := range t.appConfigPaths {
		appCfg, err := cfg.AppFromFile(path)
		if err != nil {
			return nil, err
		}

		if appCfg.Name != name {
			continue
		}

		return t.fromAppCfg(appCfg)
	}

	return nil, errors.New("app not found")
}

func (t *TaskLoader) AppTask(appName, taskName string) (*Task, error) {
	tasks, err := t.AppName(appName)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.Name == taskName {
			return task, nil
		}
	}

	return nil, fmt.Errorf("app %q has no task named %q", appName, taskName)
}

func IsAppDirectory(arg string) (string, bool) {
	cfgPath := path.Join(arg, baur.AppCfgFile)
	isFile, _ := fs.IsFile(cfgPath)

	return cfgPath, isFile
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
