package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	// TODO: Should not import entities
	dtests "mrunner/entities/tests/docker"
	"mrunner/usecases/tests"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type containerWorkload struct {
	Name           string
	Command        string
	PreExec        []string
	Exec           []string
	DockerFilePath string
	TimeoutMinutes int
}

type testFile struct {
	ContainerEngine   string
	RuntimeConfigs    []tests.RuntimeConfig
	ContainerWorkload containerWorkload
}

func runWorkload(yamlFile string) error {
	dat, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read yaml file %q %w", yamlFile, err)
	}

	yamlFileDir := path.Dir(yamlFile)
	yamlFileDir, err = filepath.Abs(yamlFileDir)
	if err != nil {
		return fmt.Errorf("failed to find abs dir for yaml file %w", err)
	}

	tf := testFile{}

	err = yaml.Unmarshal(dat, &tf)
	if err != nil {
		return err
	}

	dfilePath := tf.ContainerWorkload.DockerFilePath
	if !path.IsAbs(dfilePath) {
		dfilePath = path.Join(yamlFileDir, dfilePath)
	}

	if len(tf.ContainerWorkload.Exec) == 0 {
		return errors.New("Exec list for workload is empty")
	}

	t := dtests.Test{
		Name:           tf.ContainerWorkload.Name,
		Command:        tf.ContainerWorkload.Command,
		Exec:           tf.ContainerWorkload.Exec,
		PreExec:        tf.ContainerWorkload.PreExec,
		DockerFilePath: dfilePath,
		Timeout:        time.Duration(tf.ContainerWorkload.TimeoutMinutes) * time.Minute,
	}

	v := dtests.TestWorkDirVolume{ContainerShare: "/output"}
	t.AddVolume(v)

	rs, err := tests.RunTestForKataConfigs(&t, tf.RuntimeConfigs)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	for _, r := range rs {
		fmt.Printf("%v err=%v time=%v\n", r.TestID, r.Error, r.Duration)
	}

	return nil
}

func createTemplate() error {
	tf := testFile{}

	tf.ContainerEngine = "docker"
	dockerfileName := "Dockerfile"

	tf.ContainerWorkload = containerWorkload{
		Name:    "example",
		Command: "sleep infinity",
		PreExec: []string{
			"echo 3 > /proc/sys/vm/drop_caches",
		},
		Exec: []string{
			"echo hello",
		},
		DockerFilePath: dockerfileName,
		TimeoutMinutes: 10,
	}

	tf.RuntimeConfigs = []tests.RuntimeConfig{
		{
			Runtime: "kata-qemu-virtiofs",
			HypervisorConfigs: tests.HypervisorConfigs{
				CacheTypes: []string{
					"auto",
				},
				CacheSizesBytes: []int{
					0,
				},
				VirtiofsdArgs: []string{
					"",
				},
				KernelPaths: []string{"/opt/kata/share/kata-containers/vmlinux.container"},
			},
		},
		{
			Runtime: "kata-qemu",
			HypervisorConfigs: tests.HypervisorConfigs{
				KernelPaths: []string{"/opt/kata/share/kata-containers/vmlinux.container"},
			},
		},
		{
			Runtime: "runc",
		},
	}

	out, err := yaml.Marshal(tf)
	if err != nil {
		return err
	}

	workloadDir := "workloads/example"
	err = os.MkdirAll(workloadDir, 0744)
	if err != nil {
		return err
	}

	yamlPath := path.Join(workloadDir, "example.yaml")
	err = ioutil.WriteFile(yamlPath, out, 0600)
	if err != nil {
		return err
	}

	dockerfilePath := path.Join(workloadDir, dockerfileName)
	err = ioutil.WriteFile(dockerfilePath, []byte("FROM busybox"), 0600)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	app := &cli.App{}
	app.Name = "mrunner"
	app.Usage = "Run container workloads for diffent kata configs"
	app.Commands = []*cli.Command{
		{
			Name:  "template",
			Usage: "create a template workload",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				return createTemplate()
			},
		},
		{
			Name:  "run",
			Usage: "run a workload",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return errors.New("missing workload file")
				}
				return runWorkload(c.Args().First())
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
