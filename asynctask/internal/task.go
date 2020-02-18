package internal

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// Task is a measurementkit-like async task.
type Task struct {
	cancel context.CancelFunc
	isdone int64
	out    chan *Event
	wg     *sync.WaitGroup
}

// StartTask stars a new task.
func StartTask(settings *Settings) (*Task, error) {
	if settings == nil {
		return nil, errors.New("passed nil settings")
	}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		cancel: cancel,
		out:    make(chan *Event),
		wg:     wg,
	}
	go func() {
		defer close(task.out)
		defer wg.Done()
		defer atomic.AddInt64(&task.isdone, 1)
		r := newRunner(settings, task.out)
		r.Run(ctx)
	}()
	return task, nil
}

// WaitForNextEvent waits for next event.
func (task *Task) WaitForNextEvent() *Event {
	return <-task.out
}

// IsDone returns whether the task is done.
func (task *Task) IsDone() (done bool) {
	return task.isdone != 0
}

// Interrupt interrupts the task.
func (task *Task) Interrupt() {
	task.cancel()
}
