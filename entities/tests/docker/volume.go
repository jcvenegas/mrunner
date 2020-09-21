package docker

import "os"

type ContainerVolume interface {
	Host() (string, error)
	Container() (string, error)
}

type HostContainerMountVolume struct {
	HostShare      string
	ContainerShare string
}

func (v HostContainerMountVolume) Host() (string, error) {
	return v.HostShare, nil
}

func (v HostContainerMountVolume) Container() (string, error) {
	return v.ContainerShare, nil
}

type TestWorkDirVolume struct {
	ContainerShare string
}

func (v TestWorkDirVolume) Host() (string, error) {
	return os.Getwd()
}

func (v TestWorkDirVolume) Container() (string, error) {
	return v.ContainerShare, nil
}
