#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

script_name=${0##*/}
script_dir=$(dirname "$(readlink -f "$0")")

build_qemu(){
	(
	cd "${script_dir}/../qemu-virtiofs"
	sudo rm -rf "/opt/virtiofs/"
	make clean
	make
	sudo make install

	#kata-virtiofs
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu path '"/opt/virtiofs/bin/qemu-virtiofs-system-x86_64"'
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_daemon '"/opt/virtiofs/bin/virtiofsd"'


	#kata-qemu
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu path '"/opt/virtiofs/bin/qemu-virtiofs-system-x86_64"'

	)
	cp /opt/virtiofs/share/applied_patches log-applied-patches
}

qemu_rh_dyn(){
	export STATIC_BUILD="false"
	build_qemu 2>&1| tee log-qemu-buld-dyn
	${script_dir}/fio_4g.sh | tee log-rh-dyn
}

qemu_rh_static(){
	export STATIC_BUILD="true"
	build_qemu 2>&1| tee log-qemu-build-static
	${script_dir}/fio_4g.sh | tee log-rh-static
}

default_qemu(){
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu path '"/opt/kata/bin/qemu-system-x86_64"'
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu path '"/opt/kata/bin/qemu-system-x86_64"'
	${script_dir}/fio_4g.sh | tee log-qemu-default
}

set -x

results_dir_name="${1:-results}"
mkdir -p "${results_dir_name}"
cd "${results_dir_name}"
qemu_rh_static
#default_qemu
bash ${script_dir}/get-results.sh
