package jvm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseTargetCreds(t *testing.T) {
	t.Run("should read the filesystem ids, not the real ones", func(t *testing.T) {
		status := strings.Join([]string{
			"Name:\tjava",
			"Uid:\t0\t999\t999\t999",
			"Gid:\t0\t998\t998\t998",
			"Threads:\t104",
		}, "\n")

		creds, err := parseTargetCreds(strings.NewReader(status))

		require.NoError(t, err)
		assert.Equal(t, 999, creds.uid)
		assert.Equal(t, 998, creds.gid)
	})

	t.Run("should read a process that never changed credentials", func(t *testing.T) {
		status := "Uid:\t1000\t1000\t1000\t1000\nGid:\t1000\t1000\t1000\t1000\n"

		creds, err := parseTargetCreds(strings.NewReader(status))

		require.NoError(t, err)
		assert.Equal(t, 1000, creds.uid)
		assert.Equal(t, 1000, creds.gid)
	})

	t.Run("should read a root process", func(t *testing.T) {
		status := "Uid:\t0\t0\t0\t0\nGid:\t0\t0\t0\t0\n"

		creds, err := parseTargetCreds(strings.NewReader(status))

		require.NoError(t, err)
		assert.Equal(t, 0, creds.uid)
		assert.Equal(t, 0, creds.gid)
	})

	t.Run("should fail when the Gid line is missing", func(t *testing.T) {
		_, err := parseTargetCreds(strings.NewReader("Uid:\t0\t0\t0\t0\n"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no Uid and Gid lines found")
	})

	t.Run("should fail when a line is truncated", func(t *testing.T) {
		_, err := parseTargetCreds(strings.NewReader("Uid:\t0\t0\nGid:\t0\t0\t0\t0\n"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected 4 ids")
	})

	t.Run("should fail when an id is not a number", func(t *testing.T) {
		_, err := parseTargetCreds(strings.NewReader("Uid:\t0\t0\t0\troot\nGid:\t0\t0\t0\t0\n"))

		require.Error(t, err)
	})
}

func Test_filesystemID(t *testing.T) {
	id, err := filesystemID("Uid:\t1\t2\t3\t4")

	require.NoError(t, err)
	assert.Equal(t, 4, id)
}
