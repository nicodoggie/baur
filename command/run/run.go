package run

import (
	"fmt"
	"strings"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/command/util"
	"github.com/simplesurance/baur/log"
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

type Options struct {
	skipRecord bool
	skipUpload bool
	force      bool
}

func New() *cobra.Command {
	opts := Options{}

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

func (c *Options) Run(cmd *cobra.Command, args []string) {
	if c.skipRecord && !c.skipUpload {
		log.Fatalln("--skip-upload must be passed when --skip-record is specified")
	}

	if !c.skipUpload {
		log.Fatalln("running tasks without --skip-upload is not implemented")
	}

	repo, err := baur.FindRepositoryCwd()
	util.ExitOnErr(err)

	/*
		cfgPath, err := baur1.FindRepositoryConfig(cwd)
		if err != nil {
			log.Fatalln(err)
		}

		repo, err := baur1.NewRepository(cfgPath)
		if err != nil {
			log.Fatalln(err)
		}

		apps, err := repo.Apps()
		if err != nil {
			log.Fatalln(err)
		}
		for _, app := range apps {
			fmt.Println(app)
			fmt.Println(app.RunTask(args[0]))
		}

		/*
			app := app.App{}
			err := app.RunTask(args[0], runCmdConf.skipRecord, runCmdConf.skipUpload, runCmdConf.force)
			if err != nil {
				log.Fatalln(err)
			}
	*/
}
