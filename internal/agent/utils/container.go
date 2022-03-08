package utils

import (
	"regexp"
)

func NormalizeContainerID(containerID string) string {
	return regexp.MustCompile("docker://|cri-o://|containerd://").ReplaceAllString(containerID, "")
}
