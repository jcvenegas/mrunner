package docker

import (
	"fmt"
	"mrunner/entities/tests"
	"path"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
)

type Test struct {
	Name           string
	Command        string
	Exec           []string
	PreExec        []string
	DockerFilePath string
	volumes        []ContainerVolume
	Timeout        time.Duration
}

func (d *Test) Setup() error {
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

func (d *Test) Run(e tests.TestEnv) (tests.Result, error) {
	result := tests.Result{TestID: d.Name}
	fmt.Println("Running test:", result.TestID)
	s := sh.NewSession()
	s.PipeFail = true
	s.ShowCMD = true
	s.SetDir(e.WorkDir)

	dockerArgs := []string{}
	dockerArgs = append(dockerArgs, "run", "-dti", "--runtime", e.Runtime)

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

	dockerArgs = append(dockerArgs, "--name", d.Name, d.Name)
	if d.Command != "" {
		c := strings.Fields(d.Command)
		dockerArgs = append(dockerArgs, c...)
	}

	err := s.Command("docker", dockerArgs).Run()
	if err != nil {
		return result, err
	}

	for _, e := range d.PreExec {
		if e != "" {
			pexecCmd := []string{"-c", e}
			err = s.Command("bash", pexecCmd).SetTimeout(d.Timeout).Run()
			if err != nil {
				result.SetError(err)
			}
		}
	}
	for _, e := range d.Exec {
		if e != "" {
			dockerArgs = []string{}
			dockerArgs = append(dockerArgs, "exec", "-i", d.Name)
			dockerArgs = append(dockerArgs, strings.Fields(e)...)
			err = s.Command("docker", dockerArgs).SetTimeout(d.Timeout).Run()
			if err != nil {
				result.SetError(err)
			}
		}
	}
	return result, nil
}

func (d *Test) TearDown() error {
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
func (d *Test) ID() string {
	return d.Name
}

func (d *Test) AddVolume(v ContainerVolume) {
	d.volumes = append(d.volumes, v)
}
