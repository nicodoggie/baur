package baur1

import (
	"fmt"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/digest/sha384"
)

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

// Path returns the absolute path
func (f *InputFile) Path() string {
	return f.absPath
}

// LocalPath returns the path relative to application directory
func (f *InputFile) LocalPath() string {
	return f.localPath
}

// String returns LocalPath()
func (f *InputFile) String() string {
	return f.localPath
}

// CalcDigest calculates the digest of the file, saves it and returns it.
func (f *InputFile) CalcDigest() (*digest.Digest, error) {
	sha := sha384.New()

	err := sha.AddBytes([]byte(f.absPath))
	if err != nil {
		return nil, err
	}

	err = sha.AddFile(f.absPath)
	if err != nil {
		return nil, err
	}

	f.digest = sha.Digest()

	return f.digest, nil
}

// Digest returns the previous calculated digest.
// If the digest wasn't calculated yet, CalcDigest() is called and it's return
// values are returned.
func (f *InputFile) Digest() (*digest.Digest, error) {
	if f.digest != nil {
		return f.digest, nil
	}

	return f.CalcDigest()
}

func TotalInputDigest(files []*InputFile) (*digest.Digest, error) {
	digests := make([]*digest.Digest, 0, len(files))

	for _, file := range files {
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
