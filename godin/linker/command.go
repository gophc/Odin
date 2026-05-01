package linker

import (
	"fmt"
	"os/exec"
	"strings"
)

// systemExecCommandLineApp executes a command.
// It mimics the original system_exec_command_line_app function.
func systemExecCommandLineApp(name string, command string) (int, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty command line")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command '%s' failed: %v\nOutput: %s\n", command, err, output)
		return 1, err
	}
	return 0, nil
}

// systemExecCommandLineAppOutput executes a command and returns its output as a string.
// It mimics the original system_exec_command_line_app_output function.
func systemExecCommandLineAppOutput(command string, output *gbString) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	output.WriteString(string(out))
	return true
}
