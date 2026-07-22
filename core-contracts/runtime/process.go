package runtime

import (
	"context"
)

type Process struct {
	name     string
	critical bool
	run      func(context.Context) error
}

func Critical(name string, run func(context.Context) error) Process {
	return Process{name: name, critical: true, run: run}
}

func Optional(name string, run func(context.Context) error) Process {
	return Process{name: name, run: run}
}

func (p Process) Name() string {
	return p.name
}

func (p Process) Critical() bool {
	return p.critical
}

func (p Process) Run(ctx context.Context) error {
	return p.run(ctx)
}
