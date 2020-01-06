package baur1

import (
	"net/url"

	"github.com/simplesurance/baur/fs"
)

// OutputFile is a file build artifact
type OutputFile struct {
	localPath         string
	absPath           string
	uploadDestination *url.URL
}

// Exists returns true if the artifact exist
func (f *OutputFile) Exists() bool {
	return fs.FileExists(f.absPath)
}

// String returns LocalPath()
func (f *OutputFile) String() string {
	return f.localPath
}

// LocalPath returns the local path of the file
func (f *OutputFile) LocalPath() string {
	return f.localPath
}

// UploadDestination returns the upload destination
func (f *OutputFile) UploadDestination() *url.URL {
	return f.uploadDestination
}

// Size returns the size of the file in bytes
func (f *OutputFile) Size() (int64, error) {
	return fs.FileSize(f.LocalPath())
}

func (f *OutputFile) Path() string {
	return f.absPath
}

// Type returns "File"
func (f *OutputFile) Type() string {
	return "File"
}
