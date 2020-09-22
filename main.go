package main

import (
	"fmt"
	"os"
	"path"
	"time"

	// TODO: Should not import entities
	dtests "mrunner/entities/tests/docker"
	"mrunner/usecases/tests"
)

func main() {

	kc := tests.Configs{
		Runtimes: []string{"kata-clh", "kata-qemu"},
		HypervisorConfigs: tests.HypervisorConfigs{
			CacheTypes:      []string{"always"},
			CacheSizesBytes: []int{1024},
			VirtiofsdArgs:   []string{""},
			KernelPaths: []string{
				"/opt/kata/share/kata-containers/vmlinux-kata-v5.6-april-09-2020-88-virtiofs",
			},
		},
	}

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
	}
	dockerFilePath := path.Join(wd, "/workloads/fio/dockerfile/Dockerfile")

	t := dtests.DockerTest{
		Name:           "large-files-4gb",
		Command:        "",
		Exec:           fioCmd,
		DockerFilePath: dockerFilePath,
		Timeout:        20 * time.Minute,
	}

	v := dtests.TestWorkDirVolume{ContainerShare: "/output"}
	t.AddVolume(v)

	fmt.Printf("Running: %#v\n", kc)

	rs, err := tests.RunTestForKataConfigs(&t, kc)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	for _, r := range rs {
		fmt.Printf("%v err=%v time=%v\n", r.TestID, r.Error, r.Duration)
	}
}
