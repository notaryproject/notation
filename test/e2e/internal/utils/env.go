package utils

import (
	"os"
	"strings"
)

// SetUpUserDir creates a tmp directory.
//
// Calling cleaner will remove the directory recursively.
func SetUpUserDir() (dir string, cleaner func(), err error) {
	dir, err = os.MkdirTemp("/tmp", "notatione2e")
	cleaner = func() {
		os.RemoveAll(dir)
	}
	return
}

// SetUpContainer creates a running container from NotationBinaryImage to run notation commands.
//
// Calling cleaner will stop and remove the container.
func SetUpContainer() (containerID string, cleaner func(), err error) {
	args := []string{"run", "-di", "--net=host", NotationBinaryImage, "bin/sh"}
	session, err := Exec("docker", ExecOpts{}, args...)
	containerID = strings.TrimSpace(string(session.Out.Contents()))
	cleaner = func() {
		Exec("docker", ExecOpts{}, "stop", containerID)
		Exec("docker", ExecOpts{}, "rm", containerID)
	}
	return
}
