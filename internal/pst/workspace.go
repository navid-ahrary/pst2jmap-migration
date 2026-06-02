package pst

import (
	"path/filepath"
	"strings"
)

func MigrationDir(pstPath string) string {
	base := strings.TrimSuffix(pstPath, filepath.Ext(pstPath))
	return base + ".migration"
}
