package shellutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectCommandShell_WindowsPrefersPwsh(t *testing.T) {
	t.Parallel()

	shell := detectCommandShell("windows", func(string) string { return "" }, func(name string) (string, error) {
		switch name {
		case "pwsh.exe":
			return `C:\Program Files\PowerShell\7\pwsh.exe`, nil
		default:
			return "", errors.New("not found")
		}
	})

	assert.Equal(t, `C:\Program Files\PowerShell\7\pwsh.exe`, shell.Path)
	assert.Equal(t, []string{"-NoProfile", "-NonInteractive", "-Command"}, shell.ArgsPrefix)
}

func TestDetectCommandShell_WindowsFallsBackToCmd(t *testing.T) {
	t.Parallel()

	shell := detectCommandShell("windows", func(key string) string {
		if key == "ComSpec" {
			return `C:\Windows\System32\cmd.exe`
		}
		return ""
	}, func(string) (string, error) {
		return "", errors.New("not found")
	})

	assert.Equal(t, `C:\Windows\System32\cmd.exe`, shell.Path)
	assert.Equal(t, []string{"/C"}, shell.ArgsPrefix)
}

func TestDetectCommandShell_UnixUsesShellEnv(t *testing.T) {
	t.Parallel()

	shell := detectCommandShell("linux", func(key string) string {
		if key == "SHELL" {
			return "/bin/bash"
		}
		return ""
	}, func(string) (string, error) {
		return "", errors.New("unused")
	})

	assert.Equal(t, "/bin/bash", shell.Path)
	assert.Equal(t, []string{"-c"}, shell.ArgsPrefix)
}
