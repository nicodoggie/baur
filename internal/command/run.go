package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/baur1"

	// TODO: not not import the util package, rename it
	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/internal/command/util"
	"github.com/simplesurance/baur/log"
	"github.com/simplesurance/baur/resolve/gitpath"
	"github.com/simplesurance/baur/resolve/glob"
	"github.com/simplesurance/baur/resolve/gosource"
	"github.com/simplesurance/baur/upload/docker"
	"github.com/simplesurance/baur/upload/filecopy"
	"github.com/simplesurance/baur/upload/s3"
	"github.com/spf13/cobra"
)

// TODO:
// - support specifying only app name, to run all tasks of the app
// - support specifying only task name, to run tasks for all apps with the same name
// TODO: Passing "*" as argument is not nice to use in a shell, without quoting it will expand

var runLongHelp = fmt.Sprintf(`
Run tasks.
If no argument is passed, all tasks in the repository are run.
By default only tasks with status %s are run.

Tasks-Specifier is in the format:
    <APPLICATION>.<TASK>
    <APPLICATION> or <TASK> can be '*' to match all applications or tasks.

The following Environment Variables are supported:
    %s

  S3 Upload:
    %s
    %s
    %s

  Docker Registry Upload:
    %s
    %s
    %s
    %s
    %s
    %s
`,
	util.ColoredBuildStatus(baur.BuildStatusPending),

	util.Highlight(util.EnvVarPSQLURL),

	util.Highlight("AWS_REGION"),
	util.Highlight("AWS_ACCESS_KEY_ID"),
	util.Highlight("AWS_SECRET_ACCESS_KEY"),

	util.Highlight(util.EnvVarDockerUsername),
	util.Highlight(util.EnvVarDockerPassword),
	util.Highlight("DOCKER_HOST"),
	util.Highlight("DOCKER_API_VERSION"),
	util.Highlight("DOCKER_CERT_PATH"),
	util.Highlight("DOCKER_TLS_VERIFY"))

const runExample = `
baur run payment-service.build	Run the build task of the payment-service application if it's status is pending.
baur run *.check		Run all check tasks in status pending of all applications
baur run --force --skip-upload	Run all tasks of all application, rerun them if their status is not pending, skip uploading outputs
`

type RunOptions struct {
	skipUpload bool
	force      bool
}

func NewRunCommand() *cobra.Command {
	opts := RunOptions{}

	cmd := cobra.Command{
		Use:     "run [<TASK-SPECIFIER>]",
		Short:   "run tasks",
		Long:    strings.TrimSpace(runLongHelp),
		Run:     opts.Run,
		Example: strings.TrimSpace(runExample),
		Args:    cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVarP(&opts.skipUpload, "skip-upload", "s", false,
		"skip uploading task outputs")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false,
		"force rebuilding of tasks with status "+baur.BuildStatusExist.String())

	return &cmd

}

type execUserData struct {
	task             *baur1.Task
	inputs           []*baur1.InputFile
	totalInputDigest *digest.Digest
}

const (
	dockerEnvUsernameVar = "BAUR_DOCKER_USERNAME"
	dockerEnvPasswordVar = "BAUR_DOCKER_PASSWORD"
)

func dockerAuthFromEnv() (string, string) {
	return os.Getenv(dockerEnvUsernameVar), os.Getenv(dockerEnvPasswordVar)
}

func (c *RunOptions) Run(cmd *cobra.Command, args []string) {
	log.StdLogger.EnableDebug(verboseFlag)

	// TODO: validate syntax of args[0] or the check in taskLoader.Load sufficient?
	taskSpec := args[0]

	if taskSpec == "" {
		taskSpec = "*.*"
	}

	repoCfg, err := baur1.FindAndLoadRepositoryConfigCwd()
	ExitOnErr(err)

	repositoryDir := filepath.Dir(repoCfg.FilePath())
	log.Debugf("found repository root: %q", repositoryDir)

	taskLoader, err := baur1.NewTaskLoader(repoCfg)
	ExitOnErr(err)

	tasks, err := taskLoader.Load(taskSpec)
	ExitOnErr(err)

	// TODO: use MustGetPostgresClt() instead
	clt, err := getPostgresCltWithEnv(repoCfg.Database.PGSQLURL)
	if err != nil {
		log.Fatalf("could not establish connection to postgreSQL db: %s", err)
	}

	inputResolver := baur1.NewInputResolver(
		log.StdLogger,
		&glob.Resolver{},
		&gitpath.Resolver{},
		gosource.NewResolver(log.StdLogger.Debugf),
	)

	taskStatusMgr := baur1.NewTaskStatusManager(
		repositoryDir,
		log.StdLogger,
		clt,
		inputResolver,
	)

	s3Uploader, err := s3.NewClient(log.StdLogger)
	ExitOnErr(err)

	var dockerUploader *docker.Client
	dockerUser, dockerPass := dockerAuthFromEnv()
	if len(dockerUser) != 0 {
		log.Debugf("using docker authentication data from %s, %s Environment variables, authenticating as '%s'",
			dockerEnvUsernameVar, dockerEnvPasswordVar, dockerUser)
		dockerUploader, err = docker.NewClientwAuth(log.StdLogger.Debugf, dockerUser, dockerPass)
	} else {
		log.Debugf("environment variable %s not set", dockerEnvUsernameVar)
		dockerUploader, err = docker.NewClient(log.StdLogger.Debugf)
	}
	if err != nil {
		log.Fatalln(err)
	}

	filecopyUploader := filecopy.New(log.Debugf)

	taskRunner := baur1.NewTaskRunner(
		log.StdLogger,
		&baur1.OutputStreams{
			Stdout: baur1.NewStream(os.Stdout),
			Stderr: baur1.NewStream(os.Stderr),
		},
		taskStatusMgr,
		&baur1.Uploaders{
			Filecopy: filecopyUploader,
			Docker:   dockerUploader,
			S3:       s3Uploader,
		},
	)

	// TODO: refactor how we retrieve tasks, remove the App struct completly?
	var filter baur1.RunFilter
	if c.force {
		filter = baur1.RunFilterAlways
	} else {
		filter = baur1.RunFilterOnlyPendingTasks
	}

	err = taskRunner.Run(tasks, filter, c.skipUpload)
	if err != nil {
		log.Fatalf("running tasks failed: %s\n", err)
	}
}
