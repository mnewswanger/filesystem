package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

// Filesystem represents the base structure for the package
type Filesystem struct {
	Verbosity uint8
	logger    *logrus.Logger
}

// BuildAbsolutePathFromHome builds an absolute path (i.e. /home/user/example) from a home-based path (~/example)
func (fs Filesystem) BuildAbsolutePathFromHome(path string) (string, error) {
	fs.initialize()

	var err error

	path, err = homedir.Expand(path)
	fs.logger.Info("Expanding " + path)
	return path, err
}

// CheckExists checks to see if the provided path exists on the machine
func (fs Filesystem) CheckExists(path string) (bool, error) {
	fs.initialize()

	var err error

	path, err = fs.BuildAbsolutePathFromHome(path)
	if err == nil {
		if _, e := os.Stat(path); !os.IsNotExist(e) {
			return true, err
		}
	}
	return false, err
}

// CreateDirectory creates a directory on the machine
//   All children will be created (behavior matches mkdir -p)
func (fs Filesystem) CreateDirectory(path string) (bool, error) {
	fs.initialize()

	var err error

	path, err = fs.BuildAbsolutePathFromHome(path)
	if err != nil {
		if fs.IsDirectory(path) {
			return true, err
		}
		err = os.MkdirAll(path, 0755)
	}
	return err == nil, err
}

// ForceTrailingSlash forces a trailing slash at the end of the path
func (fs Filesystem) ForceTrailingSlash(path string) string {
	fs.initialize()

	if string(path[len(path)-1]) != "/" {
		path += "/"
	}
	return path
}

// GetDirectoryContents gets the files and folders inside the provided path
func (fs Filesystem) GetDirectoryContents(path string) ([]string, error) {
	fs.initialize()

	var err error
	var fileNames = []string{}
	var files []os.FileInfo

	files, err = ioutil.ReadDir(path)
	if err == nil {
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
		}
	}
	return fileNames, err
}

// GetFileSHA256Checksum gets the SHA-256 checksum of the file as a hex string
//   Output matches sha256sum (Linux) / shasum -a 256 (OSX)
func (fs Filesystem) GetFileSHA256Checksum(path string) (string, error) {
	fs.initialize()

	var err error

	path, err = fs.BuildAbsolutePathFromHome(path)
	if err == nil {
		if fs.IsFile(path) {
			var contents []byte

			contents, err = ioutil.ReadFile(path)
			if err == nil {
				checksum := sha256.Sum256(contents)
				return hex.EncodeToString(checksum[:32]), err
			}
		} else {
			err = errors.New(path + " is not a file")
		}
	}
	return "", err
}

// IsDirectory returns when path exists and is a directory
func (fs Filesystem) IsDirectory(path string) bool {
	fs.initialize()

	path, _ = fs.BuildAbsolutePathFromHome(path)
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && stat.IsDir()
}

// IsFile returns when path exists and is a file
func (fs Filesystem) IsFile(path string) bool {
	fs.initialize()

	path, _ = fs.BuildAbsolutePathFromHome(path)
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()

}

// IsEmptyDirectory returns when path exists and is an empty directory
func (fs Filesystem) IsEmptyDirectory(path string) bool {
	fs.initialize()

	path, _ = fs.BuildAbsolutePathFromHome(path)
	if fs.IsDirectory(path) {
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

// LoadFileIfExists loads the contents of path into a string if the file exists
func (fs Filesystem) LoadFileIfExists(path string) (string, error) {
	fs.initialize()

	var err error

	path, err = fs.BuildAbsolutePathFromHome(path)
	if err == nil {
		if fs.IsFile(path) {
			contents, err := ioutil.ReadFile(path)
			if err == nil {
				return string(contents), err
			}
		} else {
			err = errors.New(path + " is not a file")
		}
	}
	return "", err
}

// RemoveDirectory removes the directory at path from the system
// If recursive is set to true, it will remove all children as well
func (fs Filesystem) RemoveDirectory(path string, recursive bool) (bool, error) {
	fs.initialize()

	var err error

	if fs.IsDirectory(path) {
		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
		if err == nil {
			return true, err
		}
	} else {
		err = errors.New(path + " is not a directory")
	}
	return false, err
}

// WriteFile writes contents of data to path
func (fs Filesystem) WriteFile(path string, data []byte, mode os.FileMode) error {
	fs.initialize()

	var err error

	path, err = fs.BuildAbsolutePathFromHome(path)
	if err == nil {
		err = ioutil.WriteFile(path, data, mode)
	}
	return err
}

func (fs Filesystem) initialize() {
	if fs.logger == nil {
		fs.logger = logrus.New()

		switch fs.Verbosity {
		case 0:
			logrus.SetLevel(logrus.WarnLevel)
			break
		case 1:
			logrus.SetLevel(logrus.InfoLevel)
			break
		case 2:
			logrus.SetLevel(logrus.DebugLevel)
			break
		}
	}
}
