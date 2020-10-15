#!/bin/bash

set -x
set -e
git clone https://github.com/jcvenegas/mrunner.git || true
cd mrunner/workloads/fio/dockerfile

docker rm -f large-files-4gb || true
# Running runc
docker build -f Dockerfile -t large-files-4gb .
docker run -dti --runtime runc --name large-files-4gb large-files-4gb
sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
docker exec -i large-files-4gb fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=4G --readwrite=randrw --rwmixread=75 
docker rm -f large-files-4gb

# Running kata-runtime
docker build -f Dockerfile -t large-files-4gb .
/usr/local/bin/kata-runtime kata-env
docker run -dti --runtime kata-runtime  --name large-files-4gb large-files-4gb
sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
docker exec -i large-files-4gb fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=4G --readwrite=randrw --rwmixread=75 
docker rm -f large-files-4gb

# Running kata-qemu-virtiofs
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_cache '"auto"'
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_cache_size 0
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_extra_args []
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu kernel '"/opt/kata/share/kata-containers/vmlinux.container"'
/opt/kata/bin/kata-qemu-virtiofs kata-env
docker build -f Dockerfile -t large-files-4gb .
docker run -dti --runtime kata-qemu-virtiofs  --name large-files-4gb large-files-4gb
sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
docker exec -i large-files-4gb fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=4G --readwrite=randrw --rwmixread=75 
docker rm -f large-files-4gb

# Running kata-qemu
docker build -f Dockerfile -t large-files-4gb .
/opt/kata/bin/kata-qemu kata-env
docker run -dti --runtime kata-qemu  --name large-files-4gb large-files-4gb
sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
docker exec -i large-files-4gb fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=4G --readwrite=randrw --rwmixread=75 
docker rm -f large-files-4gb
