package command

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/command/util"
	"github.com/simplesurance/baur/format"
	"github.com/simplesurance/baur/fs"
	"github.com/simplesurance/baur/log"
	"github.com/simplesurance/baur/storage"
	"github.com/simplesurance/baur/storage/postgres"
)

func findRepository() (*baur.Repository, error) {
	log.Debugln("searching for repository root...")

	repo, err := baur.FindRepositoryCwd()
	if err != nil {
		return nil, err
	}

	log.Debugf("repository root found: %s", repo.Path)

	return repo, nil
}

// MustFindRepository must find repo
func MustFindRepository() *baur.Repository {
	repo, err := findRepository()
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("could not find repository root config file "+
				"ensure the file '%s' exist in the root",
				baur.RepositoryCfgFile)
		}

		log.Fatalln(err)
	}

	return repo
}

func isAppDir(arg string) bool {
	cfgPath := path.Join(arg, baur.AppCfgFile)
	isFile, _ := fs.IsFile(cfgPath)
	return isFile
}

func mustArgToApp(repo *baur.Repository, arg string) *baur.App {
	if isAppDir(arg) {
		app, err := repo.AppByDir(arg)
		if err != nil {
			log.Fatalln(err)
		}

		return app
	}

	app, err := repo.AppByName(arg)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("could not find application with name '%s'", arg)
		}

		log.Fatalln(err)
	}

	return app
}

// getPostgresCltWithEnv returns a new postresql storage client,
// if the environment variable BAUR_PSQL_URI is set, this uri is used instead of
// the configuration specified in the baur.Repository object
func getPostgresCltWithEnv(psqlURI string) (*postgres.Client, error) {
	uri := psqlURI

	if envURI := os.Getenv(util.EnvVarPSQLURL); len(envURI) != 0 {
		log.Debugf("using postgresql connection URL from $%s environment variable",
			util.EnvVarPSQLURL)

		uri = envURI
	} else {
		log.Debugf("environment variable $%s not set", util.EnvVarPSQLURL)
	}

	return postgres.New(uri)
}

//mustHavePSQLURI calls log.Fatalf if neither util.EnvVarPSQLURL nor the postgres_url
//in the repository config is set
func mustHavePSQLURI(r *baur.Repository) {
	if len(r.Config().Database.PGSQLURL) != 0 {
		return
	}

	if len(os.Getenv(util.EnvVarPSQLURL)) == 0 {
		log.Fatalf("PostgreSQL connection information is missing.\n"+
			"- set postgres_url in your repository config or\n"+
			"- set the $%s environment variable", util.EnvVarPSQLURL)
	}
}

// MustGetPostgresClt must return the PG client
func MustGetPostgresClt(r *baur.Repository) *postgres.Client {
	mustHavePSQLURI(r)

	clt, err := getPostgresCltWithEnv(r.Config().Database.PGSQLURL)
	if err != nil {
		log.Fatalf("could not establish connection to postgreSQL db: %s", err)
	}

	return clt
}

func mustGetCommitID(r *baur.Repository) string {
	commitID, err := r.GitCommitID()
	if err != nil {
		log.Fatalln(err)
	}

	return commitID
}

func mustGetGitWorktreeIsDirty(r *baur.Repository) bool {
	isDirty, err := r.GitWorkTreeIsDirty()
	if err != nil {
		log.Fatalln(err)
	}

	return isDirty
}

func vcsStr(v *storage.VCSState) string {
	if len(v.CommitID) == 0 {
		return ""
	}

	if v.IsDirty {
		return fmt.Sprintf("%s-dirty", v.CommitID)
	}

	return v.CommitID
}

func mustArgToApps(repo *baur.Repository, args []string) []*baur.App {
	if len(args) == 0 {
		apps, err := repo.FindApps()
		if err != nil {
			log.Fatalln(err)
		}

		if len(apps) == 0 {
			log.Fatalf("could not find any applications\n"+
				"- ensure the [Discover] section is correct in %s\n"+
				"- ensure that you have >1 application dirs "+
				"containing a %s file",
				repo.Config().FilePath(), baur.AppCfgFile)
		}

		return apps
	}

	dedupMap := make(map[string]struct{}, len(args))
	apps := make([]*baur.App, 0, len(args))
	for _, arg := range args {
		app := mustArgToApp(repo, arg)
		if _, exist := dedupMap[app.Path]; exist {
			continue
		}

		dedupMap[app.Path] = struct{}{}
		apps = append(apps, mustArgToApp(repo, arg))
	}

	return apps
}

func mustWriteRow(fmt format.Formatter, row []interface{}) {
	err := fmt.WriteRow(row)
	if err != nil {
		log.Fatalln(err)
	}
}

func bytesToMib(bytes int) string {
	return fmt.Sprintf("%.3f", float64(bytes)/1024/1024)
}

func durationToStrSeconds(duration time.Duration) string {
	return fmt.Sprintf("%.3f", duration.Seconds())
}

func ExitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
