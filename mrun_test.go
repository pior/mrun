package mrun_test

import (
	"context"
	"testing"

	"github.com/pior/mrun"

	"github.com/stretchr/testify/require"
)

type ErrTest struct{}

func (e *ErrTest) Error() string { return "ErrTest" }

type mock struct {
	failWithError error
	started       bool
	cancelled     bool
}

func (m *mock) Run(ctx context.Context) error {
	if m.failWithError != nil {
		return m.failWithError
	}
	m.started = true
	<-ctx.Done()
	m.cancelled = true
	return nil
}

func TestMRun(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	sentinelA := &mock{}
	sentinelB := &mock{}
	mr := mrun.New(sentinelA, sentinelB)

	errChan := make(chan error)
	go func() {
		errChan <- mr.Run(ctx)
	}()

	cancelFunc()
	err := <-errChan
	require.NoError(t, err)
	require.True(t, sentinelA.started)
	require.True(t, sentinelB.started)
}

func TestMRun_RunnableFails(t *testing.T) {
	ctx := context.Background()

	testError := &ErrTest{}
	sentinelA := &mock{failWithError: testError}
	sentinelB := &mock{}
	mr := mrun.New(sentinelA, sentinelB)

	errChan := make(chan error)
	go func() {
		errChan <- mr.Run(ctx)
	}()

	err := <-errChan
	require.Equal(t, testError, err)
	require.True(t, sentinelB.cancelled)
}
