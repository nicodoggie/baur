package baur1

import (
	"fmt"
	"sort"

	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/resolve"
)

type Task struct {
	Directory string
	resolver  *resolve.Resolver

	AppName          string
	Command          string
	Name             string
	Outputs          *cfg.Output
	UnresolvedInputs *cfg.Input
}

func NewTask(appName string, cfg *cfg.Task, workingDir string) (*Task, error) {
	return &Task{
		Directory:        workingDir,
		Outputs:          cfg.Output,
		Command:          cfg.Command,
		Name:             cfg.Name,
		AppName:          appName,
		UnresolvedInputs: cfg.Input,
	}, nil
}

// ID returns <APP-NAME>.<TASK-NAME>
func (t *Task) ID() string {
	return fmt.Sprintf("%s.%s", t.AppName, t.Name)
}

func (t *Task) String() string {
	return t.ID()
}

func SortTasksByID(tasks []*Task) {
	sort.Slice(tasks, func(i int, j int) bool {
		return tasks[i].ID() < tasks[i].ID()
	})
}
