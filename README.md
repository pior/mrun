# MRun

[![GoDoc](https://godoc.org/github.com/pior/mrun?status.svg)](https://godoc.org/github.com/pior/mrun)
[![Go Report Card](https://goreportcard.com/badge/github.com/pior/mrun)](https://goreportcard.com/report/github.com/pior/mrun)

Simplest tool to run and properly shutdown multiple components in the same process.

When your application relies on support components like an event emitter or a trace collector, those components should
be shutdown properly to let them finish their work.

Conversely, an action should be taken when a support component dies prematurely. In such a case, MRun will shutdown all
other components.

## Example:

```go
import (
    "github.com/pior/mrun"
)
```

Define your runnables:
```go
type Server struct {}

func (s *Server) Run(ctx context.Context) error {
    // serve stuff
    <-ctx.Done()
    return nil
}

type EventEmitter struct {}

func (s *EventEmitter) Run(ctx context.Context) error {
    // emit stuff
    <-ctx.Done()
    // FLUSH STUFF !!
    return nil
}
```

Start your application:
```go
func main() {
    mrun.RunAndExit(&Server{}, &EventEmitter{})
}
```

Which is equivalent to:
```go
func main() {
    mr := New(&Server{}, &EventEmitter{})

    mr.SetSignalsHandler()

    ctx := context.Background()
    err := mr.Run(ctx)

    if err != nil {
        mr.Logger.Infof("Error: %s", err)
        os.Exit(1)
    }

    os.Exit(0)
}
```

## License

The MIT License (MIT)
