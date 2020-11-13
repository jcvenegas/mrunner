#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

script_name=${0##*/}
script_dir=$(dirname "$(readlink -f "$0")")

test_prefix="fio-results-"


setup(){
	(
	cd "${script_dir}/../../workloads/fio/dockerfile/"
	docker build -f Dockerfile -t large-files-4gb .
	)
	docker rm -f large-files-4gb || true
}


drop_caches(){
	echo "info: drop_caches"
	free -h
	sync
	sudo bash -c 'echo 3 > /proc/sys/vm/drop_caches'
	sleep 3
	free -h
}

init_test_log(){
	test_log_file=${1}
	echo "Test log started" | tee "${test_log_file}"
}
info(){
	local msg=${1}
	echo "info: $1" | tee -a "${test_log_file}"
}


docker_rm(){
	local suffix=${1:-no-suffix}
	info "docker rm"
	docker rm -f large-files-4gb
	drop_caches | tee -a "${test_log_file}"
}

exec_fio(){
	log_suffix="${1:-no-suffix}"
	{ time docker exec -i large-files-4gb fio --gtod_reduce=1 --name=test --filename=random_read_write.fio --bs=4k --iodepth=64 --size=1G --readwrite=randrw --rwmixread=75; } 2>&1 | tee -a "${test_log_file}"
	info "drop caches after workload"
	info "caches will be high because VM still running"
	drop_caches | tee -a "${test_log_file}"
}

set_base_virtiofs_config(){
	# Running kata-qemu-virtiofs
	# Defaults for virtiofs
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_cache '"auto"'
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_cache_size 0
}

docker_run(){
	local runtime=${1}
	local suffix=${2}
	echo "case: kata-qemu-virtiofs ${suffix}"
	docker run -dti --runtime "${runtime}"  --name large-files-4gb large-files-4gb
	ps aux | grep virtiofsd > "virtiofsd-cmd-${runtime}-${suffix}"
	ps aux | grep qemu > "qemu-cmd-${runtime}-${suffix}"
}

fn_name(){
	  echo "${FUNCNAME[1]}"
}

kata_env(){
	local runtime=${1}
	local suffix=${2}
	/opt/kata/bin/${runtime} kata-env > "log-kata-${runtime}-env-${suffix}"
}

run_workload(){
	local runtime="${1}"
	local suffix="${2}"
	echo "case: ${runtime} ${suffix}"

	docker_run "${runtime}" "${suffix}"
	init_test_log "${test_prefix}${suffix}"
	drop_caches | tee "${test_log_file}"
	exec_fio  "${suffix}"
	docker_rm "${suffix}"
}

run_virtiofs_tread_pool_0(){
	local runtime="kata-qemu-virtiofs"
	local suffix="$(fn_name)"

	set_base_virtiofs_config
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_extra_args '["--thread-pool-size=0"]'

	kata_env "${runtime}" "${suffix}"
	run_workload "${runtime}" "${suffix}"
}

run_virtiofs_tread_pool_1(){
	local runtime="kata-qemu-virtiofs"
	local suffix="$(fn_name)"

	set_base_virtiofs_config
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_extra_args '["--thread-pool-size=1"]'

	kata_env "${runtime}" "${suffix}"
	run_workload "${runtime}" "${suffix}"

}

run_virtiofs_tread_pool_64(){
	local runtime="kata-qemu-virtiofs"
	local suffix="$(fn_name)"

	set_base_virtiofs_config
	sudo crudini --set --existing /opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml hypervisor.qemu virtio_fs_extra_args '["--thread-pool-size=64"]'

	kata_env "${runtime}" "${suffix}"
	run_workload "${runtime}" "${suffix}"
}


run_9pfs(){
	local runtime="kata-qemu"
	local suffix="$(fn_name)"

	kata_env "${runtime}" "${suffix}"
	run_workload "${runtime}" "${suffix}"
}
run_runc(){
	local runtime="runc"
	local suffix="$(fn_name)"

	run_workload "${runtime}" "${suffix}"
}

setup
run_runc
run_virtiofs_tread_pool_0
run_virtiofs_tread_pool_1
run_virtiofs_tread_pool_64
run_9pfs
