package worker

import (
	"fmt"
	"sort"
	"time"

	models "github.com/ananduee/raker/core/proto"
	"github.com/ananduee/raker/core/storage"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// Worker is core background worker
type Worker struct {
	storage    *storage.Storage
	jobMap     map[string]Job
	stopSignal chan struct{}
}

// Task represents a type of job which we need to do
type Task struct {
	Name string
}

// WorkerJob contains main logic which needs to be executed periodically.
type Job interface {
	Execute() error
}

// New creates new instance of worker
func New() *Worker {
	return &Worker{stopSignal: make(chan struct{}), jobMap: make(map[string]Job)}
}

// Start the job processor worker.
func (worker *Worker) Start() {
	worker.startWorker()
}

// Stop the job processor worker. It will wait for any running task to finish.
func (worker *Worker) Stop() {
	worker.stopSignal <- struct{}{}
}

// AddPeriodicTask registers a task which will be invoked after each period.
func (worker *Worker) AddPeriodicTask(name string, jobFunction Job, period time.Duration) error {
	// save this periodic task in db.
	dbKey := fmt.Sprintf("jobs:details:%s", name)
	_, err := worker.storage.Get(dbKey)
	if err == storage.ErrorKeyNotFound {
		// key is missing we should save this to db.
		workerJob := &models.WorkerJob{
			Version: 1,
			Task: &models.Task{
				Version: 1,
				Name:    name,
			},
			Status: models.WorkerJobStatus_NOT_STARTED,
			Period: ptypes.DurationProto(period),
		}
		outBytes, err := proto.Marshal(workerJob)
		if err != nil {
			return err
		}
		err = worker.storage.Put(dbKey, outBytes)
	}
	if err != nil {
		worker.jobMap[name] = jobFunction
	}
	return err
}

func (worker *Worker) startWorker() {
	for {
		jobs, err := worker.getAllScheduledJobs()
		if err != nil {
			// do something here we need to send it to top.
		}
		jobToExecute := getHighestPriorityJob(jobs)
		ticker := getWaitTimerForJobExecution(jobToExecute)
		select {
		case <-ticker:
			if jobToExecute != nil {
				// We are currently executing task in same goroutine ideally it should be another goroutine as it
				// can impact timing SLA. This will be improved later.
				worker.executeJob(jobToExecute)
			}
			continue
		case <-worker.stopSignal:
			return
		}
	}
}

func (worker *Worker) getAllScheduledJobs() ([]*models.WorkerJob, error) {
	jobsBytes, err := worker.storage.GetAllByPrefix("jobs:details:")
	if err != nil {
		return nil, err
	}
	jobs := make([]*models.WorkerJob, len(jobsBytes))
	for _, jobByte := range jobsBytes {
		var job models.WorkerJob
		err = proto.Unmarshal(jobByte, &job)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &job)
	}
	return jobs, nil
}

func (worker *Worker) executeJob(jobToExecute *models.WorkerJob) {
	job := worker.jobMap[jobToExecute.Task.Name]
	job.Execute()
}

// Find a single task which can be executed. This is not the most efficient approach
// as we are fetching all tasks and then sorting but this will work for now.
func getHighestPriorityJob(jobs []*models.WorkerJob) *models.WorkerJob {
	if jobs == nil || len(jobs) == 0 {
		return nil
	}
	sort.Slice(jobs, func(i, j int) bool {
		firstJob := jobs[i]
		secondJob := jobs[j]

		if firstJob.Next == nil && secondJob.Next == nil {
			// If task has never been executed then pick first task alphabetically.
			return firstJob.Task.Name < secondJob.Task.Name
		} else if firstJob.Next == nil {
			return true
		} else if secondJob.Next == nil {
			return false
		}

		return firstJob.Next.Seconds < secondJob.Next.Seconds
	})
	return jobs[0]
}

// getWaitTimerForJobExecution creates timer object after which a job should be sent for execution.
func getWaitTimerForJobExecution(jobToExecute *models.WorkerJob) <-chan time.Time {
	if jobToExecute == nil {
		// Wait for 5 minutes if no job was registered.
		return time.After(time.Minute * 5)
	}
	if jobToExecute.Next == nil {
		return time.After(time.Nanosecond)
	}
	timeDiff := jobToExecute.Next.Nanos - int32(time.Now().Nanosecond())
	if timeDiff <= 0 {
		return time.After(time.Nanosecond)
	}
	return time.After(time.Duration(timeDiff))
}
