package baur1

import (
	"fmt"
	"path/filepath"

	"github.com/simplesurance/baur/fs"
)

// TODO: remove these interfaces if we don't use them, reference the pointers directly instead

type GitGlobPathResolver interface {
	Resolve(workingDir string, globs ...string) ([]string, error)
}

type GlobPathResolver interface {
	Resolve(globPath ...string) ([]string, error)
}

type GoSourceResolver interface {
	Resolve(env []string, directories ...string) ([]string, error)
}

type InputResolver struct {
	gitGlobPathResolver GitGlobPathResolver
	globPathResolver    GlobPathResolver
	goSourceResolver    GoSourceResolver
	logger              DebugLogger
}

type DebugLogger interface {
	Debugf(format string, v ...interface{})
}

func NewInputResolver(logger DebugLogger, glob GlobPathResolver, gitGlob GitGlobPathResolver, gosource GoSourceResolver) *InputResolver {
	return &InputResolver{
		globPathResolver:    glob,
		gitGlobPathResolver: gitGlob,
		goSourceResolver:    gosource,
	}
}

func (i *InputResolver) Resolve(repositoryRoot string, task *Task) ([]*InputFile, error) {
	// TODO: how to handle symlinks correctly?
	// Create a digest over the link instead over the content?

	goSourcePaths, err := i.goSourceResolver.Resolve(
		task.UnresolvedInputs.GolangSources.Environment,
		fs.AbsPaths(task.Directory, task.UnresolvedInputs.GolangSources.Paths)...,
	)
	if err != nil {
		return nil, fmt.Errorf("resolving golang source inputs failed: %w", err)
	}

	globPaths, err := i.globPathResolver.Resolve(fs.AbsPaths(task.Directory, task.UnresolvedInputs.Files.Paths)...)
	if err != nil {
		return nil, fmt.Errorf("resolving input files failed: %w", err)
	}

	gitPaths, err := i.gitGlobPathResolver.Resolve(task.Directory, task.UnresolvedInputs.GitFiles.Paths...)
	if err != nil {
		return nil, fmt.Errorf("resolving git-file inputs failed: %w", err)
	}

	result := make([]*InputFile, 0, len(goSourcePaths)+len(globPaths)+len(gitPaths))

	err = i.appendInputFile(&result, repositoryRoot, goSourcePaths)
	if err != nil {
		return nil, fmt.Errorf("resolving input files failed: %w", err)
	}

	err = i.appendInputFile(&result, repositoryRoot, globPaths)
	if err != nil {
		return nil, fmt.Errorf("resolving input files failed: %w", err)
	}

	err = i.appendInputFile(&result, repositoryRoot, gitPaths)
	if err != nil {
		return nil, fmt.Errorf("resolving input files failed: %w", err)
	}

	return result, nil
}

func (i *InputResolver) pathsToUniqFiles(repositoryRoot string, pathSlice ...[]string) ([]*InputFile, error) {
	var tlen int
	for _, paths := range pathSlice {
		tlen += len(paths)
	}

	res := make([]*InputFile, 0, tlen)
	dedupMap := make(map[string]struct{}, tlen)

	for _, paths := range pathSlice {
		for _, path := range paths {
			if _, exist := dedupMap[path]; exist {
				i.logger.Debugf("removed duplicate Build Input '%s'", path)
				continue
			}

			relPath, err := filepath.Rel(repositoryRoot, path)
			if err != nil {
				return nil, err
			}

			dedupMap[path] = struct{}{}
			res = append(res, NewInputFile(path, relPath))
		}
	}

	return res, nil
}

func (i *InputResolver) appendInputFile(result *[]*InputFile, repositoryRoot string, filePaths []string) error {
	for _, path := range filePaths {
		relPath, err := filepath.Rel(repositoryRoot, path)
		if err != nil {
			return err
		}

		file := NewInputFile(path, relPath)
		*result = append(*result, file)
	}

	return nil
}
