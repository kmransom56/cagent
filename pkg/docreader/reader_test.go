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
