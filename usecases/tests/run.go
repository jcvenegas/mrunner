package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	runtime "mrunner/entities/runtime"
	mtests "mrunner/entities/tests"
	"mrunner/utils"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
)

const (
	virtiofsCacheTomlKey     = "virtio_fs_cache"
	virtiofsCacheSizeTomlKey = "virtio_fs_cache_size"
	virtiofsdArgsTomlKey     = "virtio_fs_extra_args"
	kernelTomlKey            = "kernel"
)

func setupKataConfig(r runtime.DockerRuntime, c runtime.Config) error {
	fmt.Printf("Setup runtime with config %#v\n", c)
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

func genKataHypervisorConfigCombinations(h HypervisorConfigs) ([]runtime.HypervisorConfig, error) {
	hList := []runtime.HypervisorConfig{}

	if len(h.CacheSizesBytes) == 0 {
		h.CacheSizesBytes = []int{0}
	}

	if len(h.CacheTypes) == 0 {
		h.CacheTypes = []string{""}
	}

	if len(h.KernelPaths) == 0 {
		h.KernelPaths = []string{""}
	}

	if len(h.VirtiofsdArgs) == 0 {
		h.VirtiofsdArgs = []string{""}
	}

	for _, c := range h.CacheTypes {
		for _, s := range h.CacheSizesBytes {
			for _, k := range h.KernelPaths {
				for _, a := range h.VirtiofsdArgs {
					hConfig := runtime.HypervisorConfig{
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

func testConfigIDArgs(r runtime.DockerRuntime, k runtime.Config) []string {

	idArgs := []string{}

	if r.RuntimeType == runtime.Runc {
		return append(idArgs, "runc")
	}

	virtiofsArgsId := strings.Replace(k.Hypervisor.VirtiofsdArgs, " ", "", 0)
	virtiofsArgsId = strings.Replace(virtiofsArgsId, "-", "", 0)
	if virtiofsArgsId == "" {
		virtiofsArgsId = "no-args"
	}

	kernelName := "defaultKernel"
	if k.Hypervisor.KernelPath != "" {
		kernelName = path.Base(k.Hypervisor.KernelPath)

	}

	idArgs = append(idArgs, "runtime", string(r.RuntimeType))
	if r.RuntimeType == runtime.KataQemu {
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

func saveKataRuntimeConfig(r runtime.DockerRuntime) error {
	cpath, err := r.ConfigPath()
	if err != nil {
		return err
	}

	err = utils.CopyFile(cpath, "runtime-configuration.toml", 0644)
	if err != nil {
		return err
	}

	envStr, err := r.KataEnv()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("runtime-env.json", []byte(envStr), 0644)
	if err != nil {
		return err
	}
	return nil

}

func saveHypervisorCmd(r runtime.DockerRuntime) error {
	s := sh.NewSession()
	s.ShowCMD = true
	hypervisorRegex := ""

	switch r.RuntimeType {
	case runtime.KataClh:
		hypervisorRegex = "[c]loud-hypervisor"
	case runtime.KataQemu, runtime.KataQemuVirtiofs:
		hypervisorRegex = "[q]emu-"

	}

	if hypervisorRegex == "" {
		return fmt.Errorf("No hypervisorRegex for %s", r.RuntimeType)
	}

	out, err := s.Command("ps", "aux").Command("grep", hypervisorRegex).Output()
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return fmt.Errorf("No hypervisor process found: %s", hypervisorRegex)
	}
	err = ioutil.WriteFile("hypervisor-cmd", out, 0644)
	if err != nil {
		return err
	}

	return nil

}

func saveVirtiofsdCmd(r runtime.DockerRuntime) error {
	s := sh.NewSession()
	s.ShowCMD = true
	virtiofsdRegex := "[v]irtiofsd"

	switch r.RuntimeType {
	case runtime.KataClh, runtime.KataQemuVirtiofs:
	default:
		fmt.Printf("runtime %s has not virtiofsd\n", r.RuntimeType)
		return nil
	}

	out, err := s.Command("ps", "aux").Command("grep", virtiofsdRegex).Output()
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return fmt.Errorf("No virtiofsd process found: %s", virtiofsdRegex)
	}

	err = ioutil.WriteFile("virtiofsd-cmd", out, 0644)
	if err != nil {
		return err
	}

	return nil

}

func runTest(r runtime.DockerRuntime, k runtime.Config, t mtests.Test) (mtests.TestsResult, error) {
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
	testDirArgs = append(testDirArgs, testConfigIDArgs(r, k)...)
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

	if r.RuntimeType != runtime.Runc {
		err = setupKataConfig(r, k)
		if err != nil {
			return result, err
		}

		err = saveKataRuntimeConfig(r)
		if err != nil {
			return result, err
		}

	}
	err = t.Setup()
	if err != nil {
		return result, err
	}

	start := time.Now()
	result, err = t.Run(mtests.TestEnv{WorkDir: testDir, Runtime: string(r.RuntimeType)})
	elapsed := time.Since(start)
	result.Duration = elapsed
	if err != nil {
		return result, err
	}

	if r.RuntimeType != runtime.Runc {
		err = saveHypervisorCmd(r)
		if err != nil {
			return result, err
		}
		err = saveVirtiofsdCmd(r)
		if err != nil {
			return result, err
		}

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

func runTestsForRuntimeConfig(r runtime.DockerRuntime, t mtests.Test, h HypervisorConfigs) ([]mtests.TestsResult, error) {
	rList := []mtests.TestsResult{}
	hConfigs, err := genKataHypervisorConfigCombinations(h)
	if err != nil {
		return rList, err
	}
	for _, h := range hConfigs {
		kConfig := runtime.Config{
			Hypervisor: h,
		}
		r, err := runTest(r, kConfig, t)
		if err != nil {
			return rList, err
		}
		rList = append(rList, r)
	}
	return rList, nil
}

func RunTestForKataConfigs(t mtests.Test, k []RuntimeConfig) ([]mtests.TestsResult, error) {
	rList := []mtests.TestsResult{}
	for _, r := range k {
		runtime, err := runtime.NewDockerRuntime(r.Runtime)
		if err != nil {
			return rList, err
		}
		res, err := runTestsForRuntimeConfig(runtime, t, r.HypervisorConfigs)
		if err != nil {
			return rList, err
		}
		rList = append(rList, res...)
	}
	return rList, nil
}
