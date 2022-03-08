package version

import (
	"fmt"
	"runtime"
)

// populated by goreleaser
var (
	semver string
	commit string
	date   string
)

type Version struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go-version"`
}

func String() string {
	return fmt.Sprintf("Version: %s, Commit: %s, Build Date: %s, Go Version: %s", semver, commit, date, runtime.Version())
}

func GetCurrent() string {
	return semver
}
