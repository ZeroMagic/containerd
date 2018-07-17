How to use containerd-kata

这里我已经将container-kata runtime plugin加入runtime文件夹中。



1、编译

make && make install



2、启动containerd

由于我目前不知道怎么用crictl直接指定用哪个runtime plugin。所以，我对contianerd的配置文件进行了修改。

文件里的sandbox_image根据本地的进行修改，虽然我们用不到。

另外改动的地方就是，将默认的runtime_type修改为io.containerd.runtime.v1.kata-runtime

/etc/container/config.toml

    root = "/var/lib/containerd"
    state = "/run/containerd"
    oom_score = 0
    
    [grpc]
      address = "/run/containerd/containerd.sock"
      uid = 0
      gid = 0
      max_recv_message_size = 16777216
      max_send_message_size = 16777216
    
    [debug]
      address = ""
      uid = 0
      gid = 0
      level = ""
    
    [metrics]
      address = ""
      grpc_histogram = false
    
    [cgroup]
      path = "/sys/fs/cgroup"
    
    [plugins]
      [plugins.cgroups]
        no_prometheus = false
      [plugins.cri]
        stream_server_address = ""
        stream_server_port = "10010"
        enable_selinux = false
        sandbox_image = "k8s.gcr.io/pause-amd64/pause-amd64:3.0"
        stats_collect_period = 10    
        systemd_cgroup = false
        enable_tls_streaming = false
        [plugins.cri.containerd]
          snapshotter = "overlayfs"
          [plugins.cri.containerd.default_runtime]
            runtime_type = "io.containerd.runtime.v1.kata-runtime"
            runtime_engine = ""
            runtime_root = ""
          [plugins.cri.containerd.untrusted_workload_runtime]
            runtime_type = "io.containerd.runtime.v1.kata-runtime"
            runtime_engine = ""
            runtime_root = ""
        [plugins.cri.cni]
          bin_dir = "/opt/cni/bin"
          conf_dir = "/etc/cni/net.d"
        [plugins.cri.registry]
          [plugins.cri.registry.mirrors]
            [plugins.cri.registry.mirrors."docker.io"]
              endpoint = ["https://registry.docker-cn.com"]
      [plugins.diff-service]
        default = ["walking"]
      [plugins.linux]
        shim = "containerd-shim"
        runtime = "runc"
        runtime_root = ""
        no_shim = false
        shim_debug = false
      [plugins.scheduler]
        pause_threshold = 0.02
        deletion_threshold = 0
        mutation_threshold = 100
        schedule_delay = "0s"
        startup_delay = "100ms"
    

3、安装crictl

这里不要用master上最新的，因为我用的时候有问题，降了版本后就好了。

    git clone https://github.com/kubernetes-incubator/cri-tools.git -b v1.11.0
    
    make && make install

4、crictl的使用

简要说明：https://github.com/containerd/cri/blob/master/docs/crictl.md

    1、先确定image已经存在。若不存在直接用crictl pull就行
    $ sudo crictl images
    IMAGE                       TAG                 IMAGE ID            SIZE
    docker.io/library/busybox   latest              f6e427c148a76       728kB
    k8s.gcr.io/pause-amd64      3.0                 da86e6ba6ca19       746kB
    
    2、配置sandbox的config
    $ cat sandbox-config.json
    {
        "metadata": {
            "name": "busybox-sandbox",
            "namespace": "default",
            "attempt": 1,
            "uid": "hdishd83djaidwnduwk28bcsb"
        },
        "linux": {
    }
    
    3、配置container的config
    $ cat container-config.json
    {
      "metadata": {
          "name": "busybox"
      },
      "image":{
          "image": "busybox"
      },
      "command": [
          "ls"
      ],
      "linux": {
      }
    }
    
    4、运行sandbox
    $ crictl runp sandbox-config.json
    e1c83b0b8d481d4af8ba98d5f7812577fc175a37b10dc824335951f52addbb4e
    
    5、创建container
    $ crictl create e1c83 container-config.json sandbox-config.json
    0a2c761303163f2acaaeaee07d2ba143ee4cea7e3bde3d32190e2a36525c8a05
    
    6、运行container
    $ crictl start 0a2c
    
    









 

 

 
