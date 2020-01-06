package resolve

import (
	"github.com/simplesurance/baur/resolve/gitpath"
	"github.com/simplesurance/baur/resolve/glob"
)

type GitGlobPathResolver interface {
	Resolve(workingDir string, globs ...string) ([]string, error)
}

type GlobPathResolver interface {
	Resolve(globPath ...string) ([]string, error)
}

type Resolver struct {
	GitGlobPathResolver GitGlobPathResolver
	GlobPathResolver    GlobPathResolver
}

func NewResolver() *Resolver {
	return &Resolver{
		GlobPathResolver:    &glob.Resolver{},
		GitGlobPathResolver: &gitpath.Resolver{},
	}
}

func (r *Resolver) GitGlobPath(workingDir string, globs ...string) ([]string, error) {
	return r.GitGlobPathResolver.Resolve(workingDir, globs...)
}

func (r *Resolver) GlobPath(globPath ...string) ([]string, error) {
	return r.GlobPathResolver.Resolve(globPath...)
}
