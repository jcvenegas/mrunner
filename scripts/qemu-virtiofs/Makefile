MK_DIR :=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

docker-build:
	"$(MK_DIR)/build-in-docker.sh"

install:
	tar xf kata-qemu.tar.gz -C /

clean:
	rm -f kata-qemu-static.tar.gz
