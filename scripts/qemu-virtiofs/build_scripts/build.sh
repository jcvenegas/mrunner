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
STATIC_BUILD=${STATIC_BUILD:-true}
BUILD_NAME="qemu-for-kata"

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
	echo "" > applied_patches
	qemu_version="$(cat VERSION)"
	stable_branch=$(echo $qemu_version | \
		awk 'BEGIN{FS=OFS="."}{print $1 "." $2 ".x"}')
	patches_dir_base="./patches"

	patches_dir_stable="${patches_dir_base}/${stable_branch}"
	echo "Handle patches for QEMU $qemu_version (stable ${stable_branch}) in ${patches_dir_stable}"
	if [ -d $patches_dir_stable ]; then
		for patch in $(find $patches_dir_stable -name '*.patch'); do
			echo "Apply $patch"
			echo "${patch}" >> applied_patches
			git apply "$patch"
		done
	else
		echo "No patches to apply"
	fi

	patches_dir_tag="${patches_dir_base}/${QEMU_VIRTIOFS_TAG}"
	echo "Handle patches for QEMU tag $QEMU_VIRTIOFS_TAG in ${qemu_virtiofs_patches_dir}"
	if [ -d "$patches_dir_tag" ] && [ "${QEMU_VIRTIOFS_TAG}" != "" ] ; then
		for patch in $(find $patches_dir_tag -name '*.patch'); do
			echo "Apply $patch"
			echo "${patch}" >> applied_patches
			git apply "$patch"
		done
	else
		echo "No patches to apply"
	fi
}

build() {
	# Configure qemu build
	static_flag=""
	if [ "${STATIC_BUILD}" == "true" ]; then
		static_flag="-s"
	fi
	export CFLAGS+="-Wno-error"
	PREFIX="${PREFIX}" "${script_dir}/configure-hypervisor.sh" ${static_flag} "${BUILD_NAME}" | xargs ./configure \
		--with-pkgversion=kata-qemu-virtiofs

	# Build
	make -j$(nproc)

	# Install in dest dir
	destdir=/tmp/qemu-virtiofs-static
	make install DESTDIR="${destdir}"

	# Rename qemu binary to avoid collition wiht other builds
	mv "${destdir}/${PREFIX}"/bin/qemu-system-x86_64 "${destdir}/${PREFIX}"/bin/qemu-virtiofs-system-x86_64

	# Install virtiofsd
	cp "${destdir}/${PREFIX}/libexec/${BUILD_NAME}/virtiofsd" "${destdir}/${PREFIX}/bin"

	# Save applied patches
	cp applied_patches "${destdir}/${PREFIX}/share"
	# Create a tarball with binaries
	cd "${destdir}" && tar -czvf "${QEMU_TARBALL}" *
}

checks
patch_repo
build
