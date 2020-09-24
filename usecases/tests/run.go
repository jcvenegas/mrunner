package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mrunner/entities/kata"
	mtests "mrunner/entities/tests"
	"mrunner/utils"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	virtiofsCacheTomlKey     = "virtio_fs_cache"
	virtiofsCacheSizeTomlKey = "virtio_fs_cache_size"
	virtiofsdArgsTomlKey     = "virtio_fs_extra_args"
	kernelTomlKey            = "kernel"
)

func setupKataConfig(r kata.DockerRuntime, c kata.Config) error {
	fmt.Printf("Setup kata with config %#v\n", c)
	htype, err := r.HypervisorConfigKey()
	if err != nil {
		return err
	}

	err = r.SetConfigValue(htype, virtiofsCacheTomlKey, strconv.Quote(c.Hypervisor.CacheType))
	if err != nil {
		return err
	}

	r.SetConfigValue(htype, virtiofsCacheSizeTomlKey, strconv.Itoa(c.Hypervisor.CacheSize))
	if err != nil {
		return err
	}

	args := stringToTomlList(c.Hypervisor.VirtiofsdArgs)

	r.SetConfigValue(htype, virtiofsdArgsTomlKey, args)
	if err != nil {
		return err
	}
	r.SetConfigValue(htype, kernelTomlKey, strconv.Quote(c.Hypervisor.KernelPath))
	if err != nil {
		return err
	}

	return nil
}

func genKataHypervisorConfigCombinations(h HypervisorConfigs) ([]kata.HypervisorConfig, error) {
	hList := []kata.HypervisorConfig{}
	for _, c := range h.CacheTypes {
		for _, s := range h.CacheSizesBytes {
			for _, k := range h.KernelPaths {
				for _, a := range h.VirtiofsdArgs {
					hConfig := kata.HypervisorConfig{
						CacheType:     c,
						CacheSize:     s,
						VirtiofsdArgs: a,
						KernelPath:    k,
					}

					hList = append(hList, hConfig)
				}
			}
		}
	}
	return hList, nil

}

func testConfigIDArgs(runtime kata.DockerRuntime, k kata.Config) []string {
	idArgs := []string{}

	virtiofsArgsId := strings.Replace(k.Hypervisor.VirtiofsdArgs, " ", "", 0)
	virtiofsArgsId = strings.Replace(virtiofsArgsId, "-", "", 0)
	if virtiofsArgsId == "" {
		virtiofsArgsId = "no-args"
	}

	kernelName := "defaultKernel"
	if k.Hypervisor.KernelPath != "" {
		kernelName = path.Base(k.Hypervisor.KernelPath)

	}

	idArgs = append(idArgs, "runtime", string(runtime.RuntimeType))
	if runtime.RuntimeType == kata.KataQemu {
		idArgs = append(idArgs, "9pfs")

	} else {
		virtiofsIDArgs := []string{
			k.Hypervisor.CacheType,
			strconv.Itoa(k.Hypervisor.CacheSize),
			virtiofsArgsId,
		}
		idArgs = append(idArgs, "virtiofs", strings.Join(virtiofsIDArgs, "-"))

	}
	idArgs = append(idArgs, "kernel", kernelName)

	return idArgs

}

func saveRuntimeConfig(r kata.DockerRuntime) error {
	cpath, err := r.ConfigPath()
	if err != nil {
		return err
	}

	err = utils.CopyFile(cpath, "kata-configuration.toml", 0644)
	if err != nil {
		return err
	}

	envStr, err := r.KataEnv()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("kata-env.json", []byte(envStr), 0644)
	if err != nil {
		return err
	}
	return nil

}

func runTest(runtime kata.DockerRuntime, k kata.Config, t mtests.Test) (mtests.TestsResult, error) {
	result := mtests.TestsResult{}

	wd, err := os.Getwd()
	if err != err {
		return result, err
	}

	defer func() {
		os.Chdir(wd)
	}()

	testDirArgs := []string{
		wd,
		"results",
		t.ID(),
	}
	testDirArgs = append(testDirArgs, testConfigIDArgs(runtime, k)...)
	testDir := path.Join(testDirArgs...)

	err = os.MkdirAll(testDir, 0774)
	if err != err {
		return result, err
	}

	fmt.Println("[golang-sh]$ # Running workload in :", testDir)
	err = os.Chdir(testDir)
	if err != nil {
		return result, err
	}

	defer saveResult(&result)

	err = setupKataConfig(runtime, k)
	if err != nil {
		return result, err
	}

	err = saveRuntimeConfig(runtime)
	if err != nil {
		return result, err
	}

	err = t.Setup()
	if err != nil {
		return result, err
	}

	start := time.Now()
	result, err = t.Run(mtests.TestEnv{WorkDir: testDir, Runtime: string(runtime.RuntimeType)})
	elapsed := time.Since(start)
	result.Duration = elapsed
	if err != nil {
		return result, err
	}

	err = t.TearDown()
	if err != nil {
		return result, err
	}
	return result, nil
}

func saveResult(result *mtests.TestsResult) error {
	file, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("result.json", file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func runTestsForRuntimeConfig(runtime kata.DockerRuntime, t mtests.Test, h HypervisorConfigs) ([]mtests.TestsResult, error) {
	rList := []mtests.TestsResult{}
	hConfigs, err := genKataHypervisorConfigCombinations(h)
	if err != nil {
		return rList, err
	}
	for _, h := range hConfigs {
		kConfig := kata.Config{
			Hypervisor: h,
		}
		r, err := runTest(runtime, kConfig, t)
		if err != nil {
			return rList, err
		}
		rList = append(rList, r)
	}
	return rList, nil
}

func RunTestForKataConfigs(t mtests.Test, k Configs) ([]mtests.TestsResult, error) {
	rList := []mtests.TestsResult{}
	for _, r := range k.Runtimes {
		runtime, err := kata.NewDockerRuntime(r)
		if err != nil {
			return rList, err
		}
		res, err := runTestsForRuntimeConfig(runtime, t, k.HypervisorConfigs)
		if err != nil {
			return rList, err
		}
		rList = append(rList, res...)
	}
	return rList, nil
}
