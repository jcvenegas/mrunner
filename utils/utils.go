package utils

import (
	"io/ioutil"
	"os"
)

func CopyFile(f, dest string, p os.FileMode) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil
	}
	return ioutil.WriteFile(dest, data, p)
}
