package docreader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadText_ReadsPlainTextFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello world"), 0o644))

	content, err := ReadText(path)
	require.NoError(t, err)
	assert.Equal(t, "hello world", content)
}

func TestReadText_UsesPDFReaderForPDFExtension(t *testing.T) {
	original := pdfTextReader
	t.Cleanup(func() {
		pdfTextReader = original
	})

	pdfTextReader = func(path string) (string, error) {
		return "pdf text from " + filepath.Base(path), nil
	}

	content, err := ReadText(filepath.Join(t.TempDir(), "design.pdf"))
	require.NoError(t, err)
	assert.Equal(t, "pdf text from design.pdf", content)
}

func TestReadText_WrapsPDFErrors(t *testing.T) {
	original := pdfTextReader
	t.Cleanup(func() {
		pdfTextReader = original
	})

	pdfTextReader = func(string) (string, error) {
		return "", errors.New("boom")
	}

	_, err := ReadText(filepath.Join(t.TempDir(), "broken.pdf"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "reading pdf")
	assert.ErrorContains(t, err, "boom")
}

func TestResolvePath_TranslatesWindowsPathToWSL(t *testing.T) {
	t.Parallel()

	resolution := resolvePath("/workspace/project", `C:\Users\Keith\docs\spec.pdf`, "linux", func(string) (string, bool) {
		return "", false
	})

	assert.Equal(t, "/mnt/c/Users/Keith/docs/spec.pdf", resolution.ResolvedPath)
	assert.Contains(t, resolution.Warnings, "translated Windows path into WSL mount form")
}

func TestResolvePath_RemapsWindowsPathIntoBase(t *testing.T) {
	t.Parallel()

	resolution := resolvePath("/workspace/project", `C:\Users\Keith\project\documents\spec.pdf`, "linux", func(string) (string, bool) {
		return "", false
	})

	assert.Equal(t, "/workspace/project/documents/spec.pdf", resolution.ResolvedPath)
	assert.Contains(t, resolution.Warnings, "remapped Windows host path into the current workspace")
}

func TestResolvePath_UsesConfiguredPathMap(t *testing.T) {
	t.Parallel()

	resolution := resolvePath("/workspace/project", `D:\Shared\plans\roadmap.pdf`, "linux", func(key string) (string, bool) {
		if key == "CAGENT_PATH_MAP" {
			return `D:\Shared=/mounted/shared`, true
		}
		return "", false
	})

	assert.Equal(t, "/mounted/shared/plans/roadmap.pdf", filepath.ToSlash(resolution.ResolvedPath))
	assert.Contains(t, resolution.Warnings, "resolved path using CAGENT_PATH_MAP")
}
