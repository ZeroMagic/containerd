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
	"sync"
	"time"

	cgroups "github.com/containerd/cgroups"
	eventstypes "github.com/containerd/containerd/api/events"
	exchange "github.com/containerd/containerd/events/exchange"
	log "github.com/containerd/containerd/log"
    "github.com/containerd/containerd/runtime"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	
	"github.com/containerd/containerd/runtime/kata/proc"

	vc "github.com/kata-containers/runtime/virtcontainers"

)

// Task on a hypervisor based system
type Task struct {
	mu sync.Mutex

	id        string
	namespace string
	pid       uint32

    cg        cgroups.Cgroup
    monitor   runtime.TaskMonitor
	events    *exchange.Exchange
	
	processList map[string]proc.Process
}

func newTask(ctx context.Context, id, namespace string, pid uint32, monitor runtime.TaskMonitor, events *exchange.Exchange, opts runtime.CreateOpts, bundle *bundle) (*Task, error) {
	// TODO(ZeroMagic): how to load cgroup when reconnecting
	
	config :=  &proc.InitConfig{
		ID:			id,
		Rootfs:		opts.Rootfs,
		Terminal: 	opts.IO.Terminal,
		Stdin:      opts.IO.Stdin,
		Stdout:     opts.IO.Stdout,
		Stderr:     opts.IO.Stderr,
	}

	log.G(ctx).Infoln("new init process")
	init, err := proc.NewInit(ctx, bundle.path, bundle.workDir, namespace, int(pid), config)
	if err != nil {
		return nil, errors.Wrap(err, "new init process error")
	}

	processList := make(map[string]proc.Process)
	processList[fmt.Sprintf("%d", pid)] = init

	log.G(ctx).Infof("task id is %v, pid is %v  ", id, pid)
	return &Task{
		id:        id,
		pid:       pid,
		namespace: namespace,
		monitor:   monitor,
		events:    events,
		processList: processList,
	}, nil
}

// ID of the task
func (t *Task) ID() string {
	return t.id
}

// Info returns task information about the runtime and namespace
func (t *Task) Info() runtime.TaskInfo {
	return runtime.TaskInfo{
		ID:        t.id,
		Runtime:   pluginID,
		Namespace: t.namespace,
	}
}

// Start the task
func (t *Task) Start(ctx context.Context) error {

    // t.mu.Lock()
	// hasCgroup := t.cg != nil
	// t.mu.Unlock()

	t.processList[fmt.Sprintf("%d", t.pid)].Start(ctx)


	// if !hasCgroup {
	// 	log.G(ctx).Infoln("Task: load cgroups")
	// 	cg, err := cgroups.Load(cgroups.V1, cgroups.PidPath(int(t.pid)))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	t.mu.Lock()
	// 	t.cg = cg
	// 	t.mu.Unlock()
	// 	if err := t.monitor.Monitor(t); err != nil {
	// 		return err
	// 	}
	// }

	log.G(ctx).Infoln("Task: start publishing")
	t.events.Publish(ctx, runtime.TaskStartEventTopic, &eventstypes.TaskStart{
		ContainerID: t.id,
		Pid:         t.pid,
	})
	return nil
}

// State returns runtime information for the task
func (t *Task) State(ctx context.Context) (runtime.State, error) {

	p := t.processList[t.id]
	

	state, err := p.Status(ctx)
	if err != nil {
		return runtime.State{}, errors.Wrap(err, "task state error")
	}

	var status runtime.Status
	switch state {
	case string(vc.StateReady):
		status = runtime.CreatedStatus
	case string(vc.StateRunning):
		status = runtime.RunningStatus
	case string(vc.StatePaused):
		status = runtime.PausedStatus
	case string(vc.StateStopped):
		status = runtime.StoppedStatus
	}

	stdio := p.Stdio()

	return runtime.State{
		Status:     status,
		Pid:        t.pid,
		Stdin:      stdio.Stdin,
		Stdout:     stdio.Stdout,
		Stderr:     stdio.Stderr,
		Terminal:   stdio.Terminal,
		ExitStatus: uint32(p.ExitStatus()),
		ExitedAt:   p.ExitedAt(),
	}, nil
}

// Pause pauses the container process
func (t *Task) Pause(context.Context) error {
    return fmt.Errorf("task pause not implemented")
}

// Resume unpauses the container process
func (t *Task) Resume(context.Context) error {
    return fmt.Errorf("task resume not implemented")
}

// Exec adds a process into the container
func (t *Task) Exec(context.Context, string, runtime.ExecOpts) (runtime.Process, error) {
    return nil, fmt.Errorf("task exec not implemented")
}

// Pids returns all pids
func (t *Task) Pids(context.Context) ([]runtime.ProcessInfo, error) {
    return nil, fmt.Errorf("task pids not implemented")
}

// Checkpoint checkpoints a container to an image with live system data
func (t *Task) Checkpoint(context.Context, string, *types.Any) error {
    return fmt.Errorf("task checkpoint not implemented")
}

// DeleteProcess deletes a specific exec process via its id
func (t *Task) DeleteProcess(context.Context, string) (*runtime.Exit, error) {
    return nil, fmt.Errorf("task delete process not implemented")
}

// Update sets the provided resources to a running task
func (t *Task) Update(context.Context, *types.Any) error {
    return fmt.Errorf("task update not implemented")
}

// Process returns a process within the task for the provided id
func (t *Task) Process(context.Context, string) (runtime.Process, error) {
    return nil, fmt.Errorf("task process not implemented")
}

// Metrics returns runtime specific metrics for a task
func (t *Task) Metrics(context.Context) (interface{}, error) {
    return nil, fmt.Errorf("task metrics not implemented")
}

// CloseIO closes the provided IO on the task
func (t *Task) CloseIO(ctx context.Context) error {
	return fmt.Errorf("task closeIOnot implemented")
}

// Kill the task using the provided signal
func (t *Task) Kill(ctx context.Context, signal uint32, all bool) error {
	p := t.processList[t.id]
	err := p.Kill(ctx, signal, all)
	if err != nil {
		return errors.Wrap(err, "task kill error")
	}
	return nil
}

// ResizePty changes the side of the task's PTY to the provided width and height
func (t *Task) ResizePty(ctx context.Context, size runtime.ConsoleSize) error {
	return fmt.Errorf("task resizePty not implemented")
}

// Wait for the task to exit returning the status and timestamp
func (t *Task) Wait(ctx context.Context) (*runtime.Exit, error) {
	p := t.processList[fmt.Sprintf("%d", t.pid)]
	p.Wait()
    return &runtime.Exit{
		Pid:		t.pid,
		Status: 	uint32(p.ExitStatus()),
		Timestamp:	time.Time{},
	}, nil
}
