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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/containerd/cgroups"
	"github.com/containerd/console"
	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events/exchange"
	"github.com/containerd/containerd/runtime"
	"github.com/gogo/protobuf/types"
	vc "github.com/kata-containers/runtime/virtcontainers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/frakti/pkg/kata/proc"
)

// Task on a hypervisor based system
type Task struct {
	mu sync.Mutex

	id        string
	namespace string
	pid       uint32

	cg      cgroups.Cgroup
	monitor runtime.TaskMonitor
	events  *exchange.Exchange

	processList map[string]proc.Process
}

func newTask(ctx context.Context, id, namespace string, pid uint32, monitor runtime.TaskMonitor, events *exchange.Exchange, opts runtime.CreateOpts, bundle *bundle) (*Task, error) {
	// var (
	// 	err error
	// 	cg  cgroups.Cgroup
	// )
	// if pid > 0 {
	// 	cg, err = cgroups.Load(cgroups.V1, cgroups.PidPath(int(pid)))
	// 	if err != nil && err != cgroups.ErrCgroupDeleted {
	// 		return nil, err
	// 	}
	// }
	
	config := &proc.InitConfig{
		ID:       id,
		Rootfs:   opts.Rootfs,
		Terminal: opts.IO.Terminal,
		Stdin:    opts.IO.Stdin,
		Stdout:   opts.IO.Stdout,
		Stderr:   opts.IO.Stderr,
	}

	init, err := proc.NewInit(ctx, bundle.path, bundle.workDir, namespace, int(pid), config)
	if err != nil {
		return nil, errors.Wrap(err, "new init process error")
	}

	processList := make(map[string]proc.Process)
	processList[id] = init

	logrus.FieldLogger(logrus.New()).Info("new Task Successfully")
	//logrus.FieldLogger(logrus.New()).Infof("cgroupsssssss", cg)

	return &Task{
		id:          id,
		pid:         pid,
		namespace:   namespace,
		//cg:			 cg,
		monitor:     monitor,
		events:      events,
		processList: processList,
	}, nil
}

// ID of the task
func (t *Task) ID() string {
	logrus.FieldLogger(logrus.New()).Info("task ID")
	return t.id
}

// Info returns task information about the runtime and namespace
func (t *Task) Info() runtime.TaskInfo {
	logrus.FieldLogger(logrus.New()).Info("task Info")
	return runtime.TaskInfo{
		ID:        t.id,
		Runtime:   pluginID,
		Namespace: t.namespace,
	}
}

// Start the task
func (t *Task) Start(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Info("task Start")

	// t.mu.Lock()
	// hasCgroup := t.cg != nil
	// t.mu.Unlock()

	

	// if !hasCgroup {
	// 	cg, err := cgroups.Load(cgroups.V1, cgroups.PidPath(int(t.pid)))
	// 	if err != nil {
	// 		return errors.Wrap(err, "task start error")
	// 	}
	// 	t.mu.Lock()
	// 	t.cg = cg
	// 	t.mu.Unlock()
	// 	if err := t.monitor.Monitor(t); err != nil {
	// 		return err
	// 	}
	// }

	t.processList[t.id].(*proc.Init).Start(ctx)

	t.events.Publish(ctx, runtime.TaskStartEventTopic, &eventstypes.TaskStart{
		ContainerID: t.id,
		Pid:         t.pid,
	})
	return nil
}

// State returns runtime information for the task
func (t *Task) State(ctx context.Context) (runtime.State, error) {
	logrus.FieldLogger(logrus.New()).Info("task State")

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
	// logrus.FieldLogger(logrus.New()).WithFields(logrus.Fields{
	// 	"Status":     status,
	// 	"Pid":        t.pid,
	// 	"Stdin":      stdio.Stdin,
	// 	"Stdout":     stdio.Stdout,
	// 	"Stderr":     stdio.Stderr,
	// 	"Terminal":   stdio.Terminal,
	// 	"ExitStatus": uint32(p.ExitStatus()),
	// 	"ExitedAt":   p.ExitedAt(),
	// }).Info("Container State Successfully")

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
func (t *Task) Pause(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Info("task Pause")
	p := t.processList[t.id]
	err := p.(*proc.Init).Pause(ctx)
	if err != nil {
		return errors.Wrap(err, "task Pause error")
	}

	return nil
}

// Resume unpauses the container process
func (t *Task) Resume(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Info("task Resume")
	p := t.processList[t.id]
	err := p.(*proc.Init).Resume(ctx)
	if err != nil {
		return errors.Wrap(err, "task Resume error")
	}

	return nil
}

// Exec adds a process into the container
func (t *Task) Exec(ctx context.Context, id string, opts runtime.ExecOpts) (runtime.Process, error) {
	logrus.FieldLogger(logrus.New()).Info("task Exec")
	p := t.processList[t.id]
	conf := &proc.ExecConfig{
		ID:       id,
		Stdin:    opts.IO.Stdin,
		Stdout:   opts.IO.Stdout,
		Stderr:   opts.IO.Stderr,
		Terminal: opts.IO.Terminal,
		Spec:     opts.Spec,
	}
	process, err := p.(*proc.Init).Exec(ctx, id, conf)
	if err != nil {
		return nil, errors.Wrap(err, "task Exec error")
	}
	t.processList[id] = process

	return &Process{
		id: id,
		t:  t,
	}, nil
}

// Pids returns all pids
func (t *Task) Pids(ctx context.Context) ([]runtime.ProcessInfo, error) {
	logrus.FieldLogger(logrus.New()).Info("task Pids")
	return nil, fmt.Errorf("task pids not implemented")
}

// Checkpoint checkpoints a container to an image with live system data
func (t *Task) Checkpoint(ctx context.Context, path string, options *types.Any) error {
	logrus.FieldLogger(logrus.New()).Info("task Checkpoint")
	return fmt.Errorf("task checkpoint not implemented")
}

// DeleteProcess deletes a specific exec process via its id
func (t *Task) DeleteProcess(ctx context.Context, id string) (*runtime.Exit, error) {
	logrus.FieldLogger(logrus.New()).Info("task DeleteProcess")
	p := t.processList[t.id]
	err := p.(*proc.ExecProcess).Delete(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "task DeleteProcess error")
	}

	return &runtime.Exit{
		Pid:       uint32(p.Pid()),
		Status:    uint32(p.ExitStatus()),
		Timestamp: p.ExitedAt(),
	}, nil
}

// Update sets the provided resources to a running task
func (t *Task) Update(ctx context.Context, resources *types.Any) error {
	logrus.FieldLogger(logrus.New()).Info("task Update")
	return fmt.Errorf("task update not implemented")
}

// Process returns a process within the task for the provided id
func (t *Task) Process(ctx context.Context, id string) (runtime.Process, error) {
	logrus.FieldLogger(logrus.New()).Info("task Process")
	p := &Process{
		id: id,
		t:  t,
	}
	if _, err := p.State(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

// Metrics returns runtime specific metrics for a task
func (t *Task) Metrics(ctx context.Context) (interface{}, error) {
	logrus.FieldLogger(logrus.New()).Info("task Metrics")
	p := t.processList[t.id]
	stats, err := p.(*proc.Init).Metrics(ctx)
	if err != nil {
		return stats, errors.Wrap(err, "task Mertrics error")
	}

	return stats, nil
}

// CloseIO closes the provided IO on the task
func (t *Task) CloseIO(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Info("task CloseIO")
	process := t.processList[t.id]
	if stdin := process.Stdin(); stdin != nil {
		if err := stdin.Close(); err != nil {
			return errors.Wrap(err, "close stdin error")
		}
	}
	return nil
}

// Kill the task using the provided signal
func (t *Task) Kill(ctx context.Context, signal uint32, all bool) error {
	logrus.FieldLogger(logrus.New()).Info("task Kill")
	p := t.processList[t.id]
	err := p.Kill(ctx, signal, all)
	if err != nil {
		return errors.Wrap(err, "task kill error")
	}

	return nil
}

// ResizePty changes the side of the task's PTY to the provided width and height
func (t *Task) ResizePty(ctx context.Context, size runtime.ConsoleSize) error {
	logrus.FieldLogger(logrus.New()).Info("task ResizePty")
	ws := console.WinSize{
		Width:  uint16(size.Width),
		Height: uint16(size.Height),
	}

	p := t.processList[t.id]
	err := p.Resize(ws)
	if err != nil {
		return errors.Wrap(err, "task ResizePty error")
	}

	return nil
}

// Wait for the task to exit returning the status and timestamp
func (t *Task) Wait(ctx context.Context) (*runtime.Exit, error) {
	logrus.FieldLogger(logrus.New()).Info("task Wait")
	p := t.processList[t.id]
	p.Wait()
	p.SetExited(0)
	return &runtime.Exit{
		Pid:       t.pid,
		Status:    uint32(p.ExitStatus()),
		Timestamp: time.Time{},
	}, nil
}

// GetProcess gets the specify process
func (t *Task) GetProcess(id string) proc.Process {
	return t.processList[id]
}

// Cgroup returns the underlying cgroup for a linux task
func (t *Task) Cgroup() (cgroups.Cgroup, error) {
	logrus.FieldLogger(logrus.New()).Infof("task %v Cgroup", t.id)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cg == nil {
		return nil, errors.New("cgroup does not exist")
	}
	return t.cg, nil
}
