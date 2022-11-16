package file

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"os"
	"path/filepath"
)

func Exists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func List(pattern string) []string {
	var files []string
	matches, _ := filepath.Glob(pattern)
	if matches != nil {
		files = append(files, matches...)
	}
	return matches
}

func Remove(file string) error {
	if Exists(file) {
		log.DebugLogLn("Trying to remove file: " + file)
		return os.Remove(file)
	}
	return nil
}

func RemoveAll(dir string, pattern string) {
	files := List(filepath.Join(dir, "*"+pattern+"*"))
	for _, f := range files {
		err := Remove(f)
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("file could no be removed: %s", err))
		}
	}
}
