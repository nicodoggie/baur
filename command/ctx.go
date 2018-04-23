package command

import (
	"os"
	"path"

	"github.com/simplesurance/baur"
	"github.com/simplesurance/baur/cfg"
	"github.com/simplesurance/baur/discover"
	"github.com/simplesurance/baur/sblog"
)

// Ctx stores supporting informations that are required by commands
type ctx struct {
	RepositoryRoot    string
	RepositoryCfg     *cfg.Repository
	RepositoryCfgPath string
}

func mustFindRepositoryRoot() string {
	root, err := discover.RepositoryRoot(baur.RepositoryCfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			sblog.Fatalf("could not find repository root config file "+
				"ensure the file '%s' exist in the root",
				baur.RepositoryCfgFile)
		}
		sblog.Fatal("finding repository root config file failed: ", err)
	}

	sblog.Debugf("repository root found: %v", root)

	return root
}

// InitCtx returns an initialized Ctx. It terminates the application on errors.
func mustInitCtx() *ctx {
	var err error
	var ctx ctx

	ctx.RepositoryRoot = mustFindRepositoryRoot()
	ctx.RepositoryCfgPath = path.Join(ctx.RepositoryRoot, baur.RepositoryCfgFile)

	ctx.RepositoryCfg, err = cfg.RepositoryFromFile(ctx.RepositoryCfgPath)
	if err != nil {
		sblog.Fatal("reading repository config failed: ", err)
	}

	if err = ctx.RepositoryCfg.Validate(); err != nil {
		sblog.Fatalf("validating repository config (%s) failed: %s",
			ctx.RepositoryCfgPath, err)
	}

	return &ctx
}
