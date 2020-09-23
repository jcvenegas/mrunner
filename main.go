package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	// TODO: Should not import entities
	dtests "mrunner/entities/tests/docker"
	"mrunner/usecases/tests"

	"gopkg.in/yaml.v2"
)

type ContainerWorkload struct {
	Name           string
	Command        string
	Exec           string
	DockerFilePath string
	TimeoutMinutes int
}

type TestFile struct {
	ContainerEngine   string
	RuntimeConfigs    tests.Configs
	ContainerWorkload ContainerWorkload
}

func main() {

	fioCmd := "fio"
	fioCmd += " --direct=1"
	fioCmd += " --gtod_reduce=1"
	fioCmd += " --name=test"
	fioCmd += " --filename=random_read_write.fio"
	fioCmd += " --bs=4k"
	fioCmd += " --iodepth=64"
	fioCmd += " --size=10M"
	fioCmd += " --readwrite=randrw"
	fioCmd += " --rwmixread=75"
	fioCmd += " --output-format=json"
	fioCmd += " --output=/output/fio.json"

	wd, err := os.Getwd()

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	dockerFilePath := path.Join(wd, "/workloads/fio/dockerfile/Dockerfile")

	testFile := TestFile{
		ContainerEngine: "docker",
		RuntimeConfigs: tests.Configs{
			Runtimes: []string{"kata-clh", "kata-qemu"},
			HypervisorConfigs: tests.HypervisorConfigs{
				CacheTypes:      []string{"always"},
				CacheSizesBytes: []int{1024},
				VirtiofsdArgs:   []string{""},
				KernelPaths: []string{
					"/opt/kata/share/kata-containers/vmlinux-kata-v5.6-april-09-2020-88-virtiofs",
				},
			},
		},
		ContainerWorkload: ContainerWorkload{
			Name:           "large-files-4gb",
			Command:        "",
			Exec:           fioCmd,
			DockerFilePath: dockerFilePath,
			TimeoutMinutes: 10,
		},
	}
	d, err := yaml.Marshal(&testFile)
	err = ioutil.WriteFile("workloads.yaml", d, 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	t := dtests.DockerTest{
		Name:           testFile.ContainerWorkload.Name,
		Command:        testFile.ContainerWorkload.Command,
		Exec:           testFile.ContainerWorkload.Exec,
		DockerFilePath: testFile.ContainerWorkload.DockerFilePath,
		Timeout:        time.Duration(testFile.ContainerWorkload.TimeoutMinutes) * time.Minute,
	}

	v := dtests.TestWorkDirVolume{ContainerShare: "/output"}
	t.AddVolume(v)

	rs, err := tests.RunTestForKataConfigs(&t, testFile.RuntimeConfigs)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	for _, r := range rs {
		fmt.Printf("%v err=%v time=%v\n", r.TestID, r.Error, r.Duration)
	}
}
