#!/bin/bash
#
# Copyright (c) 2019 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail


script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DOCKER_CLI="docker"

set -x

# Qemu repository
source "${script_dir}/qemu_version.sh"
qemu_virtiofs_repo="$(get_qemu_repo)"
# This tag will be supported on the runtime versions.yaml
qemu_virtiofs_tag="$(get_qemu_version)"
# Name for binary tarball
qemu_virtiofs_tar="kata-qemu.tar.gz"
# Ask to build static
static_build="${STATIC_BUILD:-false}"

echo "Build ${qemu_virtiofs_repo} tag: ${qemu_virtiofs_tag}"

prefix="${prefix:-"/opt/virtiofs"}"

if [ "${static_build}" == "true" ];then
	docker_base_image="ubuntu"
	docker_base_tag="20.04"
else
	docker_base_image=$(source /etc/os-release; echo "${ID}")
	docker_base_tag=$(source /etc/os-release; echo "${VERSION_ID}")
fi

sudo "${DOCKER_CLI}" build \
	--build-arg IMAGE="${docker_base_image}" \
	--build-arg TAG="${docker_base_tag}" \
	--build-arg QEMU_VIRTIOFS_REPO="${qemu_virtiofs_repo}" \
	--build-arg QEMU_VIRTIOFS_TAG="${qemu_virtiofs_tag}" \
	--build-arg QEMU_TARBALL="${qemu_virtiofs_tar}" \
	--build-arg PREFIX="${prefix}" \
	--build-arg STATIC_BUILD="${static_build}" \
	-f "${script_dir}/Dockerfile" \
	-t qemu-virtiofs-build .

sudo "${DOCKER_CLI}" run --rm \
	-i \
	-v "${PWD}":/share qemu-virtiofs-build \
	mv "/tmp/qemu-virtiofs-static/${qemu_virtiofs_tar}" /share/

sudo chown "${USER}:${USER}" "${PWD}/${qemu_virtiofs_tar}"
