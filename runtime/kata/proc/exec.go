/*
   Copyright The containerd Authors.

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

package proc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/containerd/console"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

type execProcess struct {
	wg sync.WaitGroup

	State

	mu      sync.Mutex
	id      string
	console console.Console
	io      IO
	exitStatus  int
	exited  time.Time
	pid     int
	closers []io.Closer
	stdin   io.Closer
	stdio   Stdio
	path    string
	spec    specs.Process

	parent    *Init
	waitBlock chan struct{}
}

func (e *execProcess) ID() string {
	return e.id
}

func (e *execProcess) Pid() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.pid
}

func (e *execProcess) ExitStatus() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.exitStatus
}

func (e *execProcess) ExitedAt() time.Time {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.exited
}

func (e *execProcess) Stdin() io.Closer {
	return e.stdin
}

func (e *execProcess) Stdio() Stdio {
	return e.stdio
}

func (e *execProcess) Status(ctx context.Context) (string, error) {
	s, err := e.parent.Status(ctx)
	if err != nil {
		return "", err
	}
	// if the container as a whole is in the pausing/paused state, so are all
	// other processes inside the container, use container state here
	switch s {
	case "paused", "pausing":
		return s, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	// if we don't have a pid then the exec process has just been created
	if e.pid == 0 {
		return "created", nil
	}
	// if we have a pid and it can be signaled, the process is running
	if err := unix.Kill(e.pid, 0); err == nil {
		return "running", nil
	}
	// else if we have a pid but it can nolonger be signaled, it has stopped
	return "stopped", nil
}

func (e *execProcess) Wait() {
	<-e.waitBlock
}

func (e *execProcess) resize(ws console.WinSize) error {
	if e.console == nil {
		return nil
	}
	return e.console.Resize(ws)
}

func (e *execProcess) start(ctx context.Context) error {
	return fmt.Errorf("exec process start is not implemented")
}

func (e *execProcess) delete(ctx context.Context) error {
	e.wg.Wait()
	if e.io != nil {
		for _, c := range e.closers {
			c.Close()
		}
		e.io.Close()
	}
	pidfile := filepath.Join(e.path, fmt.Sprintf("%s.pid", e.id))
	// silently ignore error
	os.Remove(pidfile)
	return nil
}

func (e *execProcess) kill(ctx context.Context, sig uint32, _ bool) error {
	pid := e.pid
	if pid != 0 {
		if err := unix.Kill(pid, syscall.Signal(sig)); err != nil {
			return errors.Wrapf(checkKillError(err), "exec kill error")
		}
	}
	return nil
}

func (e *execProcess) setExited(status int) {
	e.exitStatus = status
	e.exited = time.Now()
	e.parent.platform.ShutdownConsole(context.Background(), e.console)
	close(e.waitBlock)
}