package mrun

import (
	"context"
	"fmt"
	"strings"
)

type errShutdown struct{}

type container struct {
	runnable Runnable
	errChan  chan error
	err      error
}

func newContainer(runnable Runnable) *container {
	return &container{
		runnable: runnable,
		errChan:  make(chan error, 1),
	}
}

func (c *container) name() string {
	return strings.TrimLeft(fmt.Sprintf("%T", c.runnable), "*")
}

func (c *container) setError(err error, shutdownChan chan error) {
	c.err = err
	c.errChan <- err
	shutdownChan <- fmt.Errorf("%s: %s", c.name(), err)
}

func (c *container) launch(ctx context.Context, shutdownChan chan error) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				err := fmt.Errorf("panic: %s", p)
				c.setError(err, shutdownChan)
			}
		}()

		err := c.runnable.Run(ctx)
		c.setError(err, shutdownChan)
	}()
}
