# Mrunner: tool to run container workloads with different kata configs

## Install docker
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```

## Install other deps
- git to clone
- make to use makefile
- crudini is a dependency for the tool

## Install kata (using kata-deploy)

```bash
kata_version="1.12.0-alpha1"
docker run -v /opt/kata:/opt/kata -v /var/run/dbus:/var/run/dbus -v /run/systemd:/run/systemd -v /etc/docker:/etc/docker -it katadocker/kata-deploy:"${kata_version}" kata-deploy-docker install
docker info | grep Runtimes
```

```bash
git clone https://github.com/jcvenegas/mrunner.git
go build
# Use sudo if your user has not permissiones to use docker client.
./mrunner run ./workloads/fio/fio-file-4G.yaml
```
## Want to try with bash the tests?
Get the list of commands that are executed in the test.
```bash
./mrunner run ./workloads/fio/fio-file-4G.yaml 2>&1 | grep 'golang-sh' --color | awk '!($1="")'
```
Example:
```bash
./mrunner run ./workloads/fio/fio-file-4G.yaml 2>&1 | grep 'golang-sh' --color | awk '!($1="")'
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
- For now just modify the workload in based in ./workloads/fio/fio-file-4G.yaml

Example:

```

results/
└── large-files-4gb
    └── runtime
        ├── kata-clh
        │   └── virtiofs
        │       └── always-1024-no-args
        │           └── kernel
        │               └── vmlinux-kata-v5.6-april-09-2020-88-virtiofs
        │                   ├── fio.json
        │                   ├── kata-configuration.toml
        │                   ├── kata-env.json
        │                   └── result.json
        └── kata-qemu
            └── 9pfs
                └── kernel
                    └── vmlinux-kata-v5.6-april-09-2020-88-virtiofs
                        ├── fio.json
                        ├── kata-configuration.toml
                        ├── kata-env.json
                        └── result.json

11 directories, 8 files
```

## Data Collected

- `fio.json`: From fio tests results in json (depends on workload)
- `kata-configuration.toml` kata config used to run kata
- `kata-env.json` output from  `kata-runtime kata-env` (using the correct kata-deploy binary)
- `result.json` a file with status of the tests, if passed or failed, if failed store the error it got, duration.

2 directories, 8 files

