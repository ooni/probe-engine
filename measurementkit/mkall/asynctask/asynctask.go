package asynctask

import (
	"context"
	"sync"
	"sync/atomic"
)

// Task is a measurementkit async task.
type Task struct {
	cancel context.CancelFunc
	isdone int64
	out    chan *Event
	wg     *sync.WaitGroup
}

// StartTask stars a new async task.
func StartTask(settings Settings) (*Task, error) {
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
		r := NewRunner(settings, task.out)
		r.Run(ctx)
	}()
	return task, nil
}

// WaitForNextEvent waits for next event.
func (task *Task) WaitForNextEvent() (ev *Event, err error) {
	if task != nil {
		ev = <-task.out
	}
	return
}

// IsDone returns whether the task is done.
func (task *Task) IsDone() (done bool) {
	if task != nil {
		done = task.isdone != 0
	}
	return
}

// Interrupt interrupts the task.
func (task *Task) Interrupt() {
	if task != nil {
		task.cancel()
	}
}
