package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mrunner/entities/kata"
	mtests "mrunner/entities/tests"
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

func testID(runtime kata.DockerRuntime, k kata.Config) string {

	virtiofsArgsId := strings.Replace(k.Hypervisor.VirtiofsdArgs, " ", "", 0)
	virtiofsArgsId = strings.Replace(virtiofsArgsId, "-", "", 0)
	if virtiofsArgsId == "" {
		virtiofsArgsId = "no-args"
	}

	kernelName := "defaultKernel"
	if k.Hypervisor.KernelPath != "" {
		kernelName = path.Base(k.Hypervisor.KernelPath)

	}

	idArgs := []string{
		string(runtime.RuntimeType),
		k.Hypervisor.CacheType,
		strconv.Itoa(k.Hypervisor.CacheSize),
		virtiofsArgsId,
		kernelName,
	}

	return strings.Join(idArgs, "-")

}

func runTest(runtime kata.DockerRuntime, k kata.Config, t mtests.Test) (mtests.TestsResult, error) {
	result := mtests.TestsResult{}

	wd, err := os.Getwd()
	if err != err {
		return result, err
	}

	tID := testID(runtime, k)
	testDir := path.Join(wd, tID)

	err = os.MkdirAll(testDir, 666)
	if err != err {
		return result, err
	}

	err = setupKataConfig(runtime, k)
	if err != nil {
		return result, err
	}

	defer func() {
		os.Chdir(wd)
	}()

	err = t.Setup()
	if err != nil {
		return result, err
	}

	fmt.Println("Cd to ", testDir)
	err = os.Chdir(testDir)
	if err != nil {
		return result, err
	}

	start := time.Now()
	result, err = t.Run(mtests.TestEnv{WorkDir: testDir})
	elapsed := time.Since(start)
	result.Duration = elapsed
	if err != nil {
		return result, err
	}

	file, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return result, err
	}

	err = ioutil.WriteFile("result.json", file, 0644)
	if err != nil {
		return result, err
	}

	err = t.TearDown()
	if err != nil {
		return result, err
	}
	return result, nil
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
		res, err := runTestsForRuntimeConfig(runtime, t, k.Hypervisor)
		if err != nil {
			return rList, err
		}
		rList = append(rList, res...)
	}
	return rList, nil
}
