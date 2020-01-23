package cfg

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

type FieldError struct {
	ElementPath []string
	err         error
}

func (f *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", strings.Join(f.ElementPath, "."), f.err)
}

// NewFieldError creates a new FieldError that wraps the passed error if the
// passed error is not of type FieldError.
// If it is of type FieldError, the passed paths are prepended to the ElementPath
// of it.
func NewFieldError(err error, path ...string) error {
	valError, ok := err.(*FieldError)
	if ok {
		valError.ElementPath = append(path, valError.ElementPath...)
		return err
	}

	return &FieldError{
		ElementPath: path,
		err:         err,
	}
}

// toFile serializes a struct to TOML format and writes it to a file.
func toFile(data interface{}, filepath string, overwrite bool) error {
	var openFlags int

	if overwrite {
		openFlags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	} else {
		openFlags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}

	f, err := os.OpenFile(filepath, openFlags, 0640)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(f)
	encoder.Order(toml.OrderPreserve)
	err = encoder.Encode(data)
	if err != nil {
		f.Close()
		return err
	}

	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "closing file failed")
	}

	return err
}
