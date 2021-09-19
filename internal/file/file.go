package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func Copy(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", src, err)
	}
	if info.IsDir() {
		return copyDir(src, dst, info)
	}
	return copyFile(src, dst, info)
}

func copyFile(src, dst string, info os.FileInfo) error {
	sf, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", src, err)
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to open/create file %q: %w", dst, err)
	}
	defer df.Close()

	if _, err := io.Copy(df, sf); err != nil {
		return fmt.Errorf("failed to copy %q to %q: %w", src, dst, err)
	}
	return nil
}

func copyDir(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dst, err)
	}
	contents, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read contents of directory %q: %w", src, err)
	}

	for _, item := range contents {
		srcItem := filepath.Join(src, item.Name())
		dstItem := filepath.Join(dst, item.Name())
		info, err := item.Info()
		if err != nil {
			return fmt.Errorf("failed to get info of %q: %w", srcItem, err)
		}

		if item.IsDir() {
			if err := copyDir(srcItem, dstItem, info); err != nil {
				return fmt.Errorf("failed to copy directory %q: %w", srcItem, err)
			}
			continue
		}
		if err := copyFile(srcItem, dstItem, info); err != nil {
			return fmt.Errorf("failed to copy file %q: %w", srcItem, err)
		}
	}
	return nil
}
