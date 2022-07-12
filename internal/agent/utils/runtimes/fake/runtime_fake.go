package fake

import (
	"fmt"
)

type RuntimeFake struct {
}

func NewRuntimeFake() *RuntimeFake {
	return &RuntimeFake{}
}

func (c *RuntimeFake) RootFileSystemLocation(containerID string) (string, error) {
	return fmt.Sprintf("/root/fs/%s", containerID), nil
}

func (c *RuntimeFake) PID(containerID string) (string, error) {
	return fmt.Sprintf("PID_%s", containerID), nil
}
