// Package seq implements a simple Sequential Uploader. Upload jobs are
// processed sequentially in-order.
package seq

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/simplesurance/baur/upload/scheduler"
)

// Logger defines the logger interface
type Logger interface {
	Debugf(format string, v ...interface{})
}

// DockerUploader is a client for uploading docker images to a registry
type DockerUploader interface {
	Upload(image, registryAddr, repository, tag string) (string, error)
}

//FileUploader is an interface for storing files in another place
type FileUploader interface {
	Upload(from, to string) (string, error)
}

// Uploader is a sequential uploader
type Uploader struct {
	filecopy       FileUploader
	s3             FileUploader
	docker         DockerUploader
	lock           sync.Mutex
	queue          []scheduler.Job
	stopProcessing bool
	statusChan     chan<- *scheduler.Result
	logger         Logger
}

// New initializes a sequential uploader
// Status chan must have a buffer count > 1 otherwise a deadlock occurs
func New(logger Logger, filecopyUploader, s3Uploader FileUploader, dockerUploader DockerUploader, status chan<- *scheduler.Result) *Uploader {
	return &Uploader{
		logger:     logger,
		s3:         s3Uploader,
		statusChan: status,
		lock:       sync.Mutex{},
		queue:      []scheduler.Job{},
		docker:     dockerUploader,
		filecopy:   filecopyUploader,
	}
}

// Add adds a new upload job, can be called after Start()
func (u *Uploader) Add(job scheduler.Job) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.queue = append(u.queue, job)
}

// Start starts uploading jobs in the queue.
// If the statusChan buffer is full, uploading will be blocked.
func (u *Uploader) Start() {
	for {
		var job scheduler.Job

		u.lock.Lock()
		if len(u.queue) > 0 {
			job = u.queue[0]
			u.queue = u.queue[1:]
		}
		u.lock.Unlock()

		if job != nil {
			var err error
			var url string
			startTs := time.Now()

			u.logger.Debugf("uploading %s", job)
			switch j := job.(type) {
			case *scheduler.FileCopyJob:
				url, err = u.filecopy.Upload(j.Src, j.Dst)
				if err != nil {
					err = errors.Wrap(err, "file copy failed")
				}

			case *scheduler.S3Job:
				url, err = u.s3.Upload(j.FilePath, j.DestURL)
				if err != nil {
					err = errors.Wrap(err, "S3 upload failed")
				}

			case *scheduler.DockerJob:
				url, err = u.docker.Upload(j.ImageID, j.Registry, j.Repository, j.Tag)
				if err != nil {
					err = errors.Wrap(err, "Docker upload failed")
				}
			default:
				panic(fmt.Sprintf("job has unsupported type %+v", job))
			}

			u.statusChan <- &scheduler.Result{
				Err:      err,
				URL:      url,
				Duration: time.Since(startTs),
				Job:      job,
			}
		}

		u.lock.Lock()
		if len(u.queue) == 0 {
			time.Sleep(time.Second)
		}

		if u.stopProcessing {
			close(u.statusChan)
			u.lock.Unlock()
			return
		}
		u.lock.Unlock()
	}
}

// Stop stops the uploader
func (u *Uploader) Stop() {
	u.lock.Lock()
	u.stopProcessing = true
	u.lock.Unlock()
}
