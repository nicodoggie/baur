package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/internal/command/util"
	"github.com/simplesurance/baur/log"
)

func init() {
	initCmd.AddCommand(initIncludeCmd)
}

const defIncludeFilename = "includes.toml"

const initIncludeLongHelp = `
Create an include config file.
If no FILENAME argument is passed, the filename will be '` + defIncludeFilename + `'.`

var initIncludeCmd = &cobra.Command{
	Use:   "include [<FILENAME>]",
	Short: "create an include config file",
	Long:  strings.TrimSpace(initIncludeLongHelp),
	Run:   initInclude,
	Args:  cobra.MaximumNArgs(1),
}

func initInclude(cmd *cobra.Command, args []string) {
	var filename string

	if len(args) == 1 {
		filename = args[0]
	} else {
		filename = defIncludeFilename
	}

	includeID := strings.TrimSuffix(filename, filepath.Ext(filename))

	cfg := cfg.ExampleInclude(includeID)
	err := cfg.IncludeToFile(filename)
	if err != nil {
		if os.IsExist(err) {
			log.Fatalf("%s already exist\n", filename)
		}

		log.Fatalln(err)
	}

	fmt.Printf("Include configuration file was written to %s\n",
		util.Highlight(filename))
}
