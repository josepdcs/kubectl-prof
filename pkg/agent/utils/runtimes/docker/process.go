// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"bufio"
	"errors"
	"github.com/fntlnz/mountinfo"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"io"
	"os"
	"path"
	"strings"
)

var (
	defaultProcessNames = map[api.ProgrammingLanguage]string{
		api.Java:   "java",
		api.Python: "python",
		api.Ruby:   "ruby",
	}
)

func getProcessName(job *config.ProfilingJob) string {
	if job.TargetProcessName != "" {
		return job.TargetProcessName
	}

	if val, ok := defaultProcessNames[job.Language]; ok {
		return val
	}

	return ""
}

func FindProcessId(job *config.ProfilingJob) (string, error) {
	name := getProcessName(job)
	foundProc := ""
	proc, err := os.Open("/proc")
	if err != nil {
		return "", err
	}

	defer proc.Close()

	for {
		dirs, err := proc.Readdir(15)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		for _, di := range dirs {
			if !di.IsDir() {
				continue
			}

			dname := di.Name()
			if dname[0] < '0' || dname[0] > '9' {
				continue
			}

			mi, err := mountinfo.GetMountInfo(path.Join("/proc", dname, "mountinfo"))
			if err != nil {
				continue
			}

			for _, m := range mi {
				root := m.Root
				if strings.Contains(root, job.PodUID) &&
					strings.Contains(root, job.ContainerName) {

					exeName, err := os.Readlink(path.Join("/proc", dname, "exe"))
					if err != nil {
						continue
					}

					if name != "" {
						// search by process name
						if strings.Contains(exeName, name) {
							return dname, nil
						}
					} else {
						if foundProc != "" {
							return "", errors.New("found more than one process on container," +
								" specify process name using --pgrep flag")
						} else {
							foundProc = dname
						}
					}
				}
			}
		}
	}

	if foundProc != "" {
		return foundProc, nil
	}

	return "", errors.New("could not find any process")
}

func FindRootProcessId(job *config.ProfilingJob) (string, error) {
	name := getProcessName(job)
	procsAndParents := make(map[string]string)
	proc, err := os.Open("/proc")
	if err != nil {
		return "", err
	}

	defer proc.Close()

	for {
		dirs, err := proc.Readdir(15)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		for _, di := range dirs {
			if !di.IsDir() {
				continue
			}

			dname := di.Name()
			if dname[0] < '0' || dname[0] > '9' {
				continue
			}

			mi, err := mountinfo.GetMountInfo(path.Join("/proc", dname, "mountinfo"))
			if err != nil {
				continue
			}

			for _, m := range mi {
				root := m.Root
				if strings.Contains(root, job.PodUID) &&
					strings.Contains(root, job.ContainerName) {

					exeName, err := os.Readlink(path.Join("/proc", dname, "exe"))
					if err != nil {
						continue
					}

					ppid, err := getProcessPPID(dname)
					if err != nil {
						return "", err
					}

					if name != "" {
						// search by process name
						if strings.Contains(exeName, name) {
							procsAndParents[dname] = ppid
						}
					} else {
						procsAndParents[dname] = ppid
					}
				}
			}
		}
	}

	return findRootProcess(procsAndParents)
}

func getProcessPPID(pid string) (string, error) {
	ppidKey := "PPid"
	statusFile, err := os.Open(path.Join("/proc", pid, "status"))
	if err != nil {
		return "", err
	}

	defer statusFile.Close()
	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, ppidKey) {
			return strings.Fields(text)[1], nil
		}
	}

	return "", errors.New("unable to get process ppid")
}

func findRootProcess(procsAndParents map[string]string) (string, error) {
	for process, ppid := range procsAndParents {
		if _, ok := procsAndParents[ppid]; !ok {
			// Found process with ppid that is not in the same programming language - this is the root
			return process, nil
		}
	}

	return "", errors.New("could not find root process")
}
