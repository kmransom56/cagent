package docreader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

var pdfTextReader = readPDFText

// ReadText reads a document and returns text content.
// Text-based files are returned as-is; PDFs are extracted natively.
func ReadText(path string) (string, error) {
	if strings.EqualFold(filepath.Ext(path), ".pdf") {
		text, err := pdfTextReader(path)
		if err != nil {
			return "", fmt.Errorf("reading pdf %s: %w", path, err)
		}
		return text, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}

	return string(data), nil
}

func readPDFText(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(textReader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
