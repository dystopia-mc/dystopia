package instance

import (
	"context"
	"errors"
	"time"
)

type CustomContext struct {
	context.Context
	cancelFunc func(reason string)
}

func NewCustomContext(parent context.Context, timeout time.Duration) *CustomContext {
	ctx, cancel := context.WithTimeout(parent, timeout)
	customCtx := &CustomContext{Context: ctx}
	customCtx.cancelFunc = func(reason string) {
		cancel()
		*customCtx = *customCtx.WithError(errors.New(reason))
	}
	return customCtx
}

func (c *CustomContext) Cancel(cause string) {
	c.cancelFunc(cause)
}

func (c *CustomContext) WithError(err error) *CustomContext {
	return &CustomContext{
		Context:    context.WithValue(c.Context, "error", err),
		cancelFunc: c.cancelFunc,
	}
}

func (c *CustomContext) Err() error {
	if err, ok := c.Value("error").(error); ok {
		return err
	}
	return c.Context.Err()
}
