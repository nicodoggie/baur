package baur1

import (
	"errors"
	"fmt"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/log"
	"github.com/simplesurance/baur/storage"
	"github.com/simplesurance/baur/term"
)

// TaskStatus describes the status of the task
type TaskStatus int

const (
	_ TaskStatus = iota
	TaskStatusUndefined
	TaskStatusExecutionExist
	TaskStatusExecutionPending
)

func (b TaskStatus) String() string {
	switch b {
	case TaskStatusUndefined:
		return "Undefined"
	case TaskStatusExecutionExist:
		return "Exist"
	case TaskStatusExecutionPending:
		return "Pending"
	default:
		panic(fmt.Sprintf("undefined TaskStatus value: %d", b))
	}
}

func (b TaskStatus) ColoredString() string {
	switch b {
	case TaskStatusUndefined:
		return term.YellowHighlight(b.String())
	case TaskStatusExecutionExist:
		return term.GreenHighlight(b.String())
	case TaskStatusExecutionPending:
		return term.RedHighlight(b.String())
	default:
		panic(fmt.Sprintf("undefined TaskStatus value: %d", b))
	}
}

// TODO: is there a better fitting name then TaskStatusManager?
// TODO: make logger and inputresolver interfaces?
type TaskStatusManager struct {
	repositoryDir string

	logger        *log.Logger
	store         storage.Storer
	inputResolver *InputResolver
	digestCalc    *DigestCalc
}

func NewTaskStatusManager(
	repositoryDir string,
	logger *log.Logger,
	store storage.Storer,
	inputResolver *InputResolver,
	digestCalc *DigestCalc,
) *TaskStatusManager {
	return &TaskStatusManager{
		repositoryDir: repositoryDir,
		logger:        logger,
		store:         store,
		inputResolver: inputResolver,
		digestCalc:    digestCalc,
	}
}

// GetTaskStatus calculates the total input digest of the app and checks in the
// storage if a build for this input digest already exist.
// If the function returns TaskStatusExist the returned build pointer is valid
// otherwise it is nil.
func (t *TaskStatusManager) TaskStatus(appName, taskName string, totalInputDigest *digest.Digest) (TaskStatus, *storage.BuildWithDuration, error) {
	build, err := t.store.GetLatestBuildByDigest(appName, totalInputDigest.String())
	if err != nil {
		if err == storage.ErrNotExist {
			return TaskStatusExecutionPending, nil, nil
		}

		return -1, nil, fmt.Errorf("fetching latest build failed: %w", err)
	}

	return TaskStatusExecutionExist, build, nil
}

func (t *TaskStatusManager) FilterTasks(tasks []*Task, status TaskStatus) ([]*Task, error) {
	// TODO: remove FilterTasks if it's not used
	switch status {
	case TaskStatusExecutionPending:
		return t.pendingTasks(tasks)
	case TaskStatusExecutionExist:
		return nil, errors.New("not implemented")

	default:
		return nil, fmt.Errorf("invalid taskstatus '%d'", status)
	}
}

// TODO: make []*Inputs a custom type and store in it the totalInputDigest? Analog to InputFile?
// TODO: remove this function?
func (t *TaskStatusManager) Status(task *Task) (TaskStatus, []*InputFile, *digest.Digest, int, error) {
	inputs, err := t.inputResolver.Resolve(t.repositoryDir, task)
	if err != nil {
		return TaskStatusUndefined, nil, nil, -1, fmt.Errorf("%s: resolving inputs failed: %w", task.AppName, err)
	}

	totalInputDigest, err := t.digestCalc.TotalInputDigest(inputs)
	if err != nil {
		return TaskStatusUndefined, nil, nil, -1, fmt.Errorf("%s: calculating total input digest failed: %w", task.AppName, err)
	}

	exist, id, err := t.runExists(task, totalInputDigest)
	if err != nil {
		return TaskStatusUndefined, nil, nil, -1, err
	}

	if exist {
		return TaskStatusExecutionExist, inputs, totalInputDigest, id, nil
	}

	return TaskStatusExecutionPending, inputs, totalInputDigest, -1, nil
}

func (t *TaskStatusManager) pendingTasks(tasks []*Task) ([]*Task, error) {
	var result []*Task

	for _, task := range tasks {
		status, _, _, _, err := t.Status(task)
		if err != nil {
			return nil, err
		}

		if status == TaskStatusExecutionPending {
			result = append(result, task)
		}
	}

	return result, nil
}

func (t *TaskStatusManager) runExists(task *Task, totalInputDigest *digest.Digest) (bool, int, error) {
	build, err := t.store.GetLatestBuildByDigest(task.AppName, totalInputDigest.String())
	if err != nil {
		if err == storage.ErrNotExist {
			return false, -1, nil
		}

		return false, -1, fmt.Errorf("%s: querying task run with totalInputDigest %q failed: %w",
			task.AppName, totalInputDigest.String(), err)
	}

	t.logger.Debugf("%s: task %s with totalInputDigest %q run successfully in the past, task ID: %s",
		task.AppName, totalInputDigest, build.ID)

	return true, build.ID, nil
}
