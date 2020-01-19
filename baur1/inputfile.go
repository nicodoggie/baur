package baur1

import (
	"fmt"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/digest/sha384"
)

// InputFile represent a file
type InputFile struct {
	*File
	localPath string
}

// NewInputFile returns a new file
func NewInputFile(absPath, localPath string) *InputFile {
	return &InputFile{
		File:      &File{AbsPath: absPath},
		localPath: localPath,
	}
}

// LocalPath returns the path relative to application directory
func (f *InputFile) LocalPath() string {
	return f.localPath
}

// String returns LocalPath()
func (f *InputFile) String() string {
	return f.localPath
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
