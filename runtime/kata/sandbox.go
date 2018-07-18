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

	"github.com/containerd/containerd/runtime"
	vc "github.com/kata-containers/runtime/virtcontainers"
	"github.com/kata-containers/runtime/virtcontainers/pkg/annotations"
	errors "github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// CreateSandbox creates a kata-runtime sandbox
func CreateSandbox(id string) (*vc.Sandbox, error) {
	// Sets the hypervisor configuration.
	hypervisorConfig := vc.HypervisorConfig{
		KernelParams: []vc.Param{
			vc.Param{
				Key:	"agent.log",
				Value:	"debug",
			},
			vc.Param{
				Key:	"qemu.cmdline",
				Value:	"-D <logfile>",
			},
			vc.Param{
				Key:   "ip",
				Value: "::::::" + id + "::off::",
			},
		},
		KernelPath:     "/usr/share/kata-containers/vmlinuz.container",
		InitrdPath:     "/usr/share/kata-containers/kata-containers-initrd.img",
		HypervisorPath: "/usr/bin/qemu-lite-system-x86_64",

		BlockDeviceDriver: "virtio-scsi",

		HypervisorMachineType: "pc",

		DefaultVCPUs:    uint32(1),
		DefaultMaxVCPUs: uint32(4),

		DefaultMemSz: uint32(128),

		DefaultBridges: uint32(1),

		Mlock:   true,
		Msize9p: uint32(8192),

		Debug: true,
	}

	// Use KataAgent for the agent.
	agConfig := vc.KataAgentConfig{
		LongLiveConn: true,
	}

	// VM resources
	vmConfig := vc.Resources{
		Memory: uint(128),
	}

	// The sandbox configuration:
	// - One container
	// - Hypervisor is QEMU
	// - Agent is KataContainers
	sandboxConfig := vc.SandboxConfig{
		ID: id,

		VMConfig: vmConfig,

		HypervisorType:   vc.QemuHypervisor,
		HypervisorConfig: hypervisorConfig,

		AgentType:   vc.KataContainersAgent,
		AgentConfig: agConfig,

		ProxyType: vc.KataBuiltInProxyType,
		ProxyConfig: vc.ProxyConfig{},

		ShimType: vc.KataBuiltInShimType,
		ShimConfig: vc.ShimConfig{},

		NetworkModel:	vc.CNMNetworkModel,
		NetworkConfig:	vc.NetworkConfig{
			NumInterfaces:		1,
			InterworkingModel:	2,
		},

		// TODO: namespace would be solved
		Annotations: map[string]string{
			annotations.BundlePathKey:	"/run/containerd/io.containerd.runtime.v1.kata-runtime/k8s.io/"+id,
		},

		ShmSize:	uint64(67108864),
		SharePidNs: false,
	}

	sandbox, err := vc.CreateSandbox(sandboxConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create sandbox")
	}

	logrus.FieldLogger(logrus.New()).WithFields(logrus.Fields{
		"sandbox": sandbox,
	}).Info("Create Sandbox Successfully")

	return sandbox.(*vc.Sandbox), err
}

// StartSandbox starts a kata-runtime sandbox
func StartSandbox(id string) (*vc.Sandbox, error) {
	sandbox, err := vc.StartSandbox(id)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not start sandbox")
	}

	return sandbox.(*vc.Sandbox), err
}

// StopSandbox stops a kata-runtime sandbox
func StopSandbox(ctx context.Context, id string, opts runtime.CreateOpts) error {
	return fmt.Errorf("stop not implemented")
}

// DeleteSandbox deletes a kata-runtime sandbox
func DeleteSandbox(ctx context.Context, id string, opts runtime.CreateOpts) error {
	return fmt.Errorf("delete not implemented")
}