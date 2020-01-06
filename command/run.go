package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/baur1"
	"github.com/simplesurance/baur/command/util"
	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/fs"
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
Run Tasks.
If no argument is passed, all tasks in the repository are run,.
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
	skipRecord bool
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
	cmd.Flags().BoolVarP(&opts.skipRecord, "skip-record", "r", false,
		"skip recording the results to the database, --skip-upload must also be passed")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false,
		"force rebuilding of tasks with status "+baur.BuildStatusExist.String())

	return &cmd

}

type execUserData struct {
	app              *baur1.App
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
	if c.skipRecord && !c.skipUpload {
		log.Fatalln("--skip-upload must be passed when --skip-record is specified")
	}

	log.StdLogger.EnableDebug(verboseFlag)

	repoCfg, err := baur1.FindAndLoadRepositoryConfigCwd()
	appLoader, err := baur1.NewAppLoader(repoCfg)
	util.ExitOnErr(err)

	var apps []*baur1.App

	// TODO: improve this if-condition, make it beautiful
	if len(args) == 0 {
		var err error
		apps, err = appLoader.All()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// TODO handle err and Isfile correctly,
		if isFile, _ := fs.IsRegularFile(args[0]); isFile {
			app, err := appLoader.Path(args[0])
			util.ExitOnErr(err)
			apps = append(apps, app)
		} else {
			app, err := appLoader.Name(args[0])
			util.ExitOnErr(err)
			apps = append(apps, app)
		}

	}

	// TODO: use MustGetPostgresClt() instead
	clt, err := getPostgresCltWithEnv(repoCfg.Database.PGSQLURL)
	if err != nil {
		log.Fatalf("could not establish connection to postgreSQL db: %s", err)
	}

	repositoryDir := filepath.Dir(repoCfg.FilePath())

	inputResolver := baur1.NewInputResolver(
		log.StdLogger,
		&glob.Resolver{},
		&gitpath.Resolver{},
		gosource.NewResolver(log.StdLogger.Debugf),
	)

	digestCalc := baur1.DigestCalc{}

	taskStatusMgr := baur1.NewTaskStatusManager(
		repositoryDir,
		log.StdLogger,
		clt,
		inputResolver,
		&digestCalc,
	)

	s3Uploader, err := s3.NewClient(log.StdLogger)
	if err != nil {
		log.Fatalln(err.Error())
	}

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
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
		taskStatusMgr,
		&digestCalc,
		&baur1.Uploaders{
			Filecopy: filecopyUploader,
			Docker:   dockerUploader,
			S3:       s3Uploader,
		},
	)

	// TODO: refactor how we retrieve tasks, remove the App struct completly?
	var tasks []*baur1.Task
	for _, app := range apps {
		appTasks, err := app.Tasks()

		if err != nil {
			log.Fatalf("initializing tasks failed: %s", err)
		}

		tasks = append(tasks, appTasks...)
	}

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
