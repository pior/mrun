package mrun

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Runnable can be any long running components that respect context cancellation and returns an error.
type Runnable interface {
	Run(context.Context) error
}

// MRun orchestrates the execution of a set of Runnable.
type MRun struct {
	containers   []*container
	shutdownChan chan error
	gracePeriod  time.Duration
	Logger       Logger
}

// New returns a MRun instance with a set of runnables.
func New(runnables ...Runnable) *MRun {
	var containers []*container
	for _, r := range runnables {
		containers = append(containers, newContainer(r))
	}

	return &MRun{
		containers:   containers,
		shutdownChan: make(chan error, len(runnables)+2),
		gracePeriod:  time.Second * 3,
		Logger:       &standardLogger{},
	}
}

// Run runs all the runnables and manages the graceful shutdown.
func (m *MRun) Run(ctx context.Context) error {
	ctx, cancelFunc := context.WithCancel(ctx)

	for _, c := range m.containers {
		c.launch(ctx, m.shutdownChan)
		m.Logger.Infof("%s: started", c.name)
	}
	m.Logger.Infof("running!")

	shutdownReason := <-m.shutdownChan
	m.Logger.Infof("shutdown! (reason: %s)", shutdownReason)

	cancelFunc()

	m.Logger.Infof("waiting %s for runnables to shutdown", m.gracePeriod)
	deadline := time.After(m.gracePeriod)

	for _, c := range m.containers {
		select {
		case err := <-c.errChan:
			m.Logger.Infof("%s: stopped with: %s", c.name, err)
		case <-deadline:
			m.Logger.Warnf("%s: did not stop during the grace period", c.name)
		}
	}

	for _, c := range m.containers {
		if c.err != nil {
			return c.err
		}
	}
	return nil
}

// SetSignalsHandler installs the signal handler that trigger a shutdown. The signals defaults to SIGINT and SIGTERM.
func (m *MRun) SetSignalsHandler(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = append(signals, syscall.SIGINT)
		signals = append(signals, syscall.SIGTERM)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	go func() {
		defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		m.shutdownChan <- fmt.Errorf("received signal %s", sig)
	}()
}

// RunAndExit runs the set of runnables with MRun, listens to SIGTERM/SIGINT and terminates the process with a non-zero
// code when a runnable fails.
func RunAndExit(runnables ...Runnable) {
	ctx := context.Background()

	mr := New(runnables...)
	mr.SetSignalsHandler()

	err := mr.Run(ctx)
	if err != nil {
		mr.Logger.Infof("Error: %s", err)
		os.Exit(1)
	}

	os.Exit(0)
}
