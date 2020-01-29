package command

import (
	"os"

	"github.com/simplesurance/baur/baur1"
	"github.com/simplesurance/baur/format"
	"github.com/simplesurance/baur/format/csv"
	"github.com/simplesurance/baur/format/table"
	"github.com/spf13/cobra"
)

// TODO: add support for choosing fields that are shown in output

type lsTasksOptions struct {
	csv      bool
	absPaths bool
}

const taskSpecDoc = `  TASK-SPEC
	The argument specifies the apps and tasks to match.
	The format is: <PATH>|<APP-NAME>[.<TASK-NAME>]

	The wildcard '*' is supported for <APP-NAME> and <TASK-NAME>.
	'*' will match all apps or tasks.`

const lsTasksLongHelp = `List tasks and their status.

Arguments:
` + taskSpecDoc

func NewLsTasksCommand() *cobra.Command {
	var opts lsTasksOptions

	cmd := cobra.Command{
		Use:   "tasks [<TASK-SPEC>]",
		Short: "list tasks and their status",
		Long:  lsTasksLongHelp,
		Run:   opts.Run,
		Args:  cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVar(&lsAppsConfig.csv, "csv", false,
		"List applications in RFC4180 CSV format")

	return &cmd
}

func (l *lsTasksOptions) Run(cmd *cobra.Command, args []string) {
	_, tasks := MustArgToTasks(args)

	baur1.SortTasksByID(tasks)

	headers := []string{
		"Task",
		"Command",
		"Path",
	}

	var formatter format.Formatter
	if l.csv {
		formatter = csv.New(headers, os.Stdout)
	} else {
		formatter = table.New(headers, os.Stdout)
	}

	// TODO: printed task is the same then the one that is passed as cmdline parameter instead of the abs or repository rel path,
	// e.g. passing "../currency/ as arg, will also show the same path
	for _, task := range tasks {
		err := formatter.WriteRow([]interface{}{task.ID(), task.Command, task.Directory})
		ExitOnErr(err)
	}

	err := formatter.Flush()
	ExitOnErr(err)
}
