package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/docopt/docopt-go"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultPort       = 22
	DefaultConfigPath = "/usr/local/etc/certmgr/certmgr.conf"
	DefaultPrivateKey = "/usr/local/etc/certmgr/id_rsa"
	DefaultStore      = "/var/certmgr"
)

type Config struct {
	Store      string
	Hostname   string
	User       string
	Port       int
	PrivateKey string

	CertUser   string
	ServerCert string
	ServerKey  string
	Fullchain  string

	ReloadCommand   string
	ReloadArguments []string
}

func (c Config) lastmodified() string {
	return path.Join(c.Store, "lastmodified")
}

type Client struct {
	*sftp.Client
}

func (c *Client) ReadFile(p string) ([]byte, error) {
	remoteFile, err := c.Open(p)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(remoteFile)
}

func (c *Client) DownloadFile(source, dest string, conf *Config) error {
	if verbose {
		log.Printf("%s ==> %s\n", source, dest)
	}

	data, err := c.ReadFile(source)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dest, data, 0640); err != nil {
		return err
	}

	if conf != nil {
		uid, err := conf.UID()
		if err != nil {
			return err
		}
		gid, err := conf.GID()
		if err != nil {
			return err
		}
		if err := Chown(dest, uid, gid); err != nil {
			return err
		}
	}

	return nil
}

func newSSHClient(conf *Config) (*ssh.Client, error) {

	pemkey, err := ioutil.ReadFile(conf.PrivateKey)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey([]byte(pemkey))
	if err != nil {
		return nil, fmt.Errorf("signer key %w", err)
	}

	config := &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if verbose {
		log.Printf("sftp -i %s -P %d %s@%s\n", conf.PrivateKey, conf.Port, conf.User, conf.Hostname)
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conf.Hostname, conf.Port), config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %w", err)
	}

	return client, nil
}

func newClient(conf *Config) (*Client, error) {
	sshClient, err := newSSHClient(conf)
	if err != nil {
		return nil, err
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

func loadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := toml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	if conf.User == "" {
		return nil, errors.New("missing: User")
	}

	if conf.Hostname == "" {
		return nil, errors.New("missing: Hostname")
	}

	if conf.Port == 0 {
		conf.Port = DefaultPort
	}

	if conf.PrivateKey == "" {
		conf.PrivateKey = DefaultPrivateKey
	}

	if conf.Store == "" {
		conf.Store = DefaultStore
	}

	if conf.Fullchain == "" {
		conf.Fullchain = path.Join(conf.Store, "fullchain.crt")
	}

	if conf.ServerCert == "" {
		conf.ServerCert = path.Join(conf.Store, "server.crt")
	}

	if conf.ServerKey == "" {
		conf.ServerKey = path.Join(conf.Store, "server.key")
	}

	return &conf, nil
}

func compareLastmodified(client *Client, conf *Config) (bool, error) {
	localDate, err := ioutil.ReadFile(conf.lastmodified())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	remoteDate, err := client.ReadFile("lastmodified")
	if err != nil {
		return false, err
	}

	return bytes.Compare(remoteDate, localDate) != 0, nil
}

func renew(client *Client, conf *Config) error {
	if verbose {
		log.Println("renewing")
	}

	if err := client.DownloadFile("lastmodified", conf.lastmodified(), nil); err != nil {
		return err
	}

	if err := client.DownloadFile("server.pem", conf.ServerCert, conf); err != nil {
		return err
	}

	if err := client.DownloadFile("server.key", conf.ServerKey, conf); err != nil {
		return err
	}

	if err := client.DownloadFile("fullchain.pem", conf.Fullchain, conf); err != nil {
		return err
	}

	if conf.ReloadCommand != "" {
		fmt.Println("reloading")
		out, err := exec.Command(conf.ReloadCommand, conf.ReloadArguments...).Output()
		if err != nil {
			return err
		}
		fmt.Printf("=======\n%s\n======", out)
	}

	return nil
}

var verbose bool

func main() {

	usage := `Certmgr.

Usage:
  certmgr [--verbose] [--force] [--config=<config>]
  certmgr -h | --help
  certmgr --version

Options:
  -h --help                       Show this screen.
  --version                       Show version.
  -v, --verbose					  Verbose
  -f, --force                     Force
  -c <config>, --config=<config>  Path to config [default: /usr/local/etc/certmgr/certmgr.conf].
`
	arguments, _ := docopt.ParseDoc(usage)
	confPath, err := arguments.String("--config")
	if err != nil {
		log.Fatal(err)
	}

	force, err := arguments.Bool("--force")
	if err != nil {
		log.Fatal(err)
	}

	verbose, err = arguments.Bool("--verbose")
	if err != nil {
		log.Fatal(err)
	}

	conf, err := loadConfig(confPath)
	if err != nil {
		log.Panic(err)
	}

	client, err := newClient(conf)
	if err != nil {
		log.Panic(err)
	}

	shouldRenew, err := compareLastmodified(client, conf)
	if err != nil {
		log.Panic(err)
	}

	if !shouldRenew && !force {
		if verbose {
			log.Println("nothing to do")
		}
		return
	}

	err = renew(client, conf)
	if err != nil {
		log.Panic(err)
	}

}
