package version

import (
	"fmt"
	"runtime"
)

// populated by goreleaser
var (
	semver string
)

type Version struct {
	Version   string `json:"version"`
	GoVersion string `json:"go-version"`
}

func String() string {
	return fmt.Sprintf("Version: %s, Go Version: %s", semver, runtime.Version())
}

func GetCurrent() string {
	return semver
}
