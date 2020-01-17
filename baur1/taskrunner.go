package baur1

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/exec"
	"github.com/simplesurance/baur/storage"
	"github.com/simplesurance/baur/term"

	"github.com/simplesurance/baur/log"
	"github.com/simplesurance/baur/upload/docker"
	"github.com/simplesurance/baur/upload/filecopy"
	"github.com/simplesurance/baur/upload/s3"
	"github.com/simplesurance/baur/upload/scheduler"
)

// TODO: use interfaces instead of pointers to the structs?
type TaskRunner struct {
	statusMgr *TaskStatusManager
	logger    *log.Logger
	streams   *OutputStreams
	uploaders *Uploaders
}

type Uploaders struct {
	Filecopy *filecopy.Client
	Docker   *docker.Client
	S3       *s3.Client
}

type OutputStreams struct {
	Stdout io.WriteCloser
	Stderr io.WriteCloser
}

func NewTaskRunner(logger *log.Logger, streams *OutputStreams, statusMgr *TaskStatusManager, uploaders *Uploaders) *TaskRunner {
	return &TaskRunner{
		statusMgr: statusMgr,
		uploaders: uploaders,

		streams: streams,
		logger:  logger,
	}
}

// TODO: implement a print function to that task is passed as first argument prefixes the message with task: in colored yellow
// TODO: we need this type or can we merge it with TaskStatus?
type RunFilter int

const (
	// RunFilterOnlyPendingTasks only run tasks that have a different totalInputDigest then recorded past runs.
	RunFilterOnlyPendingTasks RunFilter = iota
	// RunFilterAlways runs tasks always, also if they run in the same with the same totalInputDigest
	RunFilterAlways
)

func (t *TaskRunner) uploadCount(taskRuns []*taskRun) int {
	var cnt int

	for _, taskRun := range taskRuns {
		for _, file := range taskRun.task.Outputs.File {
			if !file.S3Upload.IsEmpty() {
				cnt++
			}

			if !file.FileCopy.IsEmpty() {
				cnt++
			}
		}

		cnt += len(taskRun.task.Outputs.DockerImage)
	}

	return cnt
}

// TODO: do we need to lock taskRun? Or don't we access it concurrently?
type taskRun struct {
	task            *Task
	runStartTs      time.Time
	runStopTs       time.Time
	finishedUploads []*scheduler.UploadResult

	totalInputDigest *digest.Digest
	inputs           []*InputFile
	outputs          []Output
}

// TODO: streams neeed to be protected with a lock, because we write to it from goroutines, best is to have a constructor for the streams struct that wraps the writer in a write method with a lock

func (t *TaskRunner) filterTasks(tasks []*Task, runFilter RunFilter) ([]*taskRun, error) {
	var result []*taskRun

	// TODO: improve this output message, if we run it with forceFlag it's not clear from the message why the status is still evaluated
	fmt.Fprintf(t.streams.Stdout, "Evaluating status of tasks:\n")

	// TODO: do not query database when --force is passed, not needed
	for _, task := range tasks {
		status, inputs, totalInputDigest, id, err := t.statusMgr.Status(task)
		if err != nil {
			return nil, err
		}

		if status == TaskStatusExecutionExist {
			fmt.Fprintf(t.streams.Stdout, "%s => %s (%s)\n", task, status.ColoredString(), term.GreenHighlight(id))
		} else {
			fmt.Fprintf(t.streams.Stdout, "%s => %s\n", task, status.ColoredString())
		}

		if runFilter == RunFilterAlways || (runFilter == RunFilterOnlyPendingTasks && status == TaskStatusExecutionPending) {
			result = append(result, &taskRun{
				task:             task,
				totalInputDigest: totalInputDigest,
				inputs:           inputs,
			})
		}
	}

	fmt.Fprintln(t.streams.Stdout)

	switch runFilter {
	case RunFilterAlways:
		// TODO: improve this output message somehow
		fmt.Fprintf(t.streams.Stdout, "Running all (%d) tasks independent of their status.\n", len(result))
	case RunFilterOnlyPendingTasks:
		fmt.Fprintf(t.streams.Stdout, "Running %d task(s) with status %s.\n", len(result), TaskStatusExecutionPending.ColoredString())
	default:
		return nil, fmt.Errorf("undefined RunFilter value %d passed", runFilter)
	}

	term.PrintSep(t.streams.Stdout)

	return result, nil
}

func (t *TaskRunner) Run(tasks []*Task, runFilter RunFilter, skipUploading bool) error {
	SortTasksByID(tasks)

	taskRuns, err := t.filterTasks(tasks, runFilter)
	if err != nil {
		return err
	}

	if len(taskRuns) == 0 {
		return nil
	}

	var uploader *scheduler.Sequential
	var processUploadsResultChan chan []error
	var uploadQueue chan scheduler.UploadJob

	uploadCnt := t.uploadCount(taskRuns)

	if skipUploading {
		fmt.Fprintf(t.streams.Stdout, "tasks outputs will not be uploaded\n")
	} else {
		var err error

		// TODO: do not start uploader when skipUploading==true
		uploadResultChan := make(chan *scheduler.UploadResult, uploadCnt)
		uploadQueue = make(chan scheduler.UploadJob, uploadCnt)
		uploader, err = scheduler.NewSequential(
			t.logger,
			uploadQueue,
			uploadResultChan,
			t.uploaders.Filecopy, t.uploaders.S3, t.uploaders.Docker,
		)
		if err != nil {
			return err
		}

		// we use context.Background() because we do not want to cancel
		// in-progress upload + result recording for successful tasks
		// if execution of another tasks fail
		uploader.Start(context.Background())

		processUploadsResultChan = make(chan []error, 1)
		go t.processUploads(context.Background(), uploadResultChan, processUploadsResultChan)
	}

	for _, taskRun := range taskRuns {
		outputs, err := t.run(taskRun)
		if err != nil {
			return err
		}

		taskRun.outputs = outputs

		if !skipUploading {
			t.queueOutputUploads(taskRun, uploader)
		}

		// TODO: record results for tasks that do not have outputs
	}

	if !skipUploading {
		close(uploadQueue) // closing the channel wil terminate the uploader go routine which will also close the uploadResultChan and then the processUploads routine terminates
		errors := <-processUploadsResultChan
		if errors != nil {
			return fmt.Errorf("%+v", errors)
		}
	}

	return nil
}

func (t *TaskRunner) recordTaskRun(taskRun *taskRun) error {
	outputs := make([]*storage.Output, 0, len(taskRun.finishedUploads))

	for _, upload := range taskRun.finishedUploads {
		var uploadMethod storage.UploadMethod

		switch upload.Job.Destination().Scheme {
		case "file":
			uploadMethod = storage.FileCopy
		case "s3":
			uploadMethod = storage.S3
		case "docker":
			uploadMethod = storage.DockerRegistry
		}

		uploadJob, ok := upload.Job.(*UploadJob)
		if !ok {
			panic(fmt.Sprintf("upload job file is not of type *UploadJob"))
		}

		outputDigest, err := uploadJob.Output.Digest()
		if err != nil {
			err := fmt.Errorf("calculating digest for output %q failed: %w", uploadJob.Output, err)
			fmt.Fprintf(t.streams.Stderr, "%s: %s\n", taskRun.task, err)
			return err
		}

		outputSize, err := uploadJob.Output.Size()
		if err != nil {
			err := fmt.Errorf(" digest for output %q failed: %w", uploadJob.Output, err)
			fmt.Fprintf(t.streams.Stderr, "%s: %s\n", taskRun.task, err)
			return err
		}

		var artifactType storage.ArtifactType
		switch uploadJob.Output.Type() {
		case DockerOutput:
			artifactType = storage.DockerArtifact

		case FileOutput:
			artifactType = storage.FileArtifact

		default:
			return fmt.Errorf("unsupported output type: %s", uploadJob.Output.Type())
		}

		//outputDigest, err :=
		outputs = append(outputs, &storage.Output{
			// TODO: do we really need to store Name?? NO!?!?
			// TODO: so far it was the repository relative
			// path for file, now it's the app relative path, do we
			// want to change that?
			Name: uploadJob.Output.String(),
			// TODO: do we need to store Type? Why?
			// We only have files and docker images, files
			// we can not store in a docker registry and
			// vice-versa. So the type can be inferred from the upload url
			// TODO: make this cast safe
			Type: artifactType,

			SizeBytes: outputSize,
			Digest:    outputDigest.String(),

			Upload: storage.Upload{
				UploadDuration: upload.EndTime.Sub(upload.StartTime),
				// TODO: is it URI or URL in uploadResult?
				URI:    upload.URL,
				Method: uploadMethod,
			},
		})
	}

	inputs := make([]*storage.Input, 0, len(taskRun.inputs))
	for _, input := range taskRun.inputs {
		// TODO: also store repositoryDir in taskRunner?
		repositoryRelPath, err := filepath.Rel(t.statusMgr.repositoryDir, input.Path())
		if err != nil {
			return err
		}

		digest, err := input.Digest()
		if err != nil {
			err := fmt.Errorf("retrieving digest for input %q failed: %w", input, err)
			fmt.Fprintf(t.streams.Stderr, "%s: %s\n", taskRun.task, err)
			return err
		}

		inputs = append(inputs, &storage.Input{
			// TODO: we only have files as inputs, so rename it to Path? Or instead really store an URI?
			URI:    repositoryRelPath,
			Digest: digest.String(),
		})
	}

	b := storage.Build{
		Application: storage.Application{
			Name: taskRun.task.AppName,
		},
		// TODO: VCSState
		StartTimeStamp:   taskRun.runStartTs,
		StopTimeStamp:    taskRun.runStopTs,
		Outputs:          outputs,
		TotalInputDigest: taskRun.totalInputDigest.String(),
		Inputs:           inputs,
	}

	t.logger.Debugf("%s: storing task run in database", taskRun.task)

	err := t.statusMgr.store.Save(&b)
	if err != nil {
		fmt.Fprintf(t.streams.Stderr, "%s: recording task run in database failed: %s\n", taskRun.task, err)
		return err
	}

	fmt.Fprintf(t.streams.Stdout, "%s: task run stored in database (id: %d)\n", taskRun.task.ID(), b.ID)

	return nil

}
func (t *TaskRunner) processUploads(ctx context.Context, uploadResultChan <-chan *scheduler.UploadResult, result chan<- []error) {
	var errors []error

	defer func() {
		result <- errors
		close(result)
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case res, chanIsOpen := <-uploadResultChan:
			if !chanIsOpen {
				return
			}

			job, ok := res.Job.(*UploadJob)
			if !ok {
				panic(fmt.Sprintf("taskrunner: Job value in UploadResult %+v is not of type UploadJob", job))
			}

			task := job.TaskRun.task

			// TODO: forward errors to caller Run() routine
			if res.Error != nil {
				err := fmt.Errorf("%s: upload failed: %s", task, res.Error)

				fmt.Fprintf(t.streams.Stderr, err.Error())

				errors = append(errors, err)

				continue
			}

			fmt.Fprintf(t.streams.Stdout, "%s: %s uploaded to %s (%.3f)\n",
				task, res.Job.Source(), res.URL, res.EndTime.Sub(res.StartTime).Seconds())

			job.TaskRun.finishedUploads = append(job.TaskRun.finishedUploads, res)

			expectedOutputCnt := len(job.TaskRun.outputs)
			t.logger.Debugf("%s: %d/%d output uploads finished",
				task, len(job.TaskRun.finishedUploads), expectedOutputCnt)

			if len(job.TaskRun.finishedUploads) != expectedOutputCnt {
				continue
			}

			err := t.recordTaskRun(job.TaskRun)
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", task, err))
			}
		}
	}
}

type UploadJob struct {
	ID      string
	TaskRun *taskRun
	Output  Output

	Src  string
	Dest *url.URL
}

func (u *UploadJob) String() string {
	return u.ID
}

func (u *UploadJob) Source() string {
	return u.Src
}

func (u *UploadJob) Destination() *url.URL {
	return u.Dest
}

func (t *TaskRunner) run(taskRun *taskRun) ([]Output, error) {
	task := taskRun.task

	taskRun.runStartTs = time.Now()

	_, err := exec.ShellCommand(task.Command).
		Directory(task.Directory).
		DebugfPrefix(color.YellowString(fmt.Sprintf("%s: ", task))).
		ExpectSuccess().
		Run()

	taskRun.runStopTs = time.Now()

	if err != nil {
		fmt.Fprintf(t.streams.Stderr, "%s: %s", task, err)
		return nil, err
	}

	fmt.Fprintf(t.streams.Stdout, "%s: task execution successful (%s)\n", task, term.SecondDuration(taskRun.runStopTs.Sub(taskRun.runStartTs)))

	outputs, err := OutputsFromTask(t.uploaders.Docker, task)
	if len(outputs) == 0 {
		t.logger.Debugf("%s: task has no outputs\n", task)
		return nil, nil
	}

	// TODO: move this to somewhere else?
	for _, output := range outputs {
		exist, err := output.Exists()
		if err != nil {
			err := fmt.Errorf("%s: checking if output %q exist failed: %w", task, output, err)
			fmt.Fprintf(t.streams.Stderr, "%s\n", err)

			return nil, err
		}

		if !exist {
			err := fmt.Errorf("output %q was not created by task run", task, output)
			fmt.Fprintf(t.streams.Stderr, "%s: %s\n", task, err)

			return nil, err
		}

		t.logger.Debugf("%s: task run created %s\n", task, output)
	}

	return outputs, nil
}

func (t *TaskRunner) queueOutputUploads(taskRun *taskRun, uploader *scheduler.Sequential) {
	task := taskRun.task

	// TODO: write tests for uploaders to ensure they process the destination url correctly
	for _, output := range taskRun.outputs {
		uploader.Queue(&UploadJob{
			ID:      fmt.Sprintf("%s: %s -> %s", task, output, output.UploadDestination()),
			TaskRun: taskRun,
			Src:     output.Path(),
			Dest:    output.UploadDestination(),
			Output:  output,
		})
	}
}
