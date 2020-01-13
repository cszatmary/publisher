package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func FileOrDirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func ReadYaml(path string, val interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", path)
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	err = dec.Decode(val)
	return errors.Wrapf(err, "failed to decode yaml file %s", path)
}

func copyFile(src, dest string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "failed to open source file %s", src)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "failed to create destination file %s", dest)
	}
	defer destFile.Close()

	err = os.Chmod(destFile.Name(), info.Mode())
	if err != nil {
		return errors.Wrapf(err, "failed to set mode for destination file %s", dest)
	}

	_, err = io.Copy(destFile, srcFile)
	return errors.Wrapf(err, "failed to copy %s to %s", src, dest)
}

func copyDir(src, dest string, info os.FileInfo) error {
	err := os.MkdirAll(dest, os.FileMode(0755))
	if err != nil {
		return errors.Wrapf(err, "failed to create destination directory %s", dest)
	}

	contents, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.Wrapf(err, "failed to read contents of source directory %s", src)
	}

	for _, file := range contents {
		srcFile := filepath.Join(src, file.Name())
		destFile := filepath.Join(dest, file.Name())
		err = copy(srcFile, destFile, file)
		if err != nil {
			return errors.Wrapf(err, "failed to copy %s to %s", src, dest)
		}
	}

	return nil
}

func copy(src, dest string, info os.FileInfo) error {
	if info.Mode().IsDir() {
		return copyDir(src, dest, info)
	}

	return copyFile(src, dest, info)
}

func Copy(src, dest string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "failed to get info of %s", src)
	}

	return copy(src, dest, srcStat)
}
