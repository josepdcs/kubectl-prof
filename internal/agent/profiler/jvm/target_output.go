package jvm

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const (
	outputDirPrefix = "kubectl-prof-"
	outputDirMode   = 0o755
	outputFileMode  = 0o600
)

type targetCreds struct {
	uid int
	gid int
}

var readTargetCredsFn = readTargetCreds

// Resolve /tmp through procfs to honor mounts in the target's namespace.
var targetTmpDir = func(pid string) string {
	return filepath.Join("/proc", pid, "root", "tmp")
}

// readTargetCreds reads the filesystem IDs used for permission checks. The owner of
// /proc/<pid> can instead reflect the process's dumpability state.
func readTargetCreds(pid string) (targetCreds, error) {
	statusPath := filepath.Join("/proc", pid, "status")
	f, err := os.Open(statusPath)
	if err != nil {
		return targetCreds{}, errors.Wrapf(err, "could not read the credentials of target process %s", pid)
	}
	defer func() { _ = f.Close() }()

	creds, err := parseTargetCreds(f)
	if err != nil {
		return targetCreds{}, errors.Wrapf(err, "could not read %s", statusPath)
	}

	return creds, nil
}

func parseTargetCreds(r io.Reader) (targetCreds, error) {
	var err error
	creds := targetCreds{uid: -1, gid: -1}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "Uid:"):
			creds.uid, err = filesystemID(line)
		case strings.HasPrefix(line, "Gid:"):
			creds.gid, err = filesystemID(line)
		default:
			continue
		}
		if err != nil {
			return targetCreds{}, errors.Wrapf(err, "could not parse %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return targetCreds{}, err
	}
	if creds.uid < 0 || creds.gid < 0 {
		return targetCreds{}, errors.New("no Uid and Gid lines found")
	}

	return creds, nil
}

func filesystemID(line string) (int, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, errors.Errorf("expected 4 ids, found %d", len(fields)-1)
	}
	return strconv.Atoi(fields[4])
}

// outputDir holds a descriptor to a per-run directory in the target's /tmp. Descriptor-
// relative operations remain anchored to the created directory if the workload renames
// its entry in a world-writable volume.
type outputDir struct {
	root         *os.Root
	hostPath     string
	containerDir string
}

// newOutputDir uses an unpredictable name so the workload cannot pre-create it.
func newOutputDir(pid string) (*outputDir, error) {
	suffix, err := randomSuffix()
	if err != nil {
		return nil, err
	}

	name := outputDirPrefix + suffix
	hostPath := filepath.Join(targetTmpDir(pid), name)
	containerDir := filepath.Join("/tmp", name)

	if err := os.Mkdir(hostPath, outputDirMode); err != nil {
		return nil, errors.Wrapf(err, "could not create the result directory %s inside the target container; "+
			"its /tmp must be writable (mount an emptyDir at /tmp, or profile with --tool jcmd)", containerDir)
	}

	root, err := os.OpenRoot(hostPath)
	if err != nil {
		_ = os.Remove(hostPath)
		return nil, errors.Wrapf(err, "could not open the result directory %s", hostPath)
	}

	return &outputDir{root: root, hostPath: hostPath, containerDir: containerDir}, nil
}

// createResultFile transfers ownership to the target because async-profiler truncates an
// existing output file without unlinking it. Keeping the descriptor open also prevents
// the workload from redirecting the later read through a renamed path.
func (d *outputDir) createResultFile(name string, creds targetCreds) (*os.File, error) {
	f, err := d.root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR|syscall.O_NOFOLLOW, outputFileMode)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create the result file %s", d.containerPath(name))
	}

	if err := f.Chown(creds.uid, creds.gid); err != nil {
		_ = f.Close()
		_ = d.root.Remove(name)
		return nil, errors.Wrapf(err, "could not transfer the result file %s to uid %d gid %d",
			d.containerPath(name), creds.uid, creds.gid)
	}

	return f, nil
}

func (d *outputDir) containerPath(name string) string {
	return filepath.Join(d.containerDir, name)
}

func (d *outputDir) remove() error {
	if d.root != nil {
		_ = d.root.Close()
	}
	return os.RemoveAll(d.hostPath)
}

func randomSuffix() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", errors.Wrap(err, "could not generate a unique result directory name")
	}
	return hex.EncodeToString(b), nil
}
