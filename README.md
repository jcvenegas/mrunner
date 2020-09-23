# Mrunner: tool to run container workloads with different kata configs

```bash
git clone https://github.com/jcvenegas/mrunner.git
go build
# Use sudo if your user has not permissiones to use docker client.
sudo -E PATH=$PATH ./mrunner
```
## Want to try with bash the tests?
Get the list of commands that are executed in the test.
```bash
./mrunner 2>&1 | grep 'golang-sh' --color | awk '!($1="")'
```
Example:
```bash
./mrunner 2>&1 | grep 'golang-sh' --color | awk '!($1="")'
 # Running workload in : /home/jcvenega/go/src/github.com/kata-containers/tests/metrics/mrunner/results/large-files-4gb/kata-clh-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-clh.toml hypervisor.clh virtio_fs_cache "always"
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-clh.toml hypervisor.clh virtio_fs_cache_size 1024
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-clh.toml hypervisor.clh virtio_fs_extra_args []
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-clh.toml hypervisor.clh kernel "/opt/kata/share/kata-containers/vmlinux-kata-v5.6-april-09-2020-88-virtiofs"
 /opt/kata/bin/kata-clh kata-env
 docker build -f Dockerfile -t large-files-4gb .
 docker run -dti -v /home/jcvenega/go/src/github.com/kata-containers/tests/metrics/mrunner/results/large-files-4gb/kata-clh-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs:/output --name large-files-4gb large-files-4gb
 docker exec -i large-files-4gb sh -c fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=10M --readwrite=randrw --rwmixread=75 --output-format=json --output=/output/fio.json
 docker rm -f large-files-4gb
 # Running workload in : /home/jcvenega/go/src/github.com/kata-containers/tests/metrics/mrunner/results/large-files-4gb/kata-qemu-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu virtio_fs_cache "always"
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu virtio_fs_cache_size 1024
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu virtio_fs_extra_args []
 sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu kernel "/opt/kata/share/kata-containers/vmlinux-kata-v5.6-april-09-2020-88-virtiofs"
 /opt/kata/bin/kata-qemu kata-env
 docker build -f Dockerfile -t large-files-4gb .
 docker run -dti -v /home/jcvenega/go/src/github.com/kata-containers/tests/metrics/mrunner/results/large-files-4gb/kata-qemu-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs:/output --name large-files-4gb large-files-4gb
 docker exec -i large-files-4gb sh -c fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=10M --readwrite=randrw --rwmixread=75 --output-format=json --output=/output/fio.json
 docker rm -f large-files-4gb
```
## What it can do?
- Run a test on top of docker
- Run the same test for multiple kata configs
  - kerne path
  - virtiofsd cache type
  - virtiofsd dax size
  - kata-runtime (based in paths from kata-deploy)
  - Extra virtiofs args

It will generate directory named: `results`

Workloads:
- For now just modify the workload in `main.go`

Example:

```
results/
├── kata-clh-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs
│   ├── fio.json
│   ├── kata-configuration.toml
│   ├── kata-env.json
│   └── result.json
└── kata-qemu-always-1024-no-args-vmlinux-kata-v5.6-april-09-2020-88-virtiofs
    ├── fio.json
    ├── kata-configuration.toml
    ├── kata-env.json
    └── result.json
```

2 directories, 8 files


TODO:
- Create workloads definitions in yaml
