package common

import "context"

// RunTasks executes a list of tasks in parallel, returns the first error encountered or nil if all tasks pass.
func RunTasks(ctx context.Context, tasks ...func() error) error {
	n := len(tasks)
	s := NewSemaphore(n)
	done := make(chan error, 1)

	for _, task := range tasks {
		<-s.Wait()
		go func(f func() error) {
			err := f()
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
