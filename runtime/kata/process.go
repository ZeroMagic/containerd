/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kata

import (
	"fmt"
	"context"
	"time"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/runtime"
)

// Process implements containerd.Process and containerd.State
type Process struct {
	id string
	t  *Task
}

// ID returns the process id
func (p *Process) ID() string {
	return p.id
}

// State returns the process state
func (p *Process) State(ctx context.Context) (runtime.State, error) {
	return runtime.State{}, fmt.Errorf("process not implmented")
}

// Kill signals a container
func (p *Process) Kill(ctx context.Context, sig uint32, all bool) error {
	return fmt.Errorf("process kill not implmented")
}

// ResizePty resizes the processes pty/console
func (p *Process) ResizePty(ctx context.Context, size runtime.ConsoleSize) error {
	return fmt.Errorf("process resizePty not implmented")
}

// CloseIO closes the processes stdin
func (p *Process) CloseIO(ctx context.Context) error {
	return fmt.Errorf("processs CloseIO not implmented")
}

// Start the container's user defined process
func (p *Process) Start(ctx context.Context) (err error) {
	p.t.events.Publish(ctx, runtime.TaskExecStartedEventTopic, &eventstypes.TaskExecStarted{
		ContainerID: p.t.id,
		Pid:         p.t.pid,
		ExecID:      p.id,
	})
	return nil
}

// Wait for the process to exit
func (p *Process) Wait(ctx context.Context) (*runtime.Exit, error) {
	// init := p.t.processeList[fmt.Sprintf("%d", p.t.pid)]
	// init.Wait()
	return &runtime.Exit{
		Timestamp: time.Time{},
		Status:    uint32(0),
	}, nil
}
