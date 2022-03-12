package test

import (
	"path"
	"path/filepath"
	"runtime"
)

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func CrioTestDataDir() string {
	return RootDir() + "/test/config/crio"
}

func DockerTestDataDir() string {
	return RootDir() + "/test/config/docker"
}
