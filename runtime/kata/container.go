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
	"io/ioutil"
	"strings"
	"sync"
	"syscall"

	"github.com/containerd/containerd/runtime"
	"github.com/containerd/fifo"
	vc "github.com/kata-containers/runtime/virtcontainers"
	"github.com/kata-containers/runtime/virtcontainers/pkg/annotations"
	errors "github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, 32<<10)
		return &buffer
	},
}

// CreateContainer creates a kata-runtime container
func CreateContainer(id, sandboxID string) (*vc.Sandbox, *vc.Container, error) {
	envs := []vc.EnvVar{
		{
			Var:   "PATH",
			Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
	}

	cmd := vc.Cmd{
		Args:    strings.Split("ls", " "),
		Envs:    envs,
		WorkDir: "/",
		User:	"0",
		PrimaryGroup:	"0",
		NoNewPrivileges: true,
	}

	configFile := "/run/containerd/io.containerd.runtime.v1.kata-runtime/k8s.io/"+id+"/config.json"
	configJ, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Print(err)
	}
	str := string(configJ)
	str = strings.Replace(str, "bounding", "Bounding", -1)
	str = strings.Replace(str, "effective", "Effective", -1)
	str = strings.Replace(str, "inheritable", "Inheritable", -1)
	str = strings.Replace(str, "permitted", "Permitted", -1)
	str = strings.Replace(str, "true", "true,\"Ambient\":null", -1)

	// TODO: namespace would be solved
	containerConfig := vc.ContainerConfig{
		ID:     id,
		RootFs: "/run/containerd/io.containerd.runtime.v1.kata-runtime/k8s.io/" + id + "/rootfs",
		Cmd:    cmd,
		Annotations: map[string]string{
			annotations.ConfigJSONKey:	str,
			annotations.BundlePathKey:	"/run/containerd/io.containerd.runtime.v1.kata-runtime/k8s.io/"+id,
			annotations.ContainerTypeKey:	 string(vc.PodContainer),
		},
	}

	logrus.FieldLogger(logrus.New()).Info("##### config.json #####", str)
	logrus.FieldLogger(logrus.New()).Info("##### Creating Container #####")

	sandbox, container, err := vc.CreateContainer(sandboxID, containerConfig)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Could not create container")
	}

	logrus.FieldLogger(logrus.New()).WithFields(logrus.Fields{
		"container": container,
	}).Info("Create Container Successfully")

	return sandbox.(*vc.Sandbox), container.(*vc.Container), err
}

// StartContainer starts a kata-runtime container
func StartContainer(ctx context.Context, id, sandboxID string, t *Task) (*vc.Container, error) {
	container, err := vc.StartContainer(sandboxID, id)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not start container")
	}

	stdin, stdout, stderr, err := t.sandbox.IOStream(sandboxID, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a container's stdio streams from kata")
	}
	t.stdin = stdin
	t.stdout = stdout
	t.stderr = stderr

	// create local stdio
	var (
		localStdout io.WriteCloser//, localStderr io.WriteCloser
		localStdin           io.ReadCloser
		wg               sync.WaitGroup
	)

	if stdin != nil {
		localStdin, err = fifo.OpenFifo(ctx, t.stdio.Stdin, syscall.O_RDONLY, 0)
		if err != nil {
			logrus.FieldLogger(logrus.New()).Errorf("open local stdin: %s", err)
		}
		wg.Add(1)
	}

	if stdout != nil {
		localStdout, err = fifo.OpenFifo(ctx, t.stdio.Stdout, syscall.O_WRONLY, 0)
		if err != nil {
			logrus.FieldLogger(logrus.New()).Errorf("open local stdout: %s", err)
		}
		wg.Add(1)
	}

	// if stderr != nil {
	// 	localStderr, err = fifo.OpenFifo(ctx, config.Stderr, syscall.O_WRONLY, 0)
	// 	if err != nil {
	// 		logrus.FieldLogger(logrus.New()).Errorf("open local stderr: %s", err)
	// 	}
	// 	wg.Add(1)
	// }

	// Connect stdin of container.
	go func() {
		if stdin == nil {
			return
		}
		logrus.FieldLogger(logrus.New()).Info("stdin: begin")
		buf := bufPool.Get().(*[]byte)
		defer bufPool.Put(buf)
		_, err = io.CopyBuffer(stdin, localStdin, *buf)
		if err == io.ErrClosedPipe {
			err = nil
		}
		if err != nil {
			logrus.FieldLogger(logrus.New()).Errorf("stdin copy: %s", err)
		}
		
		logrus.FieldLogger(logrus.New()).Info("stdin: end")
		wg.Done()
	}()

	// stdout/stderr
	attachStream := func(name string, stream io.Writer, streamPipe io.Reader) {
		if stream == nil {
			return
		}

		logrus.FieldLogger(logrus.New()).Infof("%s: begin", name)
		buf := bufPool.Get().(*[]byte)
		defer bufPool.Put(buf)

		_, err := io.CopyBuffer(stream, streamPipe, *buf)
		if err == io.ErrClosedPipe {
			err = nil
		}
		if err != nil {
			logrus.FieldLogger(logrus.New()).Errorf("%s copy: %v", name, err)
		}
		
		if localStdin != nil {
			localStdin.Close()
		}

		if closer, ok := stream.(io.Closer); ok {
			closer.Close()
		}
		logrus.FieldLogger(logrus.New()).Infof("%s: end", name)
		wg.Done()
	}

	go attachStream("stdout", localStdout, stdout)
	// go attachStream("stderr", localStderr, stderr)

	wg.Wait()

	return container.(*vc.Container), err
}

// StopContainer stops a kata-runtime container
func StopContainer(ctx context.Context, id string, opts runtime.CreateOpts) error {
	return fmt.Errorf("stop container not implemented")
}

// DeleteContainer deletes a kata-runtime container
func DeleteContainer(ctx context.Context, id string, opts runtime.CreateOpts) error {
	return fmt.Errorf("delete container not implemented")
}

// KillContainer kills one or more kata-runtime containers
func KillContainer(sandboxID, containerID string, signal syscall.Signal, all bool) error {
	err := vc.KillContainer(sandboxID, containerID, signal, all)
	if err != nil {
		return errors.Wrapf(err, "Could not kill container")
	}

	return nil
}

// StatusContainer returns the virtcontainers container status.
func StatusContainer(sandboxID, containerID string) (vc.ContainerStatus, error) {
	status, err := vc.StatusContainer(sandboxID, containerID)
	if err != nil {
		return vc.ContainerStatus{}, errors.Wrapf(err, "Could not kill container")
	}

	return status, nil
}
