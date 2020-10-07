package tests

type HypervisorConfigs struct {
	CacheTypes      []string
	CacheSizesBytes []int
	VirtiofsdArgs   []string
	KernelPaths     []string
}

type RuntimeConfig struct {
	Runtime           string
	HypervisorConfigs HypervisorConfigs
}
