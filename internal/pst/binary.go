package pst

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func ReadPSTBinary() string {

	// Linux/macOS
	if runtime.GOOS != "windows" {
		if path, err := exec.LookPath("readpst"); err == nil {
			return path
		}

		return "readpst"
	}

	// Windows
	exe, err := os.Executable()

	if err != nil {
		return filepath.Join(
			"readpst-win",
			"readpst.exe",
		)
	}

	base := filepath.Dir(exe)

	candidates := []string{
		filepath.Join(base, "readpst-win", "readpst.exe"),
		filepath.Join(base, "readpst", "readpst.exe"),
		filepath.Join(base, "readpst.exe"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return filepath.Join(
		"readpst-win",
		"readpst.exe",
	)
}
