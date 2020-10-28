#!/bin/bash

# build virtiofs code
# This should be run inside qemu code
# Env requeriments
# QEMU_VIRTIOFS_TAG env var
# patches dir inside repository
# patches/<major-verison>.<minor-verions>.x
# patches/<QEMU_VIRTIOFS_TAG>

set -o errexit
set -o nounset
set -o pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

checks() {
	QEMU_VIRTIOFS_TAG=${QEMU_VIRTIOFS_TAG:-}
	PREFIX=${PREFIX:-}
	QEMU_VERSION=$(cat VERSION | awk 'BEGIN{FS=OFS="."}{print $1 "." $2 ".x"}')

	if [ "${QEMU_VIRTIOFS_TAG}" == "" ]; then
		echo "QEMU_VIRTIOFS_TAG is not set"
		exit 1
	fi

	qemu_virtiofs_patches_dir="./patches/${QEMU_VIRTIOFS_TAG}"
	if [ ! -d "${qemu_virtiofs_patches_dir}" ]; then
		echo "no virtiofs qemu patches dir: ${qemu_virtiofs_patches_dir}, needed even if there are not patches"
		exit 1
	fi

	qemu_patches_dir="./patches/${QEMU_VERSION}"
	if [ ! -d "${qemu_patches_dir}" ]; then
		echo "no qemu patches dir: ${qemu_patches_dir}, needed even if there are not patches"
		exit 1
	fi

	if [ "${PREFIX}" == "" ]; then
		echo "missing PREFIX env var"
		exit 1
	fi
}

patch_repo() {
	ls ./patches/
	ls ./patches/5.0.x/
	ls ./patches/qemu5.0-virtiofs-with51bits-dax/
	find "${qemu_virtiofs_patches_dir}" -name '*.patch' -type f | sort -t- -k1,1n >patches_virtiofs

	echo "Patches to apply for virtiofs tree:"
	cat "patches_virtiofs"
	[ ! -s "patches_virtiofs" ] || git apply $(cat "patches_virtiofs")

	find "${qemu_patches_dir}" -name '*.patch' -type f | sort -t- -k1,1n >"patches_qemu"
	echo "Patches to apply for qemu:"
	cat "patches_qemu"
	[ ! -s "patches_qemu" ] || git apply $(cat "patches_qemu")
}

build() {
	# Configure qemu build
	static_flag=""
	if [ "${STATIC_BUILD}" == "true" ]; then
		static_flag="-s"
	fi
	PREFIX="${PREFIX}" "${script_dir}/configure-hypervisor.sh" ${static_flag} kata-qemu-carlos | xargs ./configure \
		--with-pkgversion=kata-qemu-virtiofs

	# Build
	make -j$(nproc)

	# Install in dest dir
	make install DESTDIR=/tmp/qemu-virtiofs-static

	# Rename qemu binary to avoid collition wiht other builds
	mv /tmp/qemu-virtiofs-static/"${PREFIX}"/bin/qemu-system-x86_64 /tmp/qemu-virtiofs-static/"${PREFIX}"/bin/qemu-virtiofs-system-x86_64

	# Install virtiofsd
	chmod +x virtiofsd && mv virtiofsd /tmp/qemu-virtiofs-static/opt/kata/bin/

	# Create a tarball with binaries
	cd /tmp/qemu-virtiofs-static && tar -czvf "${QEMU_TARBALL}" *
}

checks
set -x
patch_repo
build
