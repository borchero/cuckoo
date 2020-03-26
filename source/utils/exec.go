package utils

import (
	"os"
	"os/exec"
)

// ExecutableExists returns whether a binary with the given name can be found in $PATH.
func ExecutableExists(executable string) bool {
	cmd := exec.Command("which", executable)
	return cmd.Run() == nil
}

// RunCommandWithEnv runs the specified command with its args and additionally sets a set of
// environment variables on top of global environment variables.
func RunCommandWithEnv(env []string, name string, params ...string) error {
	cmd := exec.Command(name, params...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCommand runs the specified command with global environment variables.
func RunCommand(name string, params ...string) error {
	return RunCommandWithEnv([]string{}, name, params...)
}
