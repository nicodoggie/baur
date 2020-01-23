package cfg

// DefaultIncludePathResolver returns the default set of resolvers for
// ResolveIncludes()
func DefaultIncludePathResolvers(rootPath string) Resolver {
	return NewRootVarResolver(rootPath)
}

// DefaultResolvers returns the default set of resolvers.
func DefaultResolvers(rootPath, appName string) Resolver {
	return Resolvers{
		NewRootVarResolver(rootPath),
		NewAppVarResolver(appName),
		&UUIDVarResolver{},
	}
}

// ResolveIncludes resolves the include paths in the apps
func (a *App) ResolveIncludes(resolvers Resolver) error {
	for i, includePath := range a.Includes {
		var err error

		if a.Includes[i], err = resolvers.Resolve(includePath); err != nil {
			return NewFieldError(err, "Includes", includePath)
		}
	}

	if err := a.Tasks.ResolveIncludes(resolvers); err != nil {
		return NewFieldError(err, "tasks")
	}

	return nil
}

func (tasks Tasks) ResolveIncludes(resolvers Resolver) error {
	for _, t := range tasks {
		if err := t.ResolveIncludes(resolvers); err != nil {
			return NewFieldError(err, "Task", t.Name)
		}
	}

	return nil
}

func (t *Task) ResolveIncludes(resolvers Resolver) error {
	for i, includePath := range t.Includes {
		var err error

		if t.Includes[i], err = resolvers.Resolve(includePath); err != nil {
			return NewFieldError(err, "Includes", includePath)
		}
	}

	return nil
}

func (a *App) Resolve(resolvers Resolver) error {
	if err := a.Tasks.Resolve(resolvers); err != nil {
		return NewFieldError(err, "Tasks")
	}

	return nil
}

func (tasks Tasks) Resolve(resolvers Resolver) error {
	for _, t := range tasks {
		if err := t.Resolve(resolvers); err != nil {
			return NewFieldError(err, "task", t.Name)
		}
	}

	return nil
}

func (t *Task) Resolve(resolvers Resolver) error {
	var err error

	if t.Command, err = resolvers.Resolve(t.Command); err != nil {
		return NewFieldError(err, "Command")
	}

	if err := t.Input.Resolve(resolvers); err != nil {
		return NewFieldError(err, "Input")
	}

	if err := t.Output.Resolve(resolvers); err != nil {
		return NewFieldError(err, "Output")
	}

	return nil
}

func (i *Input) Resolve(resolvers Resolver) error {
	if err := i.Files.Resolve(resolvers); err != nil {
		return NewFieldError(err, "Files")
	}

	if err := i.GitFiles.Resolve(resolvers); err != nil {
		return NewFieldError(err, "Gitfiles")
	}

	if err := i.GolangSources.Resolve(resolvers); err != nil {
		return NewFieldError(err, "GoLangSources")
	}

	return nil
}

func (f *FileInputs) Resolve(resolvers Resolver) error {
	for i, p := range f.Paths {
		var err error

		if f.Paths[i], err = resolvers.Resolve(p); err != nil {
			return NewFieldError(err, "Paths", p)
		}
	}

	return nil
}

func (g *GitFileInputs) Resolve(resolvers Resolver) error {
	for i, p := range g.Paths {
		var err error

		if g.Paths[i], err = resolvers.Resolve(p); err != nil {
			return NewFieldError(err, "Paths", p)
		}
	}

	return nil
}

func (g *GolangSources) Resolve(resolvers Resolver) error {
	for i, env := range g.Environment {
		var err error

		if g.Environment[i], err = resolvers.Resolve(env); err != nil {
			return NewFieldError(err, "Environment", env)
		}
	}

	for i, p := range g.Paths {
		var err error

		if g.Paths[i], err = resolvers.Resolve(p); err != nil {
			return NewFieldError(err, "Paths", p)
		}
	}

	return nil
}

func (o *Output) Resolve(resolvers Resolver) error {
	for _, dockerImage := range o.DockerImage {
		if err := dockerImage.Resolve(resolvers); err != nil {
			return NewFieldError(err, "DockerImage")
		}
	}

	for _, file := range o.File {
		if err := file.Resolve(resolvers); err != nil {
			return NewFieldError(err, "FileOutput")
		}
	}

	return nil
}

func (f *FileOutput) Resolve(resolvers Resolver) error {
	var err error

	if f.Path, err = resolvers.Resolve(f.Path); err != nil {
		return NewFieldError(err, "path")
	}

	if err = f.FileCopy.Resolve(resolvers); err != nil {
		return NewFieldError(err, "FileCopy")
	}

	if err = f.S3Upload.Resolve(resolvers); err != nil {
		return NewFieldError(err, "S3Upload")
	}

	return nil
}

func (f *FileCopy) Resolve(resolvers Resolver) error {
	var err error

	if f.Path, err = resolvers.Resolve(f.Path); err != nil {
		return NewFieldError(err, "path")
	}

	return nil
}

func (s *S3Upload) Resolve(resolvers Resolver) error {
	var err error

	if s.Bucket, err = resolvers.Resolve(s.Bucket); err != nil {
		return NewFieldError(err, "bucket")
	}

	if s.DestFile, err = resolvers.Resolve(s.DestFile); err != nil {
		return NewFieldError(err, "dest_file")
	}

	return nil
}

func (d *DockerImageOutput) Resolve(resolvers Resolver) error {
	var err error

	if d.IDFile, err = resolvers.Resolve(d.IDFile); err != nil {
		return NewFieldError(err, "idfile")
	}

	if err = d.RegistryUpload.Resolve(resolvers); err != nil {
		return NewFieldError(err, "RegistryUpload")
	}

	return nil
}

func (d *DockerImageRegistryUpload) Resolve(resolvers Resolver) error {
	var err error

	if d.Repository, err = resolvers.Resolve(d.Repository); err != nil {
		return NewFieldError(err, "repository")
	}

	if d.Tag, err = resolvers.Resolve(d.Tag); err != nil {
		return NewFieldError(err, "tag")
	}

	return nil
}
