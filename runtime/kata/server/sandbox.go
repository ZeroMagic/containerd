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

package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd/runtime"
	log "github.com/containerd/containerd/log"
	errors "github.com/pkg/errors"

	vc "github.com/kata-containers/runtime/virtcontainers"
)

// CreateSandbox creates a kata-runtime sandbox
func CreateSandbox(ctx context.Context, id string) error {
	envs := []vc.EnvVar{
		{
			Var:   "PATH",
			Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
		{
			Var:   "PATH",
			Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		},
	}

	cmd := vc.Cmd{
		Args:    		strings.Split("sh", " "),
		Envs:    		envs,
		User:			"0",
		PrimaryGroup:	"0",
		WorkDir: 		"/",
		Capabilities:	vc.LinuxCapabilities{
			Bounding:	[]string{
				"CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE",
			},
			Effective:	[]string{
				"CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE",
			},
			Inheritable:	[]string{
				"CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE",
			},
			Permitted:	[]string{
				"CAP_CHOWN",
                "CAP_DAC_OVERRIDE",
                "CAP_FSETID",
                "CAP_FOWNER",
                "CAP_MKNOD",
                "CAP_NET_RAW",
                "CAP_SETGID",
                "CAP_SETUID",
                "CAP_SETFCAP",
                "CAP_SETPCAP",
                "CAP_NET_BIND_SERVICE",
                "CAP_SYS_CHROOT",
                "CAP_KILL",
                "CAP_AUDIT_WRITE",
			},
		},
		NoNewPrivileges:	true,
	}

	// Define the container command and bundle.
	container := vc.ContainerConfig{
		ID:     	id,
		RootFs: 	"/run/containerd/io.containerd.runtime.v1.kata-runtime/default/"+id+"/rootfs",
		Cmd:    	cmd,
		Annotations: map[string]string{
			"com.github.containers.virtcontainers.pkg.oci.config":	"{\"ociVersion\":\"1.0.1\",\"root\":{\"path\":\"rootfs\"},\"mounts\":[{\"destination\":\"/proc\",\"type\":\"proc\",\"source\":\"proc\"},{\"destination\":\"/dev\",\"type\":\"tmpfs\",\"source\":\"tmpfs\",\"options\":[\"nosuid\",\"strictatime\",\"mode=755\",\"size=65536k\"]},{\"destination\":\"/dev/pts\",\"type\":\"devpts\",\"source\":\"devpts\",\"options\":[\"nosuid\",\"noexec\",\"newinstance\",\"ptmxmode=0666\",\"mode=0620\",\"gid=5\"]},{\"destination\":\"/dev/shm\",\"type\":\"tmpfs\",\"source\":\"shm\",\"options\":[\"nosuid\",\"noexec\",\"nodev\",\"mode=1777\",\"size=65536k\"]},{\"destination\":\"/dev/mqueue\",\"type\":\"mqueue\",\"source\":\"mqueue\",\"options\":[\"nosuid\",\"noexec\",\"nodev\"]},{\"destination\":\"/sys\",\"type\":\"sysfs\",\"source\":\"sysfs\",\"options\":[\"nosuid\",\"noexec\",\"nodev\",\"ro\"]},{\"destination\":\"/run\",\"type\":\"tmpfs\",\"source\":\"tmpfs\",\"options\":[\"nosuid\",\"strictatime\",\"mode=755\",\"size=65536k\"]}],\"linux\":{\"resources\":{\"devices\":[{\"allow\":false,\"access\":\"rwm\"}]},\"cgroupsPath\":\"/default/"+id+"\",\"namespaces\":[{\"type\":\"pid\"},{\"type\":\"ipc\"},{\"type\":\"uts\"},{\"type\":\"mount\"},{\"type\":\"network\"}],\"maskedPaths\":[\"/proc/kcore\",\"/proc/latency_stats\",\"/proc/timer_list\",\"/proc/timer_stats\",\"/proc/sched_debug\",\"/sys/firmware\",\"/proc/scsi\"],\"readonlyPaths\":[\"/proc/asound\",\"/proc/bus\",\"/proc/fs\",\"/proc/irq\",\"/proc/sys\",\"/proc/sysrq-trigger\"]},\"process\":{\"user\":{\"uid\":0,\"gid\":0},\"args\":[\"sh\"],\"env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"cwd\":\"/\",\"rlimits\":[{\"type\":\"RLIMIT_NOFILE\",\"hard\":1024,\"soft\":1024}],\"noNewPrivileges\":true,\"capabilities\":{\"bounding\":[\"CAP_CHOWN\",\"CAP_DAC_OVERRIDE\",\"CAP_FSETID\",\"CAP_FOWNER\",\"CAP_MKNOD\",\"CAP_NET_RAW\",\"CAP_SETGID\",\"CAP_SETUID\",\"CAP_SETFCAP\",\"CAP_SETPCAP\",\"CAP_NET_BIND_SERVICE\",\"CAP_SYS_CHROOT\",\"CAP_KILL\",\"CAP_AUDIT_WRITE\"],\"effective\":[\"CAP_CHOWN\",\"CAP_DAC_OVERRIDE\",\"CAP_FSETID\",\"CAP_FOWNER\",\"CAP_MKNOD\",\"CAP_NET_RAW\",\"CAP_SETGID\",\"CAP_SETUID\",\"CAP_SETFCAP\",\"CAP_SETPCAP\",\"CAP_NET_BIND_SERVICE\",\"CAP_SYS_CHROOT\",\"CAP_KILL\",\"CAP_AUDIT_WRITE\"],\"inheritable\":[\"CAP_CHOWN\",\"CAP_DAC_OVERRIDE\",\"CAP_FSETID\",\"CAP_FOWNER\",\"CAP_MKNOD\",\"CAP_NET_RAW\",\"CAP_SETGID\",\"CAP_SETUID\",\"CAP_SETFCAP\",\"CAP_SETPCAP\",\"CAP_NET_BIND_SERVICE\",\"CAP_SYS_CHROOT\",\"CAP_KILL\",\"CAP_AUDIT_WRITE\"],\"permitted\":[\"CAP_CHOWN\",\"CAP_DAC_OVERRIDE\",\"CAP_FSETID\",\"CAP_FOWNER\",\"CAP_MKNOD\",\"CAP_NET_RAW\",\"CAP_SETGID\",\"CAP_SETUID\",\"CAP_SETFCAP\",\"CAP_SETPCAP\",\"CAP_NET_BIND_SERVICE\",\"CAP_SYS_CHROOT\",\"CAP_KILL\",\"CAP_AUDIT_WRITE\"]}}}",
			"com.github.containers.virtcontainers.pkg.oci.bundle_path":	"/run/containerd/io.containerd.runtime.v1.linux/default/"+id,
			"com.github.containers.virtcontainers.pkg.oci.container_type":	"pod_sandbox",
		},
		Mounts: 	[]vc.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
				Options:     nil,
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
			},
			{
				Destination: "/dev/mqueue",
				Type:        "mqueue",
				Source:      "mqueue",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev", "ro"},
			},
			{
				Destination: "/run",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
		},
	}

	log.G(ctx).Infof("container config:  %v \n", container)

	// Sets the hypervisor configuration.
	hypervisorConfig := vc.HypervisorConfig{
		KernelParams:	[]vc.Param{
			vc.Param{
				Key:	"ip",
				Value:	"::::::"+id+"::off::",
			},
			// these params are used for rootfs image
			// vc.Param{
			// 	Key:	"init",
			// 	Value:	"/usr/lib/systemd/systemd",
			// },
			// vc.Param{
			// 	Key:	"systemd.unit",
			// 	Value:	"kata-containers.target",
			// },
			// vc.Param{
			// 	Key:	"systemd.mask",
			// 	Value:	"systemd-networkd.service",
			// },
			// vc.Param{
			// 	Key:	"systemd.mask",
			// 	Value:	"systemd-networkd.socket",
			// },
			vc.Param{
				Key:	"agent.log",
				Value:	"debug",
			},
			vc.Param{
				Key:	"qemu.cmdline",
				Value:	"-D <logfile>",
			},
		},
		KernelPath:     "/usr/share/kata-containers/vmlinuz.container",
		InitrdPath:     "/usr/share/kata-containers/kata-containers-initrd.img",
		HypervisorPath: "/usr/bin/qemu-lite-system-x86_64",

		BlockDeviceDriver:	"virtio-scsi",

		HypervisorMachineType:	"pc",

		DefaultVCPUs:	uint32(1),
		DefaultMaxVCPUs:	uint32(4),

		DefaultMemSz:	uint32(2048),

		DefaultBridges:	uint32(1),

		Mlock:	true,
		Msize9p:	uint32(8192),

		Debug:	true,
	}

	// Use KataAgent default values for the agent.
	agConfig := vc.KataAgentConfig{
		LongLiveConn:	true,
	}

	// VM resources
	vmConfig := vc.Resources{
		Memory: 2048,
	}

	// The sandbox configuration:
	// - One container
	// - Hypervisor is QEMU
	// - Agent is KataContainers
	sandboxConfig := vc.SandboxConfig{
		ID:	id,

		VMConfig: vmConfig,

		HypervisorType:   vc.QemuHypervisor,
		HypervisorConfig: hypervisorConfig,

		AgentType:   vc.KataContainersAgent,
		AgentConfig: agConfig,

		ProxyType:	vc.KataBuiltInProxyType,

		ShimType:	vc.KataBuiltInShimType,

		NetworkModel:	vc.CNMNetworkModel,
		NetworkConfig:	vc.NetworkConfig{
			NumInterfaces:		1,
			InterworkingModel:	2,
		},

		Containers: []vc.ContainerConfig{container},
	}
	log.G(ctx).Infoln("Sandbox: sandbox config: ", sandboxConfig)

	log.G(ctx).Infoln("Sandbox: create kata sandbox")

	sandbox, err := vc.CreateSandbox(sandboxConfig)
	log.G(ctx).Infoln("Sandbox: config！！！")
	if err != nil {
		log.G(ctx).Errorln("Sandbox: config error！！！", err)
		return errors.Wrapf(err, "Could not create sandbox")
	}
	log.G(ctx).Infof("Sandbox: create, VCSandbox is %v", sandbox)

	return err
}

// StartSandbox starts a kata-runtime sandbox
func StartSandbox(ctx context.Context, id string) error {
	log.G(ctx).Infoln("Sandbox: start kata sandbox")
	sandbox, err := vc.StartSandbox(id)
	if err != nil {
		return errors.Wrapf(err, "Could not start sandbox")
	}
	log.G(ctx).Infof("Sandbox: start, VCSandbox is %v", sandbox)

	return err
}

// StopSandbox stops a kata-runtime sandbox
func StopSandbox(ctx context.Context, id string, opts runtime.CreateOpts) error {
	return fmt.Errorf("stop not implemented")
}

// DeleteSandbox deletes a kata-runtime sandbox
func DeleteSandbox(ctx context.Context, id string, opts runtime.CreateOpts) error {
	return fmt.Errorf("delete not implemented")
}
