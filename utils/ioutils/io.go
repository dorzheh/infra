// Generic functions

package ioutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// FindAndReplace is looking for appropriate pattern
// in the file and replace it with a new one
// Could be also applied for generating a new file
func FindAndReplace(srcFile, dstFile, oldPattern, newPattern string,
	filePermissions os.FileMode) error {

	srcFileBuf, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	// convert the string pattern to the slice of bytes
	newPat := []byte(newPattern)
	// compile reg exp
	rgxp, _ := regexp.Compile(oldPattern)
	// Not so fast : returns a new copy of the slice (copying here)
	newBuf := rgxp.ReplaceAll(srcFileBuf, newPat)
	return ioutil.WriteFile(dstFile, newBuf, filePermissions)
}

// FindAndReplaceMulti is looking for appropriate patterns
// in the file and replace them with new patterns
// The patterns must be represented as a map where key is the old pattern and value
// is a new one
// Could be also applied for generating a new file
func FindAndReplaceMulti(srcFile, dstFile string, patternMap map[string]string,
	filePermissions os.FileMode) error {
	fileBuf, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	for oldp, newp := range patternMap {
		// convert the string pattern to the slice of bytes
		newPattern := []byte(newp)
		// compile reg exp
		rgxp, _ := regexp.Compile(oldp)
		// Not so fast : returns a new copy of the slice (copying here)
		fileBuf = rgxp.ReplaceAll(fileBuf, newPattern)
	}
	return ioutil.WriteFile(dstFile, fileBuf, filePermissions)
}

func FindAndReplaceFd(fd *os.File, oldPattern, newPattern string) error {
	fbuf, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}
	fd.Seek(0, -1)
	fd.Truncate(0)
	expr, err := regexp.Compile(oldPattern)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(fbuf)
	for {
		line, err := buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if expr.MatchString(line) {
			line = expr.ReplaceAllString(line, newPattern)
		}
		if _, err := fd.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

func AppendToFd(fd *os.File, strToAdd, patToCheck string) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(fd); err != nil {
		return err
	}
	if patToCheck != "" {
		if ok, _ := regexp.Match(patToCheck, buf.Bytes()); ok {
			return nil
		}
	}
	if _, err := fd.WriteString(strToAdd); err != nil {
		return err
	}
	return nil
}

func WriteToFile(pathToFile string, opt int, data string) error {
	fd, err := os.OpenFile(pathToFile, opt, 0)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString(data)
	return err
}

// PatternFound
func PatternFound(file, pattern string) bool {
	fd, err := os.Open(file)
	if err != nil {
		return false
	}
	defer fd.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(fd); err != nil {
		return false
	}
	if ok, _ := regexp.Match(pattern, buf.Bytes()); ok {
		return true
	}
	return false
}

// CreateDirRecursively gets path to directories tree , permissions , UID
// and GID and a boolean value telling whether we should backup existing tree
// in case of necessity
// Creates the whole tree
// Returns error or nil
func CreateDirRecursively(dirToCreate string, permissions os.FileMode,
	owner, group int, backupOriginal bool) error {
	info, err := os.Stat(dirToCreate)
	if err == nil {
		if !info.IsDir() {
			return errors.New(dirToCreate + " exsists but it's not directory")
		}
		if backupOriginal {
			if err := os.Rename(dirToCreate, dirToCreate+".bkp"); err != nil {
				return err
			}
		}
	} else {
		if err := os.MkdirAll(dirToCreate, permissions); err != nil {
			return err
		}
	}
	if err = os.Chown(dirToCreate, owner, group); err != nil {
		return err
	}
	return nil
}

// RemoveIfExists cleans up appropriate stuff
// It gets a single path or multiple paths and
// boolean value telling if an error shall be thrown
// Returns error or nil
func RemoveIfExists(handleError bool, paths ...string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err = os.RemoveAll(path); err != nil {
				if handleError {
					return err
				}
			}
		}
	}
	return nil
}

///// Functions for copying/moving files and directories /////

func MyCopy(src, dst string, dstFilePermissions os.FileMode,
	dstFileOwner, dstFileGroup int, removeOriginal bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return CopyDir(src, dst)
	}
	return CopyFile(src, dst, dstFilePermissions, dstFileOwner, dstFileGroup, removeOriginal)
}

// Copy a directory(not portable function due to the "cp" usage)
func CopyDir(srcDir, dstDir string) error {
	if err := exec.Command("cp", "-a", srcDir, dstDir).Run(); err != nil {
		return errors.New(fmt.Sprintf("copying %s to %s : %s", srcDir, dstDir, err.Error()))
	}
	return nil
}

// copyFile copies a file.
// Returns amount of written bytes and error
func CopyFile(srcFile, dstFile string, dstFilePermissions os.FileMode,
	dstFileOwner, dstFileGroup int, removeOriginal bool) error {
	var srcfd, dstfd *os.File
	var err error
	// open source file
	if srcfd, err = os.Open(srcFile); err != nil {
		return err
	}
	// make sure descryptor is closed upon an error or exitting the function
	defer srcfd.Close()
	// get the file statistic
	srcFdStat, _ := srcfd.Stat()
	if dstFilePermissions == 0 {
		dstFilePermissions = srcFdStat.Mode().Perm()
	}
	finfo, err := os.Stat(dstFile)
	var newDstFile string
	if err == nil && finfo.IsDir() {
		newDstFile = filepath.Join(dstFile, filepath.Base(srcFile))
	} else {
		newDstFile = dstFile
	}
	// open destination file
	if dstfd, err = os.OpenFile(newDstFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		dstFilePermissions); err != nil {
		return err
	}
	// close destination fd upon an error or exitting the function
	defer dstfd.Close()
	// use io.Copy from standard library
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if removeOriginal {
		if err := os.Remove(srcFile); err != nil {
			return err
		}
	}
	return err
}

func CmdPipe(cmd1bin, cmd1args, cmd2bin, cmd2args string) error {
	cmd1 := exec.Command(cmd1bin, strings.Fields(cmd1args)...)
	cmd2 := exec.Command(cmd2bin, strings.Fields(cmd2args)...)
	cmd2.Stdin, _ = cmd1.StdoutPipe()
	cmd2.Stdout = os.Stdout
	if err := cmd2.Start(); err != nil {
		return err
	}
	if err := cmd1.Run(); err != nil {
		return err
	}
	cmd2.Wait()
	return nil
}
