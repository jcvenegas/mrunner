package runtime

import (
	"fmt"
	"path"

	"github.com/codeskyblue/go-sh"
)

// Tests set config
// Need to move out of here
type HypervisorConfigs struct {
	CacheTypes      []string
	CacheSizesBytes []int
	VirtiofsdArgs   []string
	KernelPaths     []string
}

type Configs struct {
	Runtimes   []string
	Hypervisor HypervisorConfigs
}

// Kata
type HypervisorConfig struct {
	CacheType     string
	CacheSize     int
	VirtiofsdArgs string
	KernelPath    string
}

type Config struct {
	Hypervisor HypervisorConfig
}

// Docker Runtime
type DockerRuntime struct {
	RuntimeType DockerRuntimeType
}

type DockerRuntimeType string

const (
	KataClh          DockerRuntimeType = "kata-clh"
	KataQemu         DockerRuntimeType = "kata-qemu"
	KataQemuVirtiofs DockerRuntimeType = "kata-qemu-virtiofs"
	Runc             DockerRuntimeType = "runc"
)

func NewDockerRuntime(runtime string) (DockerRuntime, error) {
	RuntimeType := DockerRuntimeType(runtime)
	switch RuntimeType {
	case KataClh, KataQemu, KataQemuVirtiofs, Runc:
		return DockerRuntime{RuntimeType: RuntimeType}, nil
	}
	return DockerRuntime{}, fmt.Errorf("unknown config: %s", runtime)
}

func (dr *DockerRuntime) ConfigPath() (string, error) {
	switch dr.RuntimeType {
	case KataClh:
		return "/opt/kata/share/defaults/kata-containers/configuration-clh.toml", nil
	case KataQemu:
		return "/opt/kata/share/defaults/kata-containers/configuration-qemu.toml", nil
	case KataQemuVirtiofs:
		return "/opt/kata/share/defaults/kata-containers/configuration-qemu-virtiofs.toml", nil
	case Runc:
		return "", fmt.Errorf("no config file for runc")
	default:
		return "", fmt.Errorf("failed to find config path for rdr.RuntimeType: %s", dr.RuntimeType)
	}
}

// Given a config file path set a value
// [ section ]
// attr = value
func (dr *DockerRuntime) SetConfigValue(section, attr, value string) error {
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true

	configPath, err := dr.ConfigPath()
	if err != nil {
		return err
	}
	cmd := []string{
		"crudini",
		"--set",
		"--existing",
		configPath,
		section,
		attr,
		value,
	}

	return s.Command("sudo", cmd).Run()
}

// This header changes depending the hypervisor
func (dr *DockerRuntime) HypervisorConfigKey() (string, error) {
	hypervisorKey := "hypervisor."
	switch dr.RuntimeType {
	case KataClh:
		return hypervisorKey + "clh", nil
	case KataQemu, KataQemuVirtiofs:
		return hypervisorKey + "qemu", nil
	case Runc:
		return "", fmt.Errorf("runc does not have hypervisor")
	default:
		return "", fmt.Errorf("failed to find HypervisorConfigKey for %s", dr.RuntimeType)
	}
}

func (dr *DockerRuntime) RuntimePath() (string, error) {
	prefixPath := "/opt/kata/bin/"
	var runtimeBinName string

	switch dr.RuntimeType {
	case KataClh:
		runtimeBinName = "kata-clh"
	case KataQemu:
		runtimeBinName = "kata-qemu"
	case KataQemuVirtiofs:
		runtimeBinName = "kata-qemu-virtiofs"
	case Runc:
	default:
		return "", fmt.Errorf("failed to find HypervisorConfigKey for %s", dr.RuntimeType)
	}

	return path.Join(prefixPath, runtimeBinName), nil
}

func (dr *DockerRuntime) KataEnv() (string, error) {
	kPath, err := dr.RuntimePath()
	if err != nil {
		return "", err
	}
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true
	outByte, err := s.Command(kPath, "kata-env").Output()
	if err != nil {
		return "", err
	}

	return string(outByte), nil
}
