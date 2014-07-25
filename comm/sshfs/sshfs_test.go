package sshfs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dorzheh/infra/comm/common"
)

func TestRun(t *testing.T) {
	conf := &common.Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "username",
		Password:    "password",
		PrvtKeyFile: "",
	}
	sshfsConf := &Config{
		Common:      conf,
		SshfsPath:   "",
		FusrmntPath: "",
	}

	ldir, err := ioutil.TempDir("", "sshfstest-")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := os.RemoveAll(ldir); err != nil {
			t.Error(err)
		}
	}()

	rdir, err := ioutil.TempDir("", "sshfstest-")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := os.RemoveAll(rdir); err != nil {
			t.Error(err)
		}
	}()

	c, err := NewClient(sshfsConf)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Attach(rdir, ldir); err != nil {
		t.Fatal(err)
	}
	if err := c.Detach(ldir); err != nil {
		t.Fatal(err)
	}
}
