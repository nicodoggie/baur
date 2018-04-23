package command

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/discover"
	"github.com/simplesurance/baur/sblog"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all applications in the repository",
	Run:   ls,
}

func ls(cmd *cobra.Command, args []string) {
	ctx := mustInitCtx()

	discover := discover.New(&cfg.AppFileReader{})
	apps, err := discover.Applications(ctx.RepositoryCfg.Discover.Dirs,
		baur.AppCfgFile, ctx.RepositoryCfg.Discover.SearchDepth)
	if err != nil {
		sblog.Fatal("discovering applications failed: ", err)
	}

	if len(apps) == 0 {
		sblog.Fatalf("could not find any applications\n"+
			"- ensure the [Discover] is correct in %s\n"+
			"- ensure that you have >1 application dirs "+
			"containing a %s file",
			ctx.RepositoryCfgPath, baur.AppCfgFile)
	}

	sort.Slice(apps, func(i int, j int) bool {
		return apps[i].Name < apps[j].Name
	})

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintf(tw, "# Name\tDirectory\n")
	for _, a := range apps {
		fmt.Fprintf(tw, "%s\t%s\n", a.Name, a.Dir)
	}
	tw.Flush()
}
