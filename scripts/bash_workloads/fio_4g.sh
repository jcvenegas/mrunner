#!/bin/bash

set -x
set -e

script_name=${0##*/}
script_dir=$(dirname "$(readlink -f "$0")")

cd "${script_dir}/../../workloads/fio/dockerfile/"

docker rm -f large-files-4gb || true
# Running kata-qemu-virtiofs
# Defaults for virtiofs
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_cache '"auto"'
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_cache_size 0
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_extra_args []
sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu kernel '"/opt/kata/share/kata-containers/vmlinux.container"'
/opt/kata/bin/kata-qemu-virtiofs kata-env
echo "case: kata-qemu-virtiofs"
docker build -f Dockerfile -t large-files-4gb .
docker run -dti --runtime kata-qemu-virtiofs  --name large-files-4gb large-files-4gb
sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
docker exec -i large-files-4gb fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=300M --readwrite=randrw --rwmixread=75
docker rm -f large-files-4gb

# Running kata-qemu
echo "case: kata-qemu-9pfs"
docker build -f Dockerfile -t large-files-4gb .
/opt/kata/bin/kata-qemu kata-env
docker run -dti --runtime kata-qemu  --name large-files-4gb large-files-4gb
sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
docker exec -i large-files-4gb fio --direct=1 --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=300M --readwrite=randrw --rwmixread=75
docker rm -f large-files-4gb
