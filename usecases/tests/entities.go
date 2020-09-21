package tests

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
