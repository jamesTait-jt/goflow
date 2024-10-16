package goflow

import (
	"context"

	"github.com/jamesTait-jt/goflow/pkg/task"
	"github.com/jamesTait-jt/goflow/pkg/workerpool"
)

type Broker[T task.TaskOrResult] interface {
	task.Submitter[T]
	task.Dequeuer[T]
}

type WorkerPool interface {
	Start(
		ctx context.Context,
		taskQueue task.Dequeuer[task.Task],
		results task.Submitter[task.Result],
		taskHandlers workerpool.HandlerGetter,
	)
	AwaitShutdown()
}

// KVStore defines a key-value store interface in the GoFlow framework.
// It provides methods for storing and retrieving values associated with
// keys.
//
// This interface is generic, allowing for flexibility in key and value
// types. Users can implement the KVStore interface to create custom
// key-value storage solutions tailored to their needs.
//
// Example usage of KVStore could include in-memory storage, database-backed
// storage, or any other form of key-value mapping.
type KVStore[K comparable, V any] interface {
	// Put stores the value associated with the given key.
	Put(k K, v V)

	// Get retrieves the value associated with the given key, returning
	// the value and a boolean indicating whether the key was found.
	Get(k K) (V, bool)
}

// GoFlow orchestrates the execution of tasks within the GoFlow framework.
// It provides a flexible environment for task submission, handling, and
// result management, allowing users to define custom task handlers and
// brokers according to their specific needs.
//
// Users can create an instance of GoFlow using the New function, which
// accepts various options for configuration. The design emphasizes
// extensibility, making it suitable for a wide range of task-processing
// applications.
type GoFlow struct {
	ctx           context.Context
	cancel        context.CancelFunc
	workers       WorkerPool
	taskBroker    Broker[task.Task]
	taskHandlers  KVStore[string, task.Handler]
	resultsBroker Broker[task.Result]
	results       KVStore[string, task.Result]
}

// New creates and initializes a new GoFlow instance with the provided workers
// and optional configuration settings.
//
// The workers parameter specifies the workers that will process tasks within
// the GoFlow framework. The opts variadic parameter allows users to customize
// the GoFlow instance by providing options such as custom task handler stores
// or brokers. The default options are applied if no options are provided.
func New(opts ...Option) *GoFlow {
	options := defaultOptions()

	for _, o := range opts {
		o.apply(&options)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// TODO: Add numWorkers argument
	workerPool := workerpool.New(10)

	gf := GoFlow{
		ctx:           ctx,
		cancel:        cancel,
		workers:       workerPool,
		taskBroker:    options.taskBroker,
		taskHandlers:  options.taskHandlerStore,
		results:       options.resultsStore,
		resultsBroker: options.resultBroker,
	}

	return &gf
}

// Start begins the operation of the GoFlow instance, activating the worker pool
// to start processing tasks from the task broker. It also spawns a goroutine to
// wait for the results of the workers to persist them to the resultsStore.
//
// This method sets the workers in motion, allowing them to listen for tasks
// submitted to the task broker and process them concurrently. Users should call
// Start after configuring the GoFlow instance and registering any task handlers
// to ensure tasks are processed as expected. Although, task handlers can be
// registered on the fly
func (gf *GoFlow) Start() {
	// gf.workers.Start(gf.ctx, gf.taskBroker, gf.resultsBroker, gf.taskHandlers)
	go gf.persistResults(gf.resultsBroker)
}

// RegisterHandler associates a task type with a specific handler function
// within the GoFlow instance. This method allows users to define how tasks
// of a particular type should be processed by providing the appropriate handler.
func (gf *GoFlow) RegisterHandler(taskType string, handler task.Handler) {
	gf.taskHandlers.Put(taskType, handler)
}

// Push submits a new task for processing with the specified task type and payload.
// It looks up the corresponding handler for the task type, returning an error if none is found.
func (gf *GoFlow) Push(taskType string, payload any) (string, error) {

	t := task.New(taskType, payload)

	gf.taskBroker.Submit(gf.ctx, t)

	return t.ID, nil
}

// GetResult retrieves the result associated with a given task ID.
// It returns the result and a boolean indicating whether the result exists.
func (gf *GoFlow) GetResult(taskID string) (task.Result, bool) {
	result, ok := gf.results.Get(taskID)
	return result, ok
}

// Stop terminates the GoFlow instance, signaling all workers to stop processing tasks.
// It cancels the context associated with the GoFlow instance and waits for all workers
// to complete their current work and shut down gracefully. Users should call this method
// when they no longer need the GoFlow instance to process tasks.
func (gf *GoFlow) Stop() {
	gf.cancel()
	gf.workers.AwaitShutdown()
}

func (gf *GoFlow) persistResults(results task.Dequeuer[task.Result]) {
	for {
		select {
		case <-gf.ctx.Done():
			return

		case result := <-results.Dequeue(gf.ctx):
			gf.results.Put(result.TaskID, result)
		}
	}
}
