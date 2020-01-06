// package scheduler provides a simple sequential background upload scheduler.
package scheduler

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"
)

type Uploader interface {
	Upload(ctx context.Context, source string, destinationURI *url.URL) (string, error)
	URIScheme() string
}

type UploadResult struct {
	Job UploadJob

	Error error

	URL       string
	StartTime time.Time
	EndTime   time.Time
}

type UploadJob interface {
	fmt.Stringer

	Source() string
	Destination() *url.URL
}

type Logger interface {
	Debugf(format string, v ...interface{})
}

type Sequential struct {
	logger    Logger
	uploaders map[string]Uploader

	queue      chan UploadJob
	resultChan chan<- *UploadResult

	running sync.WaitGroup
}

func NewSequential(logger Logger, queue chan UploadJob, resultChan chan<- *UploadResult, uploaders ...Uploader) (*Sequential, error) {
	uploadM := make(map[string]Uploader, len(uploaders))

	for _, uploader := range uploaders {
		if _, exist := uploadM[uploader.URIScheme()]; exist {
			return nil, fmt.Errorf("multiple uploaders for same scheme %q passed", uploader.URIScheme())
		}

		uploadM[uploader.URIScheme()] = uploader
	}

	return &Sequential{
		logger:     logger,
		uploaders:  uploadM,
		queue:      queue,
		resultChan: resultChan,
	}, nil
}

// Start starts the Uploader in a Go-Routine and returns.
// The Uploader will handle jobs that are queued.
// The current upload and the running Go-Routines can be stopped via the following ways:
// - closing the queue channel,
// - cancelling the context,
func (s *Sequential) Start(ctx context.Context) {
	s.running.Add(1)

	go s.start(ctx)

	s.logger.Debugf("uploader go-routine started\n")
}

func (s *Sequential) start(ctx context.Context) {
	defer close(s.resultChan)
	defer s.running.Done()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debugf("uploader: context cancelled, terminating go-routine")
			return

		case job, chanIsOpen := <-s.queue:
			if !chanIsOpen {
				s.logger.Debugf("uploader: queue channel closed terminating go-routine")
				return
			}

			uploader, err := s.uploader(job)
			if err != nil {
				result := &UploadResult{
					Job:   job,
					Error: fmt.Errorf("uploading %s failed: %w", job, err),
				}

				select {
				case s.resultChan <- result:

				case <-ctx.Done():
					return
				}

				continue
			}

			result := s.upload(ctx, uploader, job)
			select {
			case s.resultChan <- result:

			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *Sequential) upload(ctx context.Context, uploader Uploader, job UploadJob) *UploadResult {
	result := UploadResult{
		Job: job,
	}

	s.logger.Debugf("starting upload of %s\n", job)

	result.StartTime = time.Now()
	result.URL, result.Error = uploader.Upload(ctx, job.Source(), job.Destination())
	result.EndTime = time.Now()

	if result.Error == nil {
		s.logger.Debugf("upload of %s finished successfully\n", job)
	} else {
		s.logger.Debugf("upload of %s failed: %s\n", job, result.Error)
	}

	if result.URL == "" {
		result.Error = fmt.Errorf("%s uploader returned nil error and empty url for job %s", uploader.URIScheme(), job)
	}

	return &result
}

func (s *Sequential) uploader(job UploadJob) (Uploader, error) {
	scheme := job.Destination().Scheme

	uploader, exist := s.uploaders[scheme]
	if !exist {
		return nil, fmt.Errorf("no uploader registered for scheme %s", scheme)
	}

	return uploader, nil
}

// Queue submits the job to the queue.
// If the queue channel is full, the call blocks.
// If the channel is closed, the method will panic.
func (s *Sequential) Queue(job UploadJob) {
	s.queue <- job
	s.logger.Debugf("upload job %s queued\n", job)
}
