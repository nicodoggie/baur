package baur1

import "net/url"

type OutputType int

const (
	DockerOutput OutputType = iota
	FileOutput
)

type Output interface {
	UploadDestination() *url.URL

	// LocalPath is the path for accessing the output locally. This can be
	// the path to a file, an URL or e.g. an ID of a docker image that is
	// in the local registry.
	LocalPath() (string, error)

	Name() string
	Type() OutputType
}
