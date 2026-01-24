package async

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JobType defines the type of async job
type JobType string

const (
	JobTypeCategorySuggestion JobType = "category_suggestion"
	JobTypeNotification       JobType = "notification"
	JobTypeMetricsUpdate      JobType = "metrics_update"
	JobTypeDataExport         JobType = "data_export"
	JobTypeAIParseExpense     JobType = "ai_parse_expense"
)

// JobPriority defines job execution priority
type JobPriority int

const (
	PriorityHigh   JobPriority = 0
	PriorityNormal JobPriority = 1
	PriorityLow    JobPriority = 2
)

// Job represents an async job
type Job struct {
	ID        string
	Type      JobType
	Priority  JobPriority
	Payload   map[string]interface{}
	CreatedAt time.Time
	RetryCount int
	MaxRetries int
	Status    JobStatus
	Error     string
}

// JobStatus represents job status
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// JobQueue manages async jobs with priority queue
type JobQueue struct {
	jobs         map[string]*Job           // Job ID -> Job mapping
	queues       map[JobPriority]chan *Job // Priority queues
	workers      int
	workerPool   chan struct{}             // Semaphore for worker pool
	handlers     map[JobType]JobHandler    // Handlers for different job types
	mu           sync.RWMutex
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	processedCh  chan *Job  // Channel for completed jobs
	errorsCh     chan error // Channel for job errors
	maxJobs      int        // Maximum jobs in queue
	currentJobs  int        // Current number of jobs
	metrics      *JobQueueMetrics
}

// JobQueueMetrics tracks queue metrics
type JobQueueMetrics struct {
	Enqueued   int64
	Processing int64
	Completed  int64
	Failed     int64
	mu         sync.RWMutex
}

// NewJobQueue creates a new job queue with specified worker count
func NewJobQueue(workers int) *JobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	jq := &JobQueue{
		jobs:       make(map[string]*Job),
		queues:     make(map[JobPriority]chan *Job),
		workers:    workers,
		workerPool: make(chan struct{}, workers),
		handlers:   make(map[JobType]JobHandler),
		ctx:        ctx,
		cancel:     cancel,
		processedCh: make(chan *Job, workers*2),
		errorsCh:   make(chan error, workers*2),
		maxJobs:    10000,
		metrics:    &JobQueueMetrics{},
	}

	// Initialize priority queues
	jq.queues[PriorityHigh] = make(chan *Job, 100)
	jq.queues[PriorityNormal] = make(chan *Job, 100)
	jq.queues[PriorityLow] = make(chan *Job, 100)

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		jq.wg.Add(1)
		go jq.worker()
	}

	// Start dispatcher to manage priority queues
	jq.wg.Add(1)
	go jq.dispatcher()

	return jq
}

// RegisterHandler registers a handler for a job type
func (jq *JobQueue) RegisterHandler(jobType JobType, handler JobHandler) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.handlers[jobType] = handler
}

// Enqueue adds a job to the queue
func (jq *JobQueue) Enqueue(job *Job) error {
	jq.mu.Lock()
	if jq.currentJobs >= jq.maxJobs {
		jq.mu.Unlock()
		return fmt.Errorf("job queue is full")
	}
	jq.currentJobs++
	jq.mu.Unlock()

	job.CreatedAt = time.Now()
	job.Status = JobStatusPending
	job.MaxRetries = 3

	jq.mu.Lock()
	jq.jobs[job.ID] = job
	jq.mu.Unlock()

	jq.metrics.mu.Lock()
	jq.metrics.Enqueued++
	jq.metrics.mu.Unlock()

	// Add to appropriate priority queue
	select {
	case jq.queues[job.Priority] <- job:
		return nil
	case <-jq.ctx.Done():
		return fmt.Errorf("job queue is shutting down")
	}
}

// dispatcher manages priority queue processing
func (jq *JobQueue) dispatcher() {
	defer jq.wg.Done()

	for {
		select {
		case <-jq.ctx.Done():
			return

		// Check high priority queue first
		case job := <-jq.queues[PriorityHigh]:
			jq.processJob(job)

		// Then check normal priority (non-blocking)
		default:
			select {
			case job := <-jq.queues[PriorityNormal]:
				jq.processJob(job)
			default:
				// Finally check low priority (non-blocking)
				select {
				case job := <-jq.queues[PriorityLow]:
					jq.processJob(job)
				case <-jq.ctx.Done():
					return
				default:
					time.Sleep(10 * time.Millisecond) // Brief sleep to prevent busy-waiting
				}
			}
		}
	}
}

// processJob sends job to worker
func (jq *JobQueue) processJob(job *Job) {
	// Acquire worker slot
	select {
	case jq.workerPool <- struct{}{}:
		jq.wg.Add(1)
		go func(j *Job) {
			defer jq.wg.Done()
			defer func() { <-jq.workerPool }()

			jq.executeJob(j)
		}(job)
	case <-jq.ctx.Done():
	}
}

// executeJob executes a job with retry logic
func (jq *JobQueue) executeJob(job *Job) {
	job.Status = JobStatusRunning
	jq.metrics.mu.Lock()
	jq.metrics.Processing++
	jq.metrics.mu.Unlock()

	handler, exists := jq.handlers[job.Type]
	if !exists {
		job.Status = JobStatusFailed
		job.Error = fmt.Sprintf("no handler for job type: %s", job.Type)
		jq.metrics.mu.Lock()
		jq.metrics.Failed++
		jq.metrics.mu.Unlock()
		jq.errorsCh <- fmt.Errorf("%s", job.Error)
		return
	}

	err := handler(jq.ctx, job)
	if err != nil {
		if job.RetryCount < job.MaxRetries {
			job.RetryCount++
			job.Status = JobStatusRetrying
			// Re-queue with increased priority
			newPriority := job.Priority - 1
			if newPriority < PriorityHigh {
				newPriority = PriorityHigh
			}
			job.Priority = newPriority

			select {
			case jq.queues[job.Priority] <- job:
			case <-jq.ctx.Done():
			}
		} else {
			job.Status = JobStatusFailed
			job.Error = err.Error()
			jq.metrics.mu.Lock()
			jq.metrics.Failed++
			jq.metrics.mu.Unlock()
			jq.errorsCh <- err
		}
	} else {
		job.Status = JobStatusCompleted
		jq.metrics.mu.Lock()
		jq.metrics.Completed++
		jq.metrics.mu.Unlock()
		jq.processedCh <- job
	}

	jq.metrics.mu.Lock()
	jq.metrics.Processing--
	jq.metrics.mu.Unlock()

	jq.mu.Lock()
	jq.currentJobs--
	jq.mu.Unlock()
}

// worker processes jobs from channels
func (jq *JobQueue) worker() {
	defer jq.wg.Done()

	for {
		select {
		case <-jq.ctx.Done():
			return
		}
	}
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(id string) (*Job, bool) {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	job, exists := jq.jobs[id]
	return job, exists
}

// Size returns current queue size
func (jq *JobQueue) Size() int {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	return jq.currentJobs
}

// Metrics returns queue metrics
func (jq *JobQueue) Metrics() map[string]interface{} {
	jq.metrics.mu.RLock()
	defer jq.metrics.mu.RUnlock()

	return map[string]interface{}{
		"enqueued":     jq.metrics.Enqueued,
		"processing":   jq.metrics.Processing,
		"completed":    jq.metrics.Completed,
		"failed":       jq.metrics.Failed,
		"current_size": jq.currentJobs,
		"max_size":     jq.maxJobs,
		"workers":      jq.workers,
	}
}

// Close gracefully shuts down the job queue
func (jq *JobQueue) Close() error {
	jq.cancel()
	jq.wg.Wait()

	close(jq.processedCh)
	close(jq.errorsCh)

	for priority := range jq.queues {
		close(jq.queues[priority])
	}

	return nil
}

// ProcessedJobs returns the channel for completed jobs
func (jq *JobQueue) ProcessedJobs() <-chan *Job {
	return jq.processedCh
}

// Errors returns the channel for job errors
func (jq *JobQueue) Errors() <-chan error {
	return jq.errorsCh
}
