package testdata

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
	return RootDir() + "/testdata/crio"
}

func ContainerdTestDataDir() string {
	return RootDir() + "/testdata/containerd"
}

func ResultTestDataDir() string {
	return RootDir() + "/testdata/result"
}

func ConfigmapsTestDataDir() string {
	return RootDir() + "/testdata/configmaps"
}
