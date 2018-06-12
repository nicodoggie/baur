package baur

import (
	"github.com/simplesurance/baur/upload"
)

// Artifact is an interface for build artifacts
type Artifact interface {
	Exists() bool
	UploadJob() (upload.Job, error)
	String() string
}