package filesystem

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
)

func BuildAbsolutePathFromHome(path string) string {
	path, _ = homedir.Expand(path)
	return path
}

func CheckExists(path string) bool {
	path = BuildAbsolutePathFromHome(path)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CreateDirectory(path string) bool {
	path = BuildAbsolutePathFromHome(path)
	if IsDirectory(path) {
		return true
	}
	err := os.MkdirAll(path, 0755)
	return err == nil
}

func ForceTrailingSlash(path string) string {
	if string(path[len(path)-1]) != "/" {
		path += "/"
	}
	return path
}

func GetDirectoryContents(path string) []string {
	var fileNames = []string{}
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	return fileNames
}

func IsDirectory(path string) bool {
	path = BuildAbsolutePathFromHome(path)
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && stat.IsDir()
}

func IsFile(path string) bool {
	path = BuildAbsolutePathFromHome(path)
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}

func IsEmptyDirectory(path string) bool {
	path = BuildAbsolutePathFromHome(path)
	if IsDirectory(path) {
		if file, err := os.Open(path); err == nil {
			contents, err := file.Readdir(1)

			if err != nil && err != io.EOF {
				panic(err)
			}
			return len(contents) == 0
		}
	}
	return false
}

func LoadFileIfExists(path string) string {
	path = BuildAbsolutePathFromHome(path)
	if IsFile(path) {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		return string(contents)
	}
	return ""
}

func WriteFile(path string, data []byte, mode os.FileMode) {
	path = BuildAbsolutePathFromHome(path)
	ioutil.WriteFile(path, data, mode)
}
