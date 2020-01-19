package baur1

import (
	"net/url"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/fs"
)

// OutputFile is a file created by a task run.
type OutputFile struct {
	*File
	name              string
	uploadDestination *url.URL
}

// TODO: make the hasher implementation exchangeable

func NewOutputFile(absPath string, uploadDestination *url.URL) *OutputFile {
	return &OutputFile{
		File:              &File{AbsPath: absPath},
		uploadDestination: uploadDestination,
	}
}

// UploadDestination returns the upload destination
func (f *OutputFile) UploadDestination() *url.URL {
	return f.uploadDestination
}

// String returns Name()
func (f *OutputFile) String() string {
	return "file: " + f.Path()
}

func (f *OutputFile) Type() OutputType {
	return FileOutput
}

func (f *OutputFile) Digest() (*digest.Digest, error) {
	return f.File.Digest()
}

func (f *OutputFile) Exists() (bool, error) {
	return fs.FileExists(f.AbsPath), nil
}

func (f *OutputFile) Size() (int64, error) {
	return fs.FileSize(f.AbsPath)
}
