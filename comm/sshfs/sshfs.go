package sshfs

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dorzheh/infra/comm/common"
)

type Config struct {
	Common      *common.Config
	SshpassPath string
	SshfsPath   string
	FusrmntPath string
}

type Client struct {
	*Config
}

func NewClient(config *Config) (*Client, error) {
	var err error

	if config.SshpassPath == "" {
		config.SshpassPath, err = exec.LookPath("sshpass")
		if err != nil {
			return nil, err
		}
	}
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
	return &Client{config}, nil
}

func (c *Client) Attach(remoteShare, localMount string) error {
	var cmd *exec.Cmd
	cmdStr := fmt.Sprintf("%s@%s:%s %s -o port=%s,idmap=user,compression=no,nonempty,Ciphers=arcfour",
		c.Common.User, c.Common.Host, remoteShare, localMount, c.Common.Port)
	if c.Common.PrvtKeyFile == "" {
		cmd = exec.Command(c.SshpassPath, "-p", c.Common.Password, c.SshfsPath)
	} else {
		cmd = exec.Command(c.SshfsPath)
		cmdStr += ",IdentityFile=" + c.Common.PrvtKeyFile
	}
	cmd.Args = append(cmd.Args, strings.Fields(cmdStr)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s [%s]", out, err)
	}
	return nil
}

func (c *Client) Detach(localMount string) error {
	if out, err := exec.Command(c.FusrmntPath, "-u", localMount).CombinedOutput(); err != nil {
		return fmt.Errorf("%s [%s]", out, err)
	}
	return nil
}
