#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/qemu_version.sh"
QEMU_VIRTIOFS_REPO="$(get_qemu_repo)"
QEMU_VIRTIOFS_TAG="$(get_qemu_version)"
#
export PREFIX=/opt/kata
# use dynamic build as will be installed on host
export STATIC_BUILD=false
export QEMU_TARBALL="kata-qemu.tar.gz"



sudo ${script_dir}/build_scripts/install_deps.sh
${script_dir}/build_scripts/clone_qemu.sh
cp ${script_dir}/patches qemu-virtiofs -r
cd qemu-virtiofs
${script_dir}/build_scripts/build.sh
