package fioTest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"vfs-fio/entities/docker"
	"vfs-fio/entities/testing"

	"github.com/codeskyblue/go-sh"

	log "github.com/sirupsen/logrus"
)

const (
	virtiofsCacheTomlKey     = "virtio_fs_cache"
	virtiofsCacheSizeTomlKey = "virtio_fs_cache_size"
	virtiofsdArgsTomlKey     = "virtio_fs_extra_args"

	waitTimeSecBeforeDockerRm = 5

	timeOuMintFioTest = 10

	metricsReportDirName        = "results"
	metricsReportDirPermissions = 666

	fioLargeScriptPath = "./storage/fio-largefiles.sh"
)

type FioConfig struct {
	Name           string
	Command        string
	DockerFilePath string
}

type TestSetConfig struct {
	CacheTypes      []string
	CacheSizesBytes []int
	Runtimes        []string
	VirtiofsdArgs   []string
	KernelPaths     []string
	FioTestList     []FioConfig
}

type TestSetResult struct {
	Results []TestResult
}

type TestResult struct {
	TestConfig FioTestConfig
	Err        *TestError
	Duration   time.Duration
}

func (tr *TestResult) setError(err error) {
	if tr.Err == nil {
		tr.Err = &TestError{errors.New(err.Error())}
	}
}

type TestError struct {
	error
}

func (me TestError) MarshalJSON() ([]byte, error) {
	return json.Marshal(me.Error())
}

type FioTestConfig struct {
	Runtime       docker.DockerRuntime `json:"runtime"`
	CacheType     string               `json:"cacheType"`
	CacheSize     int                  `json:"CacheSize"`
	VirtiofsdArgs string               `json:"VirtiofsdArgs"`
	KernelPath    string               `json:"KernelPath"`
	Fio           FioConfig            `json:"Fio"`
}

func (tc *FioTestConfig) VirtiofsArgsToID() string {
	id := strings.Join(strings.Fields(tc.VirtiofsdArgs), "")
	if id == "" {
		id = "no-args"
	}
	return id
}

func (tc *FioTestConfig) ResultsID() string {

	kname := path.Base(tc.KernelPath)

	var dirNameAttrs []string
	dirNameAttrs = []string{
		"results",
		string(tc.Runtime.RuntimeType),
	}

	switch tc.Runtime.RuntimeType {
	case docker.KataQemu:
		dirNameAttrs = append(dirNameAttrs,
			"9pfs",
		)
	default:
		dirNameAttrs = append(dirNameAttrs,
			tc.CacheType,
			strconv.Itoa(tc.CacheSize),
			tc.VirtiofsArgsToID(),
			kname,
		)

	}
	return strings.Join(dirNameAttrs, "-")

}

func runFioConfig(c FioConfig, runtime string) error {
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true

	dockerFileDir := path.Dir(c.DockerFilePath)
	dockerFile := path.Base(c.DockerFilePath)

	err := s.Command("docker", "build", "-f", dockerFile, "-t", c.Name, dockerFileDir).Run()
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = s.Command("docker", "run", "-dti", "-v", wd+":/output", "--name", c.Name, c.Name).Run()
	if err != nil {
		return err
	}
	err = s.Command("docker", "exec", "-i", c.Name, "sh", "-c", c.Command).Run()
	if err != nil {
		return err
	}
	return nil
}

func renameResults(tc FioTestConfig) error {
	dirName := tc.ResultsID()

	if err := testing.BackupOldResultsDir(dirName); err != nil {
		return err
	}
	return os.Rename(metricsReportDirName, dirName)

}

//RunFioTest: Setup, run and collect results
// Returns error for fatal issues
// If failed with an managable error this is embedded in TestReuslt

func RunFioTest(tc FioTestConfig) (TestResult, error) {
	log.Infof("Running fio for %s cache=%s size=%d args=%q", tc.Runtime, tc.CacheType, tc.CacheSize, tc.VirtiofsdArgs)

	//Need to wait in case a container is already being removed
	time.Sleep(waitTimeSecBeforeDockerRm * time.Second)
	if err := docker.RmAll(); err != nil {
		return TestResult{}, err
	}

	if err := os.MkdirAll(metricsReportDirName, metricsReportDirPermissions); err != nil {
		return TestResult{}, err
	}

	tr := TestResult{TestConfig: tc}

	htype, err := tc.Runtime.HypervisorConfigKey()
	if err != nil {
		tr.setError(err)
	}

	err = tc.Runtime.SetConfigValue(htype, virtiofsCacheTomlKey, strconv.Quote(tc.CacheType))
	if err != nil {
	}

	tc.Runtime.SetConfigValue(htype, virtiofsCacheSizeTomlKey, strconv.Itoa(tc.CacheSize))
	if err != nil {
		tr.setError(err)
	}

	args := stringToTomlList(tc.VirtiofsdArgs)

	tc.Runtime.SetConfigValue(htype, virtiofsdArgsTomlKey, args)
	if err != nil {
		tr.setError(err)
	}

	start := time.Now()
	err = runFioConfig(tc.Fio, string(tc.Runtime.RuntimeType))
	if err != nil {
		tr.setError(err)
	}
	tr.Duration = time.Since(start)

	if err = saveResults(tr, metricsReportDirName); err != nil {
		tr.setError(err)
	}

	if err := renameResults(tc); err != nil {
		tr.setError(err)
	}

	return tr, nil
}

func stringToTomlList(str string) string {
	qStr := []string{}
	for _, a := range strings.Fields(str) {
		qStr = append(qStr, fmt.Sprintf("%q", a))
	}

	l := strings.Join(qStr, ", ")
	l = "[" + l + "]"
	return l
}

func saveResults(tr TestResult, metricsReportDirName string) error {
	file := path.Join(metricsReportDirName, "test_result.json")
	j, err := json.MarshalIndent(tr, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, j, metricsReportDirPermissions)

}

func runFioForRuntime(r docker.DockerRuntime, ts TestSetConfig) ([]TestResult, error) {
	rl := []TestResult{}

	if r.RuntimeType == docker.KataQemu {
		log.Infof("%s will run only one configuration", r.RuntimeType)
		tc := FioTestConfig{Runtime: r}
		tr, err := RunFioTest(tc)
		if err != nil {
			return rl, err
		}
		return []TestResult{tr}, err

	}

	// TODO find a way to  run all permutations with a lot of nested loops
	for _, c := range ts.CacheTypes {
		for _, s := range ts.CacheSizesBytes {
			for _, a := range ts.VirtiofsdArgs {
				for _, k := range ts.KernelPaths {
					for _, t := range ts.FioTestList {
						tc := FioTestConfig{
							Runtime:       r,
							CacheType:     c,
							CacheSize:     s,
							VirtiofsdArgs: a,
							KernelPath:    k,
							Fio:           t,
						}
						tr, err := RunFioTest(tc)
						if err != nil {
							return rl, err
						}
						rl = append(rl, tr)

					}
				}
			}
		}
	}
	return rl, nil
}

func RunFioTestSet(ts TestSetConfig) (TestSetResult, error) {
	result := TestSetResult{}
	for _, rt := range ts.Runtimes {
		r, err := docker.NewDockerRuntime(rt)
		if err != nil {
			return TestSetResult{}, err
		}
		rl, err := runFioForRuntime(r, ts)
		if err != nil {
			return result, err
		}
		result.Results = append(result.Results, rl...)
	}
	return result, nil
}
