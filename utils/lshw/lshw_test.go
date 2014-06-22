package lshw

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestWriteToStdout(t *testing.T) {
	config := &Config{
		Class:  []Class{Network, Processor},
		Format: FormatJSON,
	}
	c, err := New("", config)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.WriteToStdout(); err != nil {
		t.Fatal(err)
	}
}

func TestWriteToFile(t *testing.T) {
	config := &Config{
		Class:  []Class{Network, Processor},
		Format: FormatJSON,
	}
	c, err := New("", config)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := ioutil.TempFile("/tmp", "lshw.test")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		fd.Close()
		os.Remove(fd.Name())
	}()
	if err := c.WriteToFile(fd.Name()); err != nil {
		t.Fatal(err)
	}
}
