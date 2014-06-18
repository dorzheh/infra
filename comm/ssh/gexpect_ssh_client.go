package ssh

import (
	"os/exec"
	"regexp"
	"time"

	"github.com/dorzheh/gexpect"
)

type gSshClient struct {
	host     string
	user     string
	password string
}

var ask_known_hosts = regexp.MustCompile(`Are you sure you want to continue connecting (yes/no)?`)
var ask_password = regexp.MustCompile(`password:`)

func NewGsshClient(host, user, password string) *gSshClient {
	return &gSshClient{host, user, password}
}

func (c *gSshClient) Download(remote, local string) error {
	scp, err := exec.LookPath("scp")
	if err != nil {
		return err
	}
	child, err := gexpect.NewSubProcess(scp, c.user+"@"+c.host+":"+remote, local)
	if err != nil {
		return err
	}
	return c.expect(child, 0, 0)
}

func (c *gSshClient) Upload(local, remote string) error {
	scp, err := exec.LookPath("scp")
	if err != nil {
		return err
	}
	child, err := gexpect.NewSubProcess(scp, local, c.user+"@"+c.host+":"+remote)
	if err != nil {
		return err
	}
	return c.expect(child, 0, 0)
}

func (c *gSshClient) Run(cmd string) error {
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		return err
	}
	child, err := gexpect.NewSubProcess(ssh, c.user+"@"+c.host, cmd)
	if err != nil {
		return err
	}
	return c.expect(child, 0, 0)
}

func (c *gSshClient) expect(child *gexpect.SubProcess, expectTimeoutSec, interactTimeoutSec time.Duration) error {
	defer child.Close()
	if err := child.Start(); err != nil {
		return err
	}
	if idx, _ := child.ExpectTimeout(expectTimeoutSec*time.Second, ask_known_hosts, ask_password); idx >= 0 {
		if idx == 0 {
			child.SendLine("yes")
			if idx, _ := child.ExpectTimeout(expectTimeoutSec*time.Second, ask_password); idx >= 0 {
				child.SendLine(c.password)
			}
		} else if idx == 1 {
			child.SendLine(c.password)
		}
	}
	return child.InteractTimeout(interactTimeoutSec * time.Second)
}
