package path

import (
	"os"
	"path/filepath"
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsDir(filepath string) (bool, error) {
	if fstat, err := os.Stat(filepath); err == nil {
		if fstat.IsDir() {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return false, err
	}
}

func IsFile(filepath string) (bool, error) {
	if fstat, err := os.Stat(filepath); err == nil {
		if fstat.IsDir() {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		return false, err
	}
}

func ListFiles(path string, recursive bool) ([]string, error) {
	// TODO recursive为false暂时不管用
	fs := make([]string, 0)

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		fs = append(fs, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fs, nil
}
