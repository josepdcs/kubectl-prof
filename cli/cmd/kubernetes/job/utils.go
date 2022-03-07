//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.
package job

import (
	"github.com/josepdcs/kubectl-flame/cli/cmd/data"
)

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

func getContainerPath(c *data.FlameConfig) string {
	if len(c.TargetConfig.DockerPath) > 0 {
		return c.TargetConfig.DockerPath
	}
	if len(c.TargetConfig.CrioPath) > 0 {
		return c.TargetConfig.CrioPath
	}
	return c.TargetConfig.ContainerdPath
}
