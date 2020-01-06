package baur

import (
	"github.com/simplesurance/baur/digest"
)

// BuildOutput is an interface for build artifacts
type BuildOutput interface {
	Exists() bool
	Name() string
	String() string
	LocalPath() string
	UploadDestination() string
	Digest() (*digest.Digest, error)
	Size(*BuildOutputBackends) (int64, error)
	Type() string
}
