package main

import (
	"fmt"

	"github.com/dorzheh/infra/comm/gssh/common"
	"github.com/dorzheh/infra/comm/gssh/sshfs"
)

func main() {
	conf := &common.Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "user",
		Password:    "password",
		PrvtKeyFile: "",
	}
	sshfsConf := &sshfs.Config{
		Common:      conf,
		SshfsPath:   "",
		FusrmntPath: "",
	}
	c, err := sshfs.NewClient(sshfsConf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Attaching...")
	if err := c.Attach("/media", "/mnt"); err != nil {
		panic(err)
	}

	fmt.Printf("Detaching...")
	if err := c.Detach("/mnt"); err != nil {
		panic(err)
	}
	fmt.Printf("\n")
}
