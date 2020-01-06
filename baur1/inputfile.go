package baur1

import (
	"fmt"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/digest/sha384"
)

type Digester interface {
	AddBytes(b []byte) error
	Digest() *digest.Digest
	AddFile(path string) error
}

// InputFile represent a file
type InputFile struct {
	digester  Digester
	localPath string
	absPath   string
	digest    *digest.Digest
}

// NewInputFile returns a new file
func NewInputFile(absPath, localPath string) *InputFile {
	return &InputFile{
		digester:  sha384.New(),
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

// Digest calculates the Digest of the InputFile.
// The input for the digest is the LocalPath of the file and it's content.
// When the digest was successfull calculates, it's stored in the file via file.SetDigest().
func (f *InputFile) Digest() (*digest.Digest, error) {
	if f.digest != nil {
		return f.digest, nil
	}

	// TODO: Should it be localPath or repositoryLocalPath?
	// Ensure also that localPath is always normalized!
	err := f.digester.AddBytes([]byte(f.LocalPath()))
	if err != nil {
		return nil, err
	}

	err = f.digester.AddFile(f.Path())
	if err != nil {
		return nil, err
	}

	f.digest = f.digester.Digest()

	return f.digest, nil
}

type InputFiles []*InputFile

func (inputFiles InputFiles) Digest() (*digest.Digest, error) {
	digests := make([]*digest.Digest, 0, len(inputFiles))

	for _, file := range inputFiles {
		digest, err := file.Digest()
		if err != nil {
			return nil, fmt.Errorf("calculating digest for %q failed: %w", file.Path(), err)
		}

		digests = append(digests, digest)
	}

	totalDigest, err := sha384.Sum(digests)
	if err != nil {
		return nil, err
	}

	return totalDigest, nil
}
