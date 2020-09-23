package docker

import (
	"fmt"
	"mrunner/entities/tests"
	"path"
	"time"

	"github.com/codeskyblue/go-sh"
)

type DockerTest struct {
	Name           string
	Command        string
	Exec           string
	DockerFilePath string
	volumes        []ContainerVolume
	Timeout        time.Duration
}

func (d *DockerTest) Setup() error {
	fmt.Println("Running setup")
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true
	dockerFileDir := path.Dir(d.DockerFilePath)
	s.SetDir(dockerFileDir)

	dockerFile := path.Base(d.DockerFilePath)

	err := s.Command("docker", "build", "-f", dockerFile, "-t", d.Name, ".").Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerTest) Run(e tests.TestEnv) (tests.TestsResult, error) {
	result := tests.TestsResult{TestID: d.Name}
	fmt.Println("Running test")
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true
	s.SetDir(e.WorkDir)

	dockerArgs := []string{}
	dockerArgs = append(dockerArgs, "run")
	dockerArgs = append(dockerArgs, "-dti")

	for _, v := range d.volumes {
		dockerArgs = append(dockerArgs, "-v")

		hostPath, err := v.Host()
		if err != nil {
			return result, err
		}

		containerPath, err := v.Container()
		if err != nil {
			return result, err
		}
		dockerArgs = append(dockerArgs, hostPath+":"+containerPath)
	}

	dockerArgs = append(dockerArgs, "--name")
	dockerArgs = append(dockerArgs, d.Name)
	dockerArgs = append(dockerArgs, d.Name)

	err := s.Command("docker", dockerArgs).Run()
	if err != nil {
		return result, err
	}

	if d.Exec != "" {
		err = s.Command("docker", "exec", "-i", d.Name, "sh", "-c", d.Exec).SetTimeout(d.Timeout).Run()
		if err != nil {
			result.SetError(err)
		}
	}
	return result, nil
}

func (d *DockerTest) TearDown() error {
	fmt.Println("Running teardown")
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true

	err := s.Command("docker", "rm", "-f", d.Name).Run()
	if err != nil {
		return err
	}
	return nil
}
func (d *DockerTest) ID() string {
	return d.Name
}

func (d *DockerTest) AddVolume(v ContainerVolume) {
	d.volumes = append(d.volumes, v)
}
