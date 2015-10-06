package ssh

import (
	ssh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"io/ioutil"
	"os/user"
	//"github.com/ngaut/log"
)

type Config struct {
	PrivaryKey string //ssh privary key
	RemoteHost string // remote host
	RemotePort int    // remote port default 22
	RemoteUser string // remote login username
}

type SSHLinker struct {
	client *ssh.Client
}

func NewSSHLinker(conf Config) (*SSHLinker, error) {
	// init conf
	if conf.RemotePort == 0 {
		conf.RemotePort = 22
	}

	signer, err := ssh.ParsePrivateKey([]byte(conf.PrivaryKey))
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: conf.RemoteUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	c, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conf.RemoteHost, conf.RemotePort), config)
	if err != nil {
		return nil, err
	}

	return &SSHLinker{c}, nil
}

func (this *SSHLinker) GetClient() *ssh.Client {
	return this.client
}

func (this *SSHLinker) Close() error {
	return this.client.Close()
}

func GetPrivateKey() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}

	data, err := ioutil.ReadFile(u.HomeDir + "/.ssh/id_rsa")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(data)
}
