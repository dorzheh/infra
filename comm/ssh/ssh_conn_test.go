package ssh

import (
	"testing"
)

func TestRun(t *testing.T) {
	conf := &Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "test",
		Passwd:      "test",
		PrvtKeyFile: "",
	}
	c, err := NewSshConn(conf)
	if err != nil {
		t.Fatal(err)
	}
	defer c.ConnClose()
	if _, _, err := c.Run("ls /tmp"); err != nil {
		t.Fatal(err)
	}
	if err := c.Upload("/etc/hosts", "/tmp/hosts"); err != nil {
		t.Fatal(err)
	}
}
