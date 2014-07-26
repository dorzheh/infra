package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Very simple and ugly implementation.

func Extract(fileToExtract, targetLocation string) error {
	var execCmd string
	var cmdArgs string
	var cmdArgDst string

	switch {
	case strings.HasSuffix(fileToExtract, ".tgz") ||
		strings.HasSuffix(fileToExtract, ".tar.gz"):
		execCmd = "tar"
		cmdArgs = "xfzp"
		cmdArgDst = "-C"
	case strings.HasSuffix(fileToExtract, ".tgx") ||
		strings.HasSuffix(fileToExtract, ".tar.gx"):
		execCmd = "tar"
		cmdArgs = "xfJp"
		cmdArgDst = "-C"
	case strings.HasSuffix(fileToExtract, ".bz2"):
		execCmd = "tar"
		cmdArgs = "xfjp"
		cmdArgDst = "-C"
	case strings.HasSuffix(fileToExtract, ".zip"):
		execCmd = "unzip"
		cmdArgDst = "-d"
	default:
		return errors.New("unknown compression format")
	}
	out, err := exec.Command(execCmd, cmdArgs, fileToExtract, cmdArgDst, targetLocation).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s [%s]", out, err)
	}
	return nil
}

func Archive(localExtractDir, targetArchive, file string, args ...string) error {
	if err := os.Chdir(localExtractDir); err != nil {
		return err
	}
	cmd := exec.Command("tar", "cfzp", targetArchive, file)
	if len(args) > 0 {
		cmd.Args = append(cmd.Args, args[0])
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s [%s]", out, err)
	}
	args = append(args, file)
	return RemoveIfExists(false, args...)
}
