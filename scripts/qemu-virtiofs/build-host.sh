#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export QEMU_VIRTIOFS_REPO="https://gitlab.com/virtio-fs/qemu.git"
# This tag will be supported on the runtime versions.yaml
export QEMU_VIRTIOFS_TAG="qemu5.0-virtiofs-with51bits-dax"
#
export PREFIX=/opt/kata
export STATIC_BUILD=false
export QEMU_TARBALL="kata-qemu.tar.gz"



sudo ${script_dir}/build_scripts/install_deps.sh
${script_dir}/build_scripts/clone_qemu.sh
cp ${script_dir}/patches qemu-virtiofs -r
cd qemu-virtiofs
${script_dir}/build_scripts/build.sh
