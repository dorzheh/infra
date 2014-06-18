package utils

import (
	"io/ioutil"
	"os/exec"
	"fmt"
	"bytes"
	"regexp"
	"sort"
	"path/filepath"
)

// RunSorted executes the scripts in a numeric order
// The scripts must contain the following prefix : [0-9]+_
// Example: 02_clean
//returns error or nil
func RunSorted(pathToScriptsDir string) error {
        var scriptsSlice []string
        dir, _ := ioutil.ReadDir(pathToScriptsDir)
        for _, f := range dir {
                if !f.IsDir() {
                        if found, _ := regexp.MatchString("[0-9]+_", f.Name()); found {
                                scriptsSlice = append(scriptsSlice, filepath.Join(pathToScriptsDir, f.Name()))
                        }
                }
        }
        sort.Strings(scriptsSlice)
        var stderr bytes.Buffer
        for _, file := range scriptsSlice {
                cmd := exec.Command(file)
                cmd.Stderr = &stderr
                if err := cmd.Run(); err != nil {
                        return fmt.Errorf("%s [%s]", stderr.String(), err)
                }
        }
        return nil
}

