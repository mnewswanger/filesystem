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
	Logger    *logrus.Logger
}

// BuildAbsolutePathFromHome builds an absolute path (i.e. /home/user/example) from a home-based path (~/example)
func (fs *Filesystem) BuildAbsolutePathFromHome(path string) (string, error) {
	fs.initialize()

	var err error
	var fields = logrus.Fields{
		"path":     path,
		"expanded": path,
	}

	fs.Logger.WithFields(fields).Debug("Expanding path")
	path, err = homedir.Expand(path)
	return path, err
}

// CheckExists checks to see if the provided path exists on the machine
func (fs *Filesystem) CheckExists(path string) bool {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}

	fs.Logger.WithFields(fields).Debug("Checking to see if path exists")
	if err == nil {
		if _, e := os.Stat(path); !os.IsNotExist(e) {
			return true
		}
	}
	return false
}

// CreateDirectory creates a directory on the machine
//   All children will be created (behavior matches mkdir -p)
func (fs *Filesystem) CreateDirectory(path string) (bool, error) {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}

	fs.Logger.WithFields(fields).Debug("Creating directory")
	if err == nil {
		if !fs.isDirectory(path) {
			err = os.MkdirAll(path, 0755)
			if err == nil {
				fs.Logger.WithFields(fields).Debug("Directory created successfully")
			}
		}
	}
	fs.Logger.WithFields(fields).Warn("Failed to create directory")
	return err == nil, err
}

// ForceTrailingSlash forces a trailing slash at the end of the path
func (fs *Filesystem) ForceTrailingSlash(path string) string {
	fs.initialize()

	if len(path) == 0 {
		return "/"
	}

	if string(path[len(path)-1]) != "/" {
		path += "/"
	}
	return path
}

// GetDirectoryContents gets the files and folders inside the provided path
func (fs *Filesystem) GetDirectoryContents(path string) ([]string, error) {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}
	var fileNames = []string{}
	var files []os.FileInfo

	fs.Logger.WithFields(fields).Debug("Listing directory contents")
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
func (fs *Filesystem) GetFileSHA256Checksum(path string) (string, error) {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}

	if err == nil {
		if fs.isFile(path) {
			var contents []byte

			contents, err = ioutil.ReadFile(path)
			if err == nil {
				var checksum = sha256.Sum256(contents)
				var checksumString = hex.EncodeToString(checksum[:32])
				fields = logrus.Fields{
					"path":     path,
					"checksum": checksumString,
				}
				fs.Logger.WithFields(fields).Debug("Computed file checksum")
				return checksumString, err
			}
		} else {
			err = errors.New(path + " is not a file")
		}
	}
	fs.Logger.WithFields(fields).Warn("Failed to retreive file checksum")
	return "", err
}

// IsDirectory returns when path exists and is a directory
// supports ~ expansion
func (fs *Filesystem) IsDirectory(path string) bool {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}

	fs.Logger.WithFields(fields).Debug("Checking to see if path is a directory")
	return err == nil && fs.isDirectory(path)
}

// Check to see if the path provided is a directory
func (fs *Filesystem) isDirectory(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && stat.IsDir()
}

// IsEmptyDirectory returns when path exists and is an empty directory
// supports ~ expansion
func (fs *Filesystem) IsEmptyDirectory(path string) bool {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}

	fs.Logger.WithFields(fields).Debug("Checking to see if path is an empty directory")
	if err == nil {
		if fs.isDirectory(path) {
			if file, err := os.Open(path); err == nil {
				contents, err := file.Readdir(1)

				if err == nil || err == io.EOF {
					return len(contents) == 0
				}
			}
		}
	}
	return false
}

// IsFile returns when path exists and is a file
// supports ~ expansion
func (fs *Filesystem) IsFile(path string) bool {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"path": path,
	}

	fs.Logger.WithFields(fields).Debug("Checking to see if path is a file")
	return err == nil && fs.isFile(path)
}

// isFile checks to see if the file exists on the filesystem
func (fs *Filesystem) isFile(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}

// LoadFileIfExists loads the contents of path into a string if the file exists
func (fs *Filesystem) LoadFileIfExists(path string) (string, error) {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)
	var fields = logrus.Fields{
		"file": path,
	}

	fs.Logger.WithFields(fields).Debug("Attempting to load file")
	if err == nil {
		if fs.isFile(path) {
			contents, err := ioutil.ReadFile(path)
			if err == nil {
				fs.Logger.WithFields(fields).Debug("File read successfully")
				return string(contents), err
			}
		} else {
			err = errors.New(path + " is not a file")
		}
	}
	fs.Logger.WithFields(fields).Info("Could not read file")
	return "", err
}

// RemoveDirectory removes the directory at path from the system
// If recursive is set to true, it will remove all children as well
func (fs *Filesystem) RemoveDirectory(path string, recursive bool) (bool, error) {
	fs.initialize()

	var err error
	path, err = fs.BuildAbsolutePathFromHome(path)

	var fields = logrus.Fields{
		"directory": path,
	}

	if err == nil {
		fs.Logger.WithFields(fields).Debug("Attempting to remove directory")
		if fs.isDirectory(path) {
			if recursive {
				fs.Logger.WithFields(fields).Debug("Removing directory with recursion")
				err = os.RemoveAll(path)
			} else {
				fs.Logger.WithFields(fields).Debug("Removing directory without recursion")
				err = os.Remove(path)
			}
			if err == nil {
				fs.Logger.WithFields(fields).Debug("Directory was removed")
				return true, err
			}
		} else {
			err = errors.New(path + " is not a directory")
		}
	}
	fs.Logger.WithFields(fields).Warn("Failed to remove directory")
	return false, err
}

// WriteFile writes contents of data to path
func (fs *Filesystem) WriteFile(path string, data []byte, mode os.FileMode) error {
	fs.initialize()

	var err error
	var fields = logrus.Fields{
		"filename": path,
		"mode":     mode,
	}

	path, err = fs.BuildAbsolutePathFromHome(path)
	if err == nil {
		fs.Logger.Debug("Writing file", path)
		err = ioutil.WriteFile(path, data, mode)
		if err == nil {
			fs.Logger.WithFields(fields).Debug("Successfully wrote file")
		} else {
			fs.Logger.WithFields(fields).Warn("Failed to write file")
		}
	}
	return err
}

func (fs *Filesystem) initialize() {
	if fs.Logger == nil {
		fs.Logger = logrus.New()

		switch fs.Verbosity {
		case 0:
			fs.Logger.Level = logrus.ErrorLevel
			break
		case 1:
			fs.Logger.Level = logrus.WarnLevel
			break
		case 2:
			fallthrough
		case 3:
			fs.Logger.Level = logrus.InfoLevel
			break
		default:
			fs.Logger.Level = logrus.DebugLevel
			break
		}
	}
}
