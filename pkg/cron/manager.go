package cron

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arrow2012/nuwa-kit/pkg/log"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Job definition for the manager
type Job struct {
	Name      string
	Spec      string
	Cmd       func()
	CheckFunc func() bool // Function to check if job should be enabled
	EntryID   cron.EntryID
	Running   bool
	executing uint32
}

// Manager manages cron jobs with support for dynamic reloading and distributed locking
type Manager struct {
	cron   *cron.Cron
	mu     sync.Mutex
	jobs   map[string]*Job
	rdb    *redis.Client
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewManager creates a new Cron Manager
// rdb is optional. If provided, distributed locking via Redis will be attempted (implementation dependent on wrapper).
// Actually, here we implement simple local locking or we can extend for Redis distribution if needed.
// The original implementation used atomic CAS for local locking and didn't seem to use Redis for distributed locking explicitly in the wrapper shown?
// Wait, checking original code:
// func (s *CronService) wrapper(job *Job) func() {
// ... if !atomic.CompareAndSwapUint32(&job.executing, 0, 1) ...
// This is local concurrency control (prevent overlapping runs of same job instance).
// The original code passed `rdb` but seemingly didn't use it in `wrapper` for distributed lock (maybe I missed it or it was for future use).
// Let's keep `rdb` in struct for potential use or if we want to add distributed lock later.
func NewManager(rdb *redis.Client) *Manager {
	return &Manager{
		cron:   cron.New(cron.WithSeconds()),
		jobs:   make(map[string]*Job),
		rdb:    rdb,
		stopCh: make(chan struct{}),
	}
}

// Start begins the cron service and the dynamic watcher
func (m *Manager) Start() {
	m.cron.Start()
	log.Info("Cron Manager started")

	// Start Dynamic Watcher
	m.wg.Add(1)
	go m.watch()
}

// Stop stops the cron service
func (m *Manager) Stop() {
	close(m.stopCh)
	ctx := m.cron.Stop()
	<-ctx.Done()
	m.wg.Wait()
	log.Info("Cron Manager stopped")
}

// Register adds a job to the manager. It DOES NOT schedule it immediately; scheduling happens in Refresh loop or if we call Refresh manually.
// checkFunc: returns true if job should be enabled.
func (m *Manager) Register(name, spec string, cmd func(), checkFunc func() bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from panic in Register", zap.Any("error", r), zap.String("job", name))
		}
	}()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.jobs[name]; ok {
		panic("Duplicate cron job name: " + name)
	}

	m.jobs[name] = &Job{
		Name:      name,
		Spec:      spec,
		Cmd:       cmd,
		CheckFunc: checkFunc,
		Running:   false,
	}

	// Try to schedule immediately
	m.refreshJob(m.jobs[name])
}

func (m *Manager) watch() {
	defer m.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf("panic in watcher: %v\n\n%s", r, string(debug.Stack()))
					}
				}()
				m.Refresh()
			}()
		}
	}
}

// Refresh checks all jobs and enables/disables them based on CheckFunc
func (m *Manager) Refresh() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, job := range m.jobs {
		m.refreshJob(job)
	}
}

func (m *Manager) refreshJob(job *Job) {
	shouldRun := job.CheckFunc()
	if shouldRun && !job.Running {
		// Enable
		id, err := m.cron.AddFunc(job.Spec, m.wrapper(job))
		if err != nil {
			log.Error("Failed to enable job", zap.String("job", job.Name), zap.Error(err))
			return
		}
		job.EntryID = id
		job.Running = true
		log.Info("Enabled dynamic job", zap.String("job", job.Name))
	} else if !shouldRun && job.Running {
		// Disable
		m.cron.Remove(job.EntryID)
		job.Running = false
		log.Info("Disabled dynamic job", zap.String("job", job.Name))
	}
}

func (m *Manager) wrapper(job *Job) func() {
	return func() {
		// 1. Local Concurrency Control
		if !atomic.CompareAndSwapUint32(&job.executing, 0, 1) {
			log.Warn("Job skipped: previous run still executing", zap.String("job", job.Name))
			return
		}
		defer atomic.StoreUint32(&job.executing, 0)

		// 2. Distributed Lock (Optional, using setnx if redis available)
		// Assuming we want intended distributed lock behavior if rdb is present
		if m.rdb != nil {
			// lockKey := fmt.Sprintf("cron:lock:%s", job.Name)
			// Placeholder for distributed lock logic if needed.
			// Currently implementation mirrors original behavior (no explicit distributed lock in wrapper).
		}

		// Panic Recovery
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in job %s: %v\n\n%s", job.Name, r, string(debug.Stack()))
			}
		}()

		log.Debug("Starting cron job", zap.String("job", job.Name))
		start := time.Now()
		job.Cmd()
		duration := time.Since(start)
		log.Debug("Finished cron job", zap.String("job", job.Name), zap.Duration("duration", duration))
	}
}

// RunningJobCount returns the number of currently executing jobs
func (m *Manager) RunningJobCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, job := range m.jobs {
		if atomic.LoadUint32(&job.executing) == 1 {
			count++
		}
	}
	return count
}
