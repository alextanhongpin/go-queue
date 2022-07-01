package background

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/hibiken/asynq"
)

var (
	ErrDuplicateTask = errors.New("background: duplicate task")
	ErrTaskNotFound  = errors.New("background: task not found")
)

const (
	redisAddr = "127.0.0.1:6379"
)

type Task = asynq.Task

type Option = asynq.Option

type TaskHandler func(ctx context.Context, task *Task) error

type Background struct {
	client *asynq.Client
	server *asynq.Server
	logger *log.Logger
	tasks  map[string]TaskHandler
}

func New() *Background {
	opt := asynq.RedisClientOpt{Addr: redisAddr}
	client := asynq.NewClient(opt)
	server := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	log := log.Default()
	log.SetPrefix("[background]")

	return &Background{
		client: client,
		server: server,
		logger: log,
		tasks:  make(map[string]TaskHandler),
	}
}

func (bg *Background) Close() error {
	return bg.client.Close()
}

func (bg *Background) Enqueue(unk any, opts ...Option) error {
	name := GetTypeName(unk)
	if _, ok := bg.tasks[name]; !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, name)
	}

	payload, err := json.Marshal(unk)
	if err != nil {
		return err
	}

	task := asynq.NewTask(name, payload, opts...)
	info, err := bg.client.Enqueue(task)
	if err != nil {
		return err
	}

	bg.logger.Printf("enqueued task: id=%s queue=%s\n", info.ID, info.Queue)

	return nil
}

func (bg *Background) RegisterTask(name string, handler TaskHandler) error {
	if _, ok := bg.tasks[name]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicateTask, name)
	}

	bg.tasks[name] = handler

	return nil
}

func (bg *Background) Start() error {
	mux := asynq.NewServeMux()

	for name, handler := range bg.tasks {
		mux.HandleFunc(name, handler)
	}

	return bg.server.Run(mux)
}

// RegisterTask registers a task with the task name based on the request struct
// name. Automatically handles unmarshaling of payload too.
func RegisterTask[T any](bg *Background, t T, handler func(ctx context.Context, req T) error) error {
	name := GetTypeName(t)

	task := func(ctx context.Context, task *asynq.Task) error {
		var t T
		if err := json.Unmarshal(task.Payload(), &t); err != nil {

			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		return handler(ctx, t)
	}

	return bg.RegisterTask(name, task)
}

func GetTypeName[T any](unk T) string {
	t := reflect.TypeOf(unk)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.Name()
}
