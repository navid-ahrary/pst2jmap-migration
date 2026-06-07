package pst

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type Reader struct {
	OutputDir string
}

var allowedFolders = map[string]bool{
	"Inbox": true,

	"Sent-Items": true,
	"Sent Items": true,

	"Deleted-Items": true,
	"Deleted Items": true,

	"Drafts": true,

	"Junk-Email": true,
	"Junk Email": true,
}

func NewReader(ctx context.Context, pstFile string) (*Reader, error) {
	outputDir := "output"

	// cleanup old extraction
	_ = os.RemoveAll(outputDir)

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return nil, err
	}

	err = ExtractPST(ctx, pstFile, outputDir)

	if err != nil {
		_ = os.RemoveAll(outputDir)
		return nil, err
	}

	return &Reader{OutputDir: outputDir}, nil
}

func ExtractPST(ctx context.Context, pstFile string, outputDir string) error {
	cmd := exec.CommandContext(ctx, ReadPSTBinary(), "-D", "-b", "-e", "-o", outputDir, pstFile)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r *Reader) Walk(fn func(string) error) error {
	return filepath.Walk(r.OutputDir,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if info == nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if info.Size() == 0 {
				return nil
			}

			if filepath.Ext(path) != ".eml" {
				return nil
			}

			if !isAllowedFolder(path) {
				return nil
			}

			return fn(path)
		},
	)
}

func (r *Reader) Close() error { return nil }

func isAllowedFolder(path string) bool {
	dir := filepath.Dir(path)

	for {
		name := filepath.Base(dir)

		if allowedFolders[name] {
			return true
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return false
}
