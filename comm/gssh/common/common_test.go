package common

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestRun(t *testing.T) {
	conf := &Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "user",
		Password:    "password",
		PrvtKeyFile: "",
	}
	c := NewClient(conf)
	scp, err := exec.LookPath("scp")
	if err != nil {
		t.Fatal(err)
	}
	cmd := fmt.Sprintf("%s -r %s@%s:/etc/hosts /tmp", scp, conf.User, conf.Host)

	if err := c.Run(cmd, 0); err != nil {
		t.Fatal(err)
	}
}
