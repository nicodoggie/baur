package baur1

import (
	"net/url"
)

// OutputFile is a file created by a task run.
type OutputFile struct {
	name              string
	absPath           string
	uploadDestination *url.URL
}

func NewOutputFile(name, absPath string, uploadDestination *url.URL) *OutputFile {
	return &OutputFile{
		name:              name,
		absPath:           absPath,
		uploadDestination: uploadDestination,
	}
}

// UploadDestination returns the upload destination
func (f *OutputFile) UploadDestination() *url.URL {
	return f.uploadDestination
}

// Name returns the name of the file.
// TODO: can we use a better term then name?
// We use the value to when recording the artifact in the db
func (f *OutputFile) Name() string {
	return f.name
}

// String returns Name()
func (f *OutputFile) String() string {
	return "file: " + f.Name()
}

// LocalPath returns the absolute path of the file.
// It always returns a nil error.
func (f *OutputFile) LocalPath() (string, error) {
	return f.absPath, nil
}

func (f *OutputFile) Type() OutputType {
	return FileOutput
}
