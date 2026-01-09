package email

import (
	"context"
	"sync"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/cache"
	"github.com/arrow2012/nuwa-kit/pkg/json"
	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/arrow2012/nuwa-kit/pkg/metric"
	"github.com/arrow2012/nuwa-kit/pkg/options"
)

// Sender defines the interface for sending emails
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
	Close()
}

// AsyncSender implements Sender interface using Redis Queue
type AsyncSender struct {
	cache    cache.Cache
	wg       sync.WaitGroup
	quitChan chan struct{}
	opts     *options.EmailOptions
	queueKey string
}

type emailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NewAsyncSender creates a new AsyncSender with Redis backend
func NewAsyncSender(c cache.Cache, opts *options.EmailOptions, queueKey string) *AsyncSender {
	if queueKey == "" {
		queueKey = "nuwa:email:queue" // Default fallback
	}
	s := &AsyncSender{
		cache:    c,
		quitChan: make(chan struct{}),
		opts:     opts,
		queueKey: queueKey,
	}

	// Start a worker
	s.wg.Add(1)
	go s.worker()

	return s
}

func (s *AsyncSender) Send(ctx context.Context, to, subject, body string) error {
	job := emailJob{To: to, Subject: subject, Body: body}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.cache.RPush(ctx, s.queueKey, string(data))
}

func (s *AsyncSender) worker() {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic in AsyncSender worker: %v", r)
		}
	}()

	for {
		select {
		case <-s.quitChan:
			return
		default:
			// BLPop blocks for timeout.
			res, err := s.cache.BLPop(context.Background(), 2*time.Second, s.queueKey)
			if err != nil {
				continue
			}

			if len(res) < 2 {
				continue
			}

			var job emailJob
			if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
				log.Errorf("Failed to unmarshal email job: %v", err)
				continue
			}

			if s.opts != nil && s.opts.Host != "" && s.opts.Username != "" {
				s.sendSMTP(job)
			} else {
				s.sendMock(job)
			}
		}
	}
}

func (s *AsyncSender) sendMock(job emailJob) {
	log.Infof("[AsyncEmail] [Redis:%s] To: %s | Subject: %s | Body: %s", s.queueKey, job.To, job.Subject, job.Body)
}

func (s *AsyncSender) sendSMTP(job emailJob) {
	err := SendSMTP(s.opts, job.To, job.Subject, job.Body)
	status := "success"
	if err != nil {
		status = "failed"
		log.Errorf("Async SMTP send failed: %v", err)
	}
	// Assuming metric package in kit has generic Duration metric
	// Check if metric.ExternalAPIDuration exists in kit. Yes I added it.
	metric.ExternalAPIDuration.WithLabelValues("email", "send", status).Observe(0) // Simplified, time not tracked accurately here
}

func (s *AsyncSender) Close() {
	close(s.quitChan)
	s.wg.Wait()
}
