package filecopy

import (
	"context"
	"io"
	"net/url"
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/simplesurance/baur/fs"
)

var defLogFn = func(string, ...interface{}) {}

// Client copies files from one path to another
type Client struct {
	debugLogFn func(string, ...interface{})
}

// New returns a client
func New(debugLogFn func(string, ...interface{})) *Client {
	logFn := defLogFn
	if debugLogFn != nil {
		logFn = debugLogFn
	}

	return &Client{debugLogFn: logFn}
}

func copyFile(src, dst string) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "opening %s failed", src)
	}

	// nolint: errcheck
	defer srcFd.Close()

	srcFi, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "stat %s failed", src)
	}

	srcFileMode := srcFi.Mode().Perm()

	dstFd, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcFileMode)
	if err != nil {
		return errors.Wrapf(err, "opening %s failed", dst)
	}

	_, err = io.Copy(dstFd, srcFd)
	if err != nil {
		_ = dstFd.Close()

		return err
	}

	return dstFd.Close()
}

// Upload copies the file with src path to the dst path.
// If the destination directory does not exist, it is created.
// If the destination path exist and is not a regular file an error is returned.
// If it exist and is a file, the file is overwritten if it's not the same.
func (c *Client) Upload(_ context.Context, src string, dest *url.URL) (string, error) {
	fpath := dest.Path
	destDir := path.Dir(fpath)

	isDir, err := fs.IsDir(destDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		err = fs.Mkdir(destDir)
		if err != nil {
			return "", errors.Wrapf(err, "creating directory '%s' failed", destDir)
		}

		c.debugLogFn("filecopy: created directory '%s'", destDir)
	} else {
		if !isDir {
			return "", errors.Wrapf(err, "%s is not a directory", destDir)
		}
	}

	regFile, err := fs.IsRegularFile(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		return fpath, copyFile(src, fpath)
	}

	if !regFile {
		return "", errors.Wrapf(err, "'%s' exist but is not a regular file", fpath)
	}

	sameFile, err := fs.SameFile(src, fpath)
	if err != nil {
		return "", err
	}

	if sameFile {
		c.debugLogFn("filecopy: '%s' already exist and is the same then '%s'", fpath, src)
		return fpath, nil
	}

	c.debugLogFn("filecopy: '%s' already exist, overwriting file", fpath)

	return fpath, copyFile(src, fpath)
}

func (c *Client) URIScheme() string {
	return "file"
}
