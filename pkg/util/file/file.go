package file

import (
	"crypto/md5"
	"encoding/hex"
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

// List lists all files that accomplish the given pattern
func List(pattern string) []string {
	var files []string
	matches, _ := filepath.Glob(pattern)
	if matches != nil {
		files = append(files, matches...)
	}
	return matches
}

// Remove removes a file
func Remove(file string) error {
	if Exists(file) {
		log.DebugLogLn("Trying to remove file: " + file)
		return os.Remove(file)
	}
	return nil
}

// RemoveAll removes all files that accomplish the given pattern
func RemoveAll(dir string, pattern string) {
	files := List(filepath.Join(dir, "*"+pattern+"*"))
	for _, f := range files {
		err := Remove(f)
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("file could no be removed: %s", err))
		}
	}
}

// GetSize returns the file size
func GetSize(file string) int64 {
	fileInfo, err := os.Stat(file)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("file could no be obtained: %s", err))
		return 0
	}
	return fileInfo.Size()
}

// IsEmpty returns if file is empty
func IsEmpty(file string) bool {
	return GetSize(file) == 0
}

// Write writes a file with the given content
func Write(file string, content string) {
	err := os.WriteFile(file, []byte(content), 0666)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("file [%s] could no be written: %s", file, err))
	}
}

// Read reads a file and returns its contents
func Read(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("file [%s] could no be read: %s", file, err))
	}

	return string(content)
}

// Append appends content to a file
func Append(file string, content string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("file [%s] could no be open to append content: %s", file, err))
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("file [%s] could no be closed after append content: %s", file, err))
		}
	}(f)
	if _, err := f.WriteString(content); err != nil {
		log.WarningLogLn(fmt.Sprintf("file [%s] could no be written to append content: %s", file, err))
	}
}

// GetChecksum returns the checksum of a file applying md5
// if file could not be read, returns empty string
func GetChecksum(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("file [%s] could no be read: %s", file, err))
		return ""
	}
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

// MergeFiles merges all files into a single one
func MergeFiles(outputPath string, inputPaths []string) {
	for _, f := range inputPaths {
		Append(outputPath, Read(f))
	}
}
