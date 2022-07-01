package background

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type BackgroundMock struct {
	tasks map[string]TaskHandler
}

func NewMock() *BackgroundMock {
	return &BackgroundMock{
		tasks: make(map[string]TaskHandler),
	}
}

func (bg *BackgroundMock) RegisterTask(name string, handler TaskHandler) error {
	if _, ok := bg.tasks[name]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicateTask, name)
	}

	bg.tasks[name] = handler

	return nil
}

func (bg *BackgroundMock) Enqueue(unk any, opts ...Option) error {
	name := GetTypeName(unk)
	if _, ok := bg.tasks[name]; !ok {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, name)
	}

	payload, err := json.Marshal(unk)
	if err != nil {
		return err
	}

	task := asynq.NewTask(name, payload, opts...)
	return bg.tasks[name](context.Background(), task)
}

func (bg *BackgroundMock) Start() error {
	return nil
}

func (bg *BackgroundMock) Close() error {
	return nil
}
