package ssh

import (
	"testing"

	"github.com/dorzheh/infra/comm/gssh/common"
)

func TestRun(t *testing.T) {
	conf := &common.Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "user",
		Password:    "password",
		PrvtKeyFile: "",
	}
	c := NewClient(conf)
	if err := c.Run("ls /tmp"); err != nil {
		t.Fatal(err)
	}
	if err := c.Upload("/etc/hosts", "/tmp/hosts"); err != nil {
		t.Fatal(err)
	}
}
