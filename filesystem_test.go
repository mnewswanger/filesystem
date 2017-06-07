package filesystem

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var fs = Filesystem{
	Verbosity: 4,
}

func TestHomedirExpansion(t *testing.T) {
	color.Yellow("Testing ~ expansion functionality")
	var expandedPath, err = fs.BuildAbsolutePathFromHome("~/test-dir")
	if err != nil {
		t.Error(err)
	}
	if strings.Index(expandedPath, "/") != 0 {
		t.Error("Absolute path should start with /")
	}
	color.Yellow("Test complete")
	println()
}

func TestFilesystemOperations(t *testing.T) {
	color.Yellow("Testing Filesystem Operations")
	// Create a temp directory
	var tempDir, err = ioutil.TempDir("/tmp/", ".filesystem-test-")
	if err != nil {
		t.Error(err)
	}

	// Test CheckExists - the temp directory now exists since err == nil above
	if !fs.CheckExists(tempDir) {
		t.Error("Check exists failed; returned false:", tempDir)
	}
	// Test Remaining Directory Functions
	testFilesystemOperations(t, tempDir+"/does-not-exist")
	testFilesystemOperations(t, tempDir+"/this/has/subfolders/that/dont/exist")

	// Verify the loading a non-existent files / folders returns properly
	color.Yellow("Test failure handling")
	if c, err := fs.LoadFileString(tempDir + "/file-dne"); err == nil || c != "" {
		t.Error("Load non-existent file string test failed")
	}
	if c, err := fs.LoadFileBytes(tempDir + "/file-dne"); err == nil || !reflect.DeepEqual(c, []byte{}) {
		t.Error("Load non-existent file bytes test failed")
	}
	if fs.IsEmptyDirectory(tempDir + "/dne/") {
		t.Error("Non-existent directory says it's an empty directory")
	}
	if c, err := fs.GetFileSHA256Checksum(tempDir + "/file-dne"); err == nil || c != "" {
		t.Error("Non-existent file checksum test failed")
	}
	if err := fs.WriteFile(tempDir+"/dne/expect-error", []byte("test"), 0644); err == nil {
		t.Error("Write should have failed but did not return an error")
	}

	fs.RemoveDirectory(tempDir, true)
	if fs.CheckExists(tempDir) {
		t.Error("Recursive directory removal failed; returned true:", tempDir)
	}
	color.Yellow("Test Complete")
	println()
}

func testFilesystemOperations(t *testing.T, dir string) {
	color.Yellow("Make sure the directory doesn't exist before starting")
	if fs.CheckExists(dir) {
		t.Error("Directory already exists:" + dir)
	}

	color.Yellow("Attempt to create the directory")
	if err := fs.CreateDirectory(dir); err != nil {
		t.Error("Create directory failed:", dir, err)
	}
	if !fs.CheckExists(dir) {
		t.Error("Create diretory did not create the directory properly:", dir)
	}

	color.Yellow("Try creating again now that the directory exists")
	if err := fs.CreateDirectory(dir); err != nil {
		t.Error("Create directory failed:", dir, err)
	}
	if c, err := fs.GetDirectoryContents(dir); err != nil || len(c) > 0 {
		t.Error("Directory is not empty:", dir)
	}

	color.Yellow("Remove the directory")
	if err := fs.RemoveDirectory(dir, false); err != nil {
		t.Error("Directory could not be deleted:", dir)
	}
	if fs.CheckExists(dir) {
		t.Error("Directory should have been removed but was found:", dir)
	}

	color.Yellow("Recreate the directory")
	if err := fs.CreateDirectory(dir); err != nil {
		t.Error("Create directory failed:", dir, err)
	}
	if !fs.CheckExists(dir) {
		t.Error("Create diretory did not create the directory properly:", dir)
	}
	if !fs.IsEmptyDirectory(dir) {
		t.Error("Directory is reported as not empty but has no contents:", dir)
	}

	color.Yellow("Write a file inside the directory")
	var testFile = dir + "/test.file"
	var testFileContents = "test"
	var testFileBytes = []byte(testFileContents)
	if err := fs.WriteFile(testFile, testFileBytes, 0644); err != nil {
		t.Error("Error occured while writing file:", testFile, err)
	}
	if !fs.IsDirectory(dir) {
		t.Error("IsDirectory test failed: ", dir, " is a directory")
	}
	if fs.IsDirectory(testFile) {
		t.Error("IsDirectory test failed: ", testFile, " is a file")
	}
	if fs.IsFile(dir) {
		t.Error("IsFile test failed: ", dir, " is a directory")
	}
	if !fs.IsFile(testFile) {
		t.Error("IsDirectory test failed: ", testFile, " is a file")
	}

	// Make sure the folder now has contents
	if c, err := fs.GetDirectoryContents(dir); err != nil || len(c) == 0 {
		t.Error("Directory is empty and should not be: ", dir)
	}
	if fs.IsEmptyDirectory(dir) {
		t.Error("Directory is reported as empty but has contents:", dir)
	}

	// Verify the contents of the file match what was intended
	if c, err := fs.LoadFileString(testFile); err != nil || c != testFileContents {
		t.Error("File contents (string) don't match what was saved: ", testFile, "Got:", c, "Wanted:", testFileContents)
	}
	if c, err := fs.LoadFileBytes(testFile); err != nil || !reflect.DeepEqual(c, testFileBytes) {
		t.Error("File contents (bytes) don't match what was saved: ", testFile, "Got:", c, "Wanted:", testFileBytes)
	}
	// Test deprecated function call
	if c, err := fs.LoadFileIfExists(testFile); err != nil || c != testFileContents {
		t.Error("File contents don't match what was saved: ", testFile, "Got:", c, "Wanted:", testFileContents)
	}

	// Verify the file checksum
	if c, err := fs.GetFileSHA256Checksum(testFile); err != nil || c != "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" {
		t.Error("File checksum incorrect for:", testFile, "Got:", c, "Wanted:", "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")
	}

	color.Yellow("Remove the directory")
	if err := fs.RemoveDirectory(dir, true); err != nil {
		t.Error("Directory could not be deleted:", dir)
	}
	if fs.CheckExists(dir) {
		t.Error("Directory should have been removed but was found: " + dir)
	}

	// Attempt to remove the directory again
	if err := fs.RemoveDirectory(dir, false); err == nil {
		t.Error("Directory could not be deleted:", dir)
	}
}

func TestTrailingSlash(t *testing.T) {
	color.Yellow("Testing force trailing slash functionality")
	var testData = map[string]string{
		"":        "/",
		"/test/":  "/test/",
		"/test":   "/test/",
		"/test//": "/test//",
	}

	var got string
	for k, v := range testData {
		got = fs.ForceTrailingSlash(k)
		if got != v {
			t.Error("ForceTrailingSlash failed:", "Got:", got, "Wanted:", v)
		}
	}
	color.Yellow("Test Complete")
	println()
}

func TestLoggingOptions(t *testing.T) {
	color.Yellow("Testing logging options")
	var err error
	// Try different verbosity levels
	for i := uint8(0); i <= 3; i++ {
		fs = Filesystem{
			Verbosity: i,
		}
		_, err = fs.BuildAbsolutePathFromHome("~/test/")
		if err != nil {
			t.Error(err)
		}
	}
	// Try passing in
	var logger = logrus.New()
	fs = Filesystem{
		Logger: logger,
	}
	fs.BuildAbsolutePathFromHome("~/test")
	color.Yellow("Test Complete")
	println()
}

func TestFileExtensionFunctionality(t *testing.T) {
	var extensionTestData = map[string]string{
		"none":                "",
		"file.ext":            "ext",
		"file.bk.ext":         "ext",
		"/full/path.txt":      "txt",
		"~/relative/path.pdf": "pdf",
		"test.":               "",
	}

	var got string
	for value, expected := range extensionTestData {
		got = fs.GetFileExtension(value)
		if got != expected {
			t.Error("Got back unexpected extension.", "Expected:", expected, "Got:", got)
		}
	}
}
