package ssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"github.com/dorzheh/infra/comm/common"
)

type SshConn struct {
	Client *ssh.Client
}

func NewSshConn(c *common.Config) (conn *SshConn, err error) {
	auth := []ssh.AuthMethod{
		ssh.Password(c.Password),
	}
	var key ssh.Signer
	if _, err = os.Stat(c.PrvtKeyFile); err == nil {
		key, err = getKey(c.PrvtKeyFile)
		if err != nil {
			return
		}
		auth = append(auth, ssh.PublicKeys(key))
	}

	clientConfig := &ssh.ClientConfig{
		User: c.User,
		Auth: auth,
	}
	var client *ssh.Client
	// connect to remote host
	client, err = ssh.Dial("tcp", c.Host+":"+c.Port, clientConfig)
	if err != nil {
		return
	}
	conn = &SshConn{Client: client}
	return
}

func (c *SshConn) ConnClose() {
	c.Client.Close()
}

func (c *SshConn) Upload(src, dst string) error {
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}
	fd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fd.Close()

	session, err := c.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C0644", stat.Size(), filepath.Base(dst))
		io.Copy(w, fd)
		fmt.Fprint(w, "\x00")
	}()
	cmd := "scp  -qrt " + filepath.Dir(dst)
	return session.Run(cmd)
}

// Returns output from descryptors 1(result output) and 2( error output) and  error/nil
func (c *SshConn) Run(cmd string) (string, string, error) {
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer

	session, err := c.Client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	session.Stdout = &outputBuffer
	session.Stderr = &errorBuffer
	if err := session.Run(cmd); err != nil {
		return "", errorBuffer.String(), err
	}
	return outputBuffer.String(), "", nil
}

func getKey(pathToKeyFile string) (key ssh.Signer, err error) {
	var buf []byte
	buf, err = ioutil.ReadFile(pathToKeyFile)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}
