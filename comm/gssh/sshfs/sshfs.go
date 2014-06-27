package sshfs

import (
	"fmt"
	"os/exec"

	"github.com/dorzheh/infra/comm/gssh/common"
)

type Config struct {
	Common      *common.Config
	SshfsPath   string
	FusrmntPath string
}

type Client struct {
	*Config
	*common.Client
}

func NewClient(config *Config) (*Client, error) {
	var err error

	if config.SshfsPath == "" {
		config.SshfsPath, err = exec.LookPath("sshfs")
		if err != nil {
			return nil, err
		}
	}
	if config.FusrmntPath == "" {
		config.FusrmntPath, err = exec.LookPath("fusermount")
		if err != nil {
			return nil, err
		}
	}
	return &Client{config, common.NewClient(config.Common)}, nil
}

func (c *Client) Attach(remoteShare, localMount string) error {
	cmd := fmt.Sprintf("%s %s@%s:%s %s -o port=%s,idmap=user,compression=no,nonempty,Ciphers=arcfour",
		c.SshfsPath, c.User, c.Host, remoteShare, localMount, c.Port)
	if c.PrvtKeyFile != "" {
		cmd += ",IdentityFile=" + c.PrvtKeyFile
	}
	return c.Run(cmd, 0)
}

func (c *Client) Detach(localMount string) error {
	return exec.Command(c.FusrmntPath, "-u", localMount).Run()
}
