package shellutil

import (
	"cmp"
	"os"
	"os/exec"
	"runtime"
)

// CommandShell describes the shell binary and argument prefix used to execute a command string.
type CommandShell struct {
	Path       string
	ArgsPrefix []string
}

// DetectCommandShell returns the preferred non-interactive command shell for the current OS.
func DetectCommandShell() CommandShell {
	return detectCommandShell(runtime.GOOS, os.Getenv, exec.LookPath)
}

func detectCommandShell(goos string, getenv func(string) string, lookPath func(string) (string, error)) CommandShell {
	if goos == "windows" {
		powershellArgs := []string{"-NoProfile", "-NonInteractive", "-Command"}
		for _, candidate := range []string{"pwsh.exe", "powershell.exe"} {
			if path, err := lookPath(candidate); err == nil {
				return CommandShell{Path: path, ArgsPrefix: powershellArgs}
			}
		}

		return CommandShell{Path: cmp.Or(getenv("ComSpec"), "cmd.exe"), ArgsPrefix: []string{"/C"}}
	}

	return CommandShell{Path: cmp.Or(getenv("SHELL"), "/bin/sh"), ArgsPrefix: []string{"-c"}}
}
