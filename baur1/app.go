package baur1

import (
	"github.com/simplesurance/baur/cfg"
)

type App struct {
	cfg       *cfg.App
	directory string

	repositoryRoot string
}

func (a *App) String() string {
	return a.cfg.Name
}

func (a *App) Tasks() ([]*Task, error) {
	result := make([]*Task, 0, len(a.cfg.Tasks))

	for _, task := range a.cfg.Tasks {
		task, err := NewTask(a.Name(), task, a.directory)
		if err != nil {
			return nil, err
		}

		result = append(result, task)
	}

	return result, nil
}

func (a *App) Directory() string {
	return a.directory
}

func (a *App) Name() string {
	return a.cfg.Name
}

func (a *App) Config() *cfg.App {
	return a.cfg
}
