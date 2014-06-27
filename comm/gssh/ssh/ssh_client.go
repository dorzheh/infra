package ssh

import (
	"fmt"
	"os/exec"

	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/gssh/common"
)

type Client struct {
	common *common.Client
}

func NewClient(config *ssh.Config) *Client {
	return &Client{common.NewClient(config)}
}

func (c *Client) Download(remote, local string) error {
	scp, err := exec.LookPath("scp")
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s -P %s -r %s@%s:%s %s", scp, c.common.Port, c.common.User, c.common.Host, remote, local)
	return c.common.Run(cmd, 0)
}

func (c *Client) Upload(local, remote string) error {
	scp, err := exec.LookPath("scp")
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s -P %s -r  %s %s@%s:%s", scp, c.common.Port, local, c.common.User, c.common.Host, remote)
	return c.common.Run(cmd, 100)
}

func (c *Client) Run(cmd string) error {
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}
	cmnd := fmt.Sprintf("%s -p %s %s@%s %s", ssh, c.common.Port, c.common.User, c.common.Host, cmd)
	return c.common.Run(cmnd, 0)
}
