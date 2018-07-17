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
	"io"
	"sync"
	"syscall"
	"time"

	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events/exchange"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/cri/pkg/annotations"
	"github.com/gogo/protobuf/types"
	vc "github.com/kata-containers/runtime/virtcontainers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Stdio of a task
type Stdio struct {
	Stdin    string
	Stdout   string
	Stderr   string
	Terminal bool
}



// Task on a hypervisor based system
type Task struct {
	mu sync.Mutex

	id        string
	pid       uint32
	namespace string

	containerType string
	sandboxID     string

	container	*vc.Container
	sandbox 	*vc.Sandbox

	stdin      io.WriteCloser
	stdout     io.Reader
	stderr     io.Reader
	stdio      Stdio

	exitCode	uint32

	monitor runtime.TaskMonitor
	events  *exchange.Exchange
}

func newTask(id string, pid uint32, namespace string, monitor runtime.TaskMonitor, events *exchange.Exchange) (*Task, error) {
	return &Task{
		id:          id,
		pid:         pid,
		namespace:   namespace,
		monitor:     monitor,
		events:      events,
	}, nil
}

// ID of the task
func (t *Task) ID() string {
	return t.id
}

// Info returns task information about the runtime and namespace
func (t *Task) Info() runtime.TaskInfo {
	// logrus.FieldLogger(logrus.New()).Info("task Info")
	return runtime.TaskInfo{
		ID:        t.id,
		Runtime:   pluginID,
		Namespace: t.namespace,
	}
}

// Start the task
func (t *Task) Start(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Infof("[Task] %s Start", t.id)

	var err error

	if t.containerType == annotations.ContainerTypeSandbox {
		t.sandbox, err = StartSandbox(t.id)
		if err != nil {
			return errors.Wrapf(err, "task start error")
		}
	} else if t.containerType == annotations.ContainerTypeContainer {
		t.container, err = StartContainer(ctx, t.id, t.sandboxID, t)
		if err != nil {
			return errors.Wrapf(err, "task start error")
		}
	} else {
		return errors.New(ErrContainerType)
	}

	t.events.Publish(ctx, runtime.TaskStartEventTopic, &eventstypes.TaskStart{
		ContainerID: t.id,
		Pid:         t.pid,
	})
	return nil
}

// State returns runtime information for the task
func (t *Task) State(ctx context.Context) (runtime.State, error) {
	logrus.FieldLogger(logrus.New()).Infof("[Task] %s State", t.id)

	var taskState string

	if t.containerType == annotations.ContainerTypeSandbox {
		sandboxStatus, err := vc.StatusSandbox(t.id)
		if err != nil {
			return runtime.State{}, errors.Wrapf(err, "Could not get sandbox's status")
		}
		taskState = string(sandboxStatus.State.State)
	} else if t.containerType == annotations.ContainerTypeContainer {
		containerStatus, err := vc.StatusContainer(t.sandboxID, t.id)
		if err != nil {
			return runtime.State{}, errors.Wrapf(err, "Could not get container's status")
		}
		taskState = string(containerStatus.State.State)
	} else {
		return runtime.State{}, errors.New(ErrContainerType)
	}

	var status runtime.Status
	switch taskState {
	case string(vc.StateReady):
		status = runtime.CreatedStatus
	case string(vc.StateRunning):
		status = runtime.RunningStatus
	case string(vc.StatePaused):
		status = runtime.PausedStatus
	case string(vc.StateStopped):
		status = runtime.StoppedStatus
	}

	return runtime.State{
		Status:     status,
		Pid:        t.pid,
		ExitStatus: t.exitCode,
		ExitedAt:   time.Now(),
		Stdin:      t.stdio.Stdin,
		Stdout:     t.stdio.Stdout,
		Stderr:     t.stdio.Stderr,
		Terminal:   t.stdio.Terminal,
	}, nil
}

// Pause pauses the container process
func (t *Task) Pause(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Infof("[Task] %s Pause", t.id)
	if t.containerType == annotations.ContainerTypeSandbox {
		err := t.sandbox.Pause()
		if err != nil {
			return errors.Wrapf(err, "task pause error")
		}
	} else if t.containerType == annotations.ContainerTypeContainer {
		err := vc.PauseContainer(t.sandboxID, t.id)
		if err != nil {
			return errors.Wrapf(err, "task pause error")
		}
	} else {
		return errors.New(ErrContainerType)
	}

	return nil
}

// Resume unpauses the container process
func (t *Task) Resume(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Info("[Task] %s Resume", t.id)

	if t.containerType == annotations.ContainerTypeSandbox {
		err := t.sandbox.Resume()
		if err != nil {
			return errors.Wrapf(err, "task resume error")
		}
	} else if t.containerType == annotations.ContainerTypeContainer {
		err := vc.ResumeContainer(t.sandboxID, t.id)
		if err != nil {
			return errors.Wrapf(err, "task resume error")
		}
	} else {
		return errors.New(ErrContainerType)
	}

	return nil
}

// Exec adds a process into the container
func (t *Task) Exec(ctx context.Context, id string, opts runtime.ExecOpts) (runtime.Process, error) {
	logrus.FieldLogger(logrus.New()).Info("task Exec")
	// p := t.processList[t.id]
	// conf := &proc.ExecConfig{
	// 	ID:       id,
	// 	Stdin:    opts.IO.Stdin,
	// 	Stdout:   opts.IO.Stdout,
	// 	Stderr:   opts.IO.Stderr,
	// 	Terminal: opts.IO.Terminal,
	// 	Spec:     opts.Spec,
	// }
	// process, err := p.(*proc.Init).Exec(ctx, id, conf)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "task Exec error")
	// }
	// t.processList[id] = process

	// return &Process{
	// 	id: id,
	// 	t:  t,
	// }, nil
	return nil, errors.New("task Exec error")
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
	// p := t.processList[t.id]
	// err := p.(*proc.ExecProcess).Delete(ctx)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "task DeleteProcess error")
	// }

	// return &runtime.Exit{
	// 	Pid:       uint32(p.Pid()),
	// 	Status:    uint32(p.ExitStatus()),
	// 	Timestamp: p.ExitedAt(),
	// }, nil
	return nil, errors.New("task DeleteProcess error")
}

// Update sets the provided resources to a running task
func (t *Task) Update(ctx context.Context, resources *types.Any) error {
	logrus.FieldLogger(logrus.New()).Info("task Update")
	return fmt.Errorf("task update not implemented")
}

// Process returns a process within the task for the provided id
func (t *Task) Process(ctx context.Context, id string) (runtime.Process, error) {
	logrus.FieldLogger(logrus.New()).Info("task Process")
	// p := &Process{
	// 	id: id,
	// 	t:  t,
	// }
	// if _, err := p.State(ctx); err != nil {
	// 	return nil, err
	// }
	return nil, nil
}

// Metrics returns runtime specific metrics for a task
func (t *Task) Metrics(ctx context.Context) (interface{}, error) {
	logrus.FieldLogger(logrus.New()).Info("[Task] %s Metrics", t.id)

	stats, err := vc.StatsContainer(t.sandboxID, t.id)
	if err != nil {
		return vc.ContainerStats{}, errors.Wrap(err, "failed to get the stats of a container")
	}

	return stats, nil
}

// CloseIO closes the provided IO on the task
func (t *Task) CloseIO(ctx context.Context) error {
	logrus.FieldLogger(logrus.New()).Info("[Task] %s CloseIO", t.id)
	
	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			return errors.Wrap(err, "close stdin error")
		}
	}

	return nil
}

// Kill the task using the provided signal
func (t *Task) Kill(ctx context.Context, signal uint32, all bool) error {
	logrus.FieldLogger(logrus.New()).Infof("[Task] %s Kill", t.id)

	if t.containerType == annotations.ContainerTypeSandbox {
		vcSandbox, err := vc.StopSandbox(t.sandboxID)
		if err != nil {
			return errors.Wrapf(err, "task kill error")
		}
		t.sandbox = vcSandbox.(*vc.Sandbox)
	} else if t.containerType == annotations.ContainerTypeContainer {
		err := vc.KillContainer(t.sandboxID, t.id, syscall.Signal(signal), all)
		if err != nil {
			return errors.Wrap(err, "task kill error")
		}

		_, err = vc.StopContainer(t.sandboxID, t.id)
		if err != nil {
			errors.Wrap(err, "failed to stop container")
			return err
		}
	} else {
		return errors.New(ErrContainerType)
	}
	logrus.FieldLogger(logrus.New()).Infof("[Task] %s Kill END", t.id)

	return nil
}

// ResizePty changes the side of the task's PTY to the provided width and height
func (t *Task) ResizePty(ctx context.Context, size runtime.ConsoleSize) error {
	logrus.FieldLogger(logrus.New()).Info("task ResizePty")
	// ws := console.WinSize{
	// 	Width:  uint16(size.Width),
	// 	Height: uint16(size.Height),
	// }

	// p := t.processList[t.id]
	// err := p.Resize(ws)
	// if err != nil {
	// 	return errors.Wrap(err, "task ResizePty error")
	// }

	return nil
}

// Wait for the task to exit returning the status and timestamp
func (t *Task) Wait(ctx context.Context) (*runtime.Exit, error) {
	logrus.FieldLogger(logrus.New()).Info("[Task] %s Wait", t.id)

	// TODO: here we need to identify a exec process and a container process
	exitCode, err := t.sandbox.WaitProcess(t.id, t.id)
	if err != nil {
		return &runtime.Exit{}, err
	}

	t.exitCode = uint32(exitCode)
	
	return &runtime.Exit{
		Pid:       t.pid,
		Status:    t.exitCode,
		Timestamp: time.Now(),
	}, nil
}
