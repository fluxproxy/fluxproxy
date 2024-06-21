package common

import (
	"context"
)

type TaskFunc func() error

func LoopTask(ctx context.Context, task TaskFunc) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := task(); err != nil {
				return err
			}
		}
	}
}

// RunTasks executes a list of tasks in parallel, returns the first error encountered or nil if all tasks pass.
func RunTasks(ctx context.Context, tasks ...TaskFunc) error {
	n := len(tasks)
	s := NewSemaphore(n)
	done := make(chan error, 1)

	for _, task := range tasks {
		<-s.Wait()
		go func(taskFunc func() error) {
			err := taskFunc()
			if err == nil {
				s.Signal()
				return
			}
			select {
			case done <- err:
			default:
			}
		}(task)
	}

	for i := 0; i < n; i++ {
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-s.Wait():
		}
	}
	return nil
}
