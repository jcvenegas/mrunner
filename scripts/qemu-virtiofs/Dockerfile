ARG IMAGE=ubuntu
ARG TAG=20.04
FROM ${IMAGE}:${TAG}

# Args
#-----------------------------
# Repository to pull for qemu
ARG QEMU_VIRTIOFS_REPO
# commit/tag/branch
ARG QEMU_VIRTIOFS_TAG
# Name of the binary tarball
ARG QEMU_TARBALL=qemu.targ.gz
# Prefix where to install qemu
ARG PREFIX=/opt/virtiofs
# Build static
ARG STATIC_BUILD=true

WORKDIR /root/

# Dir to add build scripts
RUN mkdir /root/build_scripts

ENV http_proxy http://proxy-chain.intel.com:911
ENV https_proxy http://proxy-chain.intel.com:912
ENV ftp_proxy http://proxy-chain.intel.com:911
ENV socks_proxy http://proxy-us.intel.com:1080
ENV no_proxy intel.com,.intel.com,localhost,127.0.0.1,10.0.0.0/8,192.168.0.0/16,172.16.0.0/12
# Install deps
ADD build_scripts/install_deps.sh /root/build_scripts
RUN ./build_scripts/install_deps.sh

# Get qemu-virtiofs code
ADD build_scripts/clone_qemu.sh /root/build_scripts
RUN bash -x ./build_scripts/clone_qemu.sh
WORKDIR /root/qemu-virtiofs

# Add patches
ADD patches /root/qemu-virtiofs/patches

# Build
ADD build_scripts/build.sh /root/build_scripts
ADD build_scripts/configure-hypervisor.sh /root/build_scripts
RUN bash -x /root/build_scripts/build.sh
