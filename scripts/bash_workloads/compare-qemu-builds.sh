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
	sudo rm -f "/opt/kata/bin/qemu-virtiofs-system-x86_64"
	make clean
	make
	sudo make install

	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu path '"/opt/kata/bin/qemu-virtiofs-system-x86_64"'

	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu path '"/opt/kata/bin/qemu-virtiofs-system-x86_64"'
	)
}




qemu_rh_dyn(){
	export QEMU_VIRTIOFS_REPO="https://gitlab.com/virtio-fs/qemu"
	export QEMU_VIRTIOFS_TAG="qemu5.0-virtiofs-with51bits-dax"
	export STATIC_BUILD="false"
	build_qemu

	${script_di}/fio_4g.sh | tee log-rh-dyn
}

qemu_rh_static(){
	export QEMU_VIRTIOFS_REPO="https://gitlab.com/virtio-fs/qemu"
	export QEMU_VIRTIOFS_TAG="qemu5.0-virtiofs-with51bits-dax"
	export STATIC_BUILD="true"
	build_qemu | tee qemu-build

	${script_dir}/fio_4g.sh | tee log-rh-static
}

default_qemu(){
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu path '"/opt/kata/bin/qemu-system-x86_64"'
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu.toml hypervisor.qemu path '"/opt/kata/bin/qemu-system-x86_64"'
	./fio_4g.sh | tee log-qemu-default
}


set -x
qemu_rh_static
default_qemu

bash ${script_dir}/get-results.sh
