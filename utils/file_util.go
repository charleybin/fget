package utils

import (
	"os"
)

func CheckFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func MakeDirs(path string) bool {
	if err := os.MkdirAll(path, os.ModePerm); err == nil {
		return true
	}

	return false
}

func MakeDir(path string) bool {
	if err := os.Mkdir(path, os.ModePerm); err == nil {
		return true
	}

	return false
}

func CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}
