package baur1

import (
	"fmt"

	"github.com/simplesurance/baur/digest"
	"github.com/simplesurance/baur/digest/sha384"
)

type DigestCalc struct{}

// TODO: add a cache

/// TODO: storing the digest here in the file via SetDigest() seems ugly. What is a better to retrieve it for a file later?
// Recalculating it is not an option to slow and theoretically the content might have been changed by the task run.
// Retrieving it from a cache via it's filepath is also not ideal, the cache and the real implementation would not be exchangeable anymore, with the real implenetation the previous described issue exist.
//also it's faster to store + access it in the file

func (d *DigestCalc) TotalInputDigest(files []*InputFile) (*digest.Digest, error) {
	digests := make([]*digest.Digest, 0, len(files))

	for _, file := range files {
		digest, err := d.InputFileDigest(file)
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
