package baur1

import "github.com/simplesurance/baur/digest"

// InputFile represent a file
type InputFile struct {
	localPath string
	absPath   string
	digest    *digest.Digest
}

// NewInputFile returns a new file
func NewInputFile(absPath, localPath string) *InputFile {
	return &InputFile{
		localPath: localPath,
		absPath:   absPath,
	}
}

// Path returns it's absolute path
func (f *InputFile) Path() string {
	return f.absPath
}

// LocalPath returns the files local path
func (f *InputFile) LocalPath() string {
	return f.localPath
}

// String returns LocalPath()
func (f *InputFile) String() string {
	return f.localPath
}

// Digest returns the stored digest, must have been set before via SetDigest().
// Otherwise nil is returned.
func (f *InputFile) Digest() *digest.Digest {
	return f.digest
}

func (f *InputFile) SetDigest(d *digest.Digest) {
	f.digest = d
}
