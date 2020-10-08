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
QEMU_VIRTIOFS_TAG=${QEMU_VIRTIOFS_TAG:-}

if [ "${QEMU_VIRTIOFS_TAG}" == "" ]; then
	echo "QEMU_VIRTIOFS_TAG is not set"
	exit 1
fi

QEMU_VIRTIOFS_TAG=${QEMU_VIRTIOFS_TAG:-}

if [ "${QEMU_VIRTIOFS_TAG}" == "" ]; then
	echo "QEMU_VIRTIOFS_TAG is not set"
	exit 1
fi

[ -d qemu-virtiofs ] || git clone "${QEMU_VIRTIOFS_REPO}" qemu-virtiofs
cd qemu-virtiofs
git checkout "${QEMU_VIRTIOFS_TAG}"
