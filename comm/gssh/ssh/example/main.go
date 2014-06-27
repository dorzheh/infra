package main

import (
	"os"

	"github.com/dorzheh/infra/comm/gssh/common"
	"github.com/dorzheh/infra/comm/gssh/ssh"
)

func main() {
	conf := &common.Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "user",
		Password:    "password",
		PrvtKeyFile: "",
	}
	c := ssh.NewClient(conf)
	if err := c.Run("ls /tmp"); err != nil {
		panic(err)
	}
	if err := c.Upload("/etc/hosts", "/tmp/hosts"); err != nil {
		panic(err)
	}
	os.Remove("/tmp/hosts")
}
