package common

import (
	"regexp"
	"strings"
	"time"

	"github.com/dorzheh/gexpect"
)

var ask_known_hosts = regexp.MustCompile(`Are you sure you want to continue connecting (yes/no)?`)
var ask_password = regexp.MustCompile(`password:`)

type Config struct {
	Host        string
	Port        string
	User        string
	Password    string
	PrvtKeyFile string
}

type Client struct {
	*Config
}

func NewClient(config *Config) *Client {
	return &Client{config}
}

func (c *Client) Run(cmd string, timeout time.Duration) error {
	cmdSlice := strings.Split(cmd, " ")
	child, err := gexpect.NewSubProcess(cmdSlice[0], cmdSlice[1:]...)
	if err != nil {
		return err
	}
	return c.expect(child, 0, 0)
}

func (c *Client) expect(child *gexpect.SubProcess, expectTimeoutSec, interactTimeoutSec time.Duration) error {
	defer child.Close()
	if err := child.Start(); err != nil {
		return err
	}
	if idx, _ := child.ExpectTimeout(expectTimeoutSec*time.Second, ask_known_hosts, ask_password); idx >= 0 {
		if idx == 0 {
			child.SendLine("yes")
			if idx, _ := child.ExpectTimeout(expectTimeoutSec*time.Second, ask_password); idx >= 0 {
				child.SendLine(c.Password)
			}
		} else if idx == 1 {
			child.SendLine(c.Password)
		}
	}
	return child.InteractTimeout(interactTimeoutSec * time.Second)
}
