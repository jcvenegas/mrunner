MK_DIR :=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

docker-build:
	"$(MK_DIR)/build-in-docker.sh"

install:
	tar xvf kata-qemu.tar.gz -C /

clean:
	rm -f kata-qemu.tar.gz
