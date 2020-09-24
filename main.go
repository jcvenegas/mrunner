package main

import (
	"fmt"
	"io/ioutil"
	"os"
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

	if len(os.Args) < 2 {
		fmt.Println("Missing yaml file")
		os.Exit(1)
	}
	yamlFile := os.Args[1]
	dat, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	testFile := TestFile{}

	err = yaml.Unmarshal(dat, &testFile)
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
