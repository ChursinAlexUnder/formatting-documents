package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	appRootOnce sync.Once
	appRoot     string
)

func AppRoot() string {
	appRootOnce.Do(func() {
		if root := os.Getenv("APP_ROOT"); root != "" {
			appRoot = root
			return
		}

		wd, err := os.Getwd()
		if err != nil {
			appRoot = "."
			return
		}

		if looksLikeAppRoot(wd) {
			appRoot = wd
			return
		}

		parent := filepath.Dir(wd)
		if looksLikeAppRoot(parent) {
			appRoot = parent
			return
		}

		appRoot = wd
	})

	return appRoot
}

func RootPath(parts ...string) string {
	allParts := append([]string{AppRoot()}, parts...)
	return filepath.Join(allParts...)
}

func BufferDir() string {
	if dir := os.Getenv("APP_BUFFER_DIR"); dir != "" {
		return dir
	}
	return RootPath("buffer")
}

func DataFile() string {
	if file := os.Getenv("APP_DATA_FILE"); file != "" {
		return file
	}
	return RootPath("data.json")
}

func PythonBin() string {
	if bin := os.Getenv("PYTHON_BIN"); bin != "" {
		return bin
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

func EnsureRuntimeDirs() error {
	return os.MkdirAll(BufferDir(), 0755)
}

func looksLikeAppRoot(path string) bool {
	if _, err := os.Stat(filepath.Join(path, "web", "templates")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, "scripts", "editdocument.py")); err != nil {
		return false
	}
	return true
}
