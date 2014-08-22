package sshfs

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/utils/ioutils"
)

type Config struct {
	Common      *common.Config
	SshfsPath   string
	FusrmntPath string
}

type Client struct {
	*Config
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
	return &Client{config}, nil
}

func (c *Client) Attach(remoteShare, localMount string) error {
	cmdStr := fmt.Sprintf(" -t fuse %s#%s@%s:%s %s -o port=%s,idmap=user,compression=no,allow_root,nonempty,Ciphers=arcfour,reconnect,transform_symlinks,StrictHostKeyChecking=no",
		c.SshfsPath, c.Common.User, c.Common.Host, remoteShare, localMount, c.Common.Port)
	if c.Common.PrvtKeyFile == "" {
		cmdStr += ",password_stdin"
		if err := ioutils.CmdPipe("echo", c.Common.Password, "mount", cmdStr); err != nil {
			return err
		}
	} else {
		cmd := exec.Command("mount")
		cmdStr += ",IdentityFile=" + c.Common.PrvtKeyFile
		cmd.Args = append(cmd.Args, strings.Fields(cmdStr)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s [%s]", out, err)
		}
	}
	return nil
}

func (c *Client) Detach(localMount string) error {
	if out, err := exec.Command(c.FusrmntPath, "-u", localMount).CombinedOutput(); err != nil {
		return fmt.Errorf("%s [%s]", out, err)
	}
	return nil
}
