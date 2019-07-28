package main

import (
	"context"
	"errors"
	"time"

	"github.com/pior/mrun"
)

type ServerNoShutdown struct{}

func (s *ServerNoShutdown) Run(ctx context.Context) error {
	<-make(chan struct{})
	return nil
}

type ServerPanic struct{}

func (s *ServerPanic) Run(ctx context.Context) error {
	time.Sleep(time.Second * 1)
	panic("yooooolooooooo")
}

type Server struct {
	deadline time.Duration
}

func (s *Server) Run(ctx context.Context) error {
	if s.deadline.Seconds() != 0 {
		ctx, cancel := context.WithTimeout(ctx, s.deadline)
		defer cancel()
		<-ctx.Done()
		return errors.New("sepuku")
	}
	<-ctx.Done()
	return nil
}

func main() {
	mrun.RunAndExit(
		// &Server{time.Second * 3},
		&Server{},
		&ServerNoShutdown{},
		&ServerPanic{},
	)
}
