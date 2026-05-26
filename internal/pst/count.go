package pst

import (
	"os"
	"path/filepath"
)

func CountMessages(root string) (int, error) {

	total := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

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

		name := filepath.Base(path)

		if name == ".DS_Store" ||
			name == "index" ||
			name == "desc" {
			return nil
		}

		if !isAllowedFolder(path) {
			return nil
		}

		total++

		return nil
	},
	)

	return total, err
}
