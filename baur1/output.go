package baur1

import "net/url"

type Output interface {
	UploadDestination() *url.URL
	Path() string
	LocalPath() string
	Exists() bool
	Type() string
}
