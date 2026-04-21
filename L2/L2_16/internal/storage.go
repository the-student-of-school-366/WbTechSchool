package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

type Storage struct {
	outputDir string
}

func NewStorage(outputDir string) *Storage {
	return &Storage{
		outputDir: outputDir,
	}
}

func (s *Storage) URLToLocalPath(rawURL string) string {
	return URLToLocalPath(rawURL, s.outputDir)
}

func (s *Storage) SaveHTML(doc *goquery.Document, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating directories for %s: %w", localPath, err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", localPath, err)
	}
	defer file.Close()

	html, err := doc.Html()
	if err != nil {
		return fmt.Errorf("rendering HTML: %w", err)
	}

	if _, err := file.WriteString(html); err != nil {
		return fmt.Errorf("writing HTML to %s: %w", localPath, err)
	}

	return nil
}

func (s *Storage) SaveFile(localPath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm); err != nil {
		return fmt.Errorf("creating directories for %s: %w", localPath, err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", localPath, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("writing data to %s: %w", localPath, err)
	}

	return nil
}
