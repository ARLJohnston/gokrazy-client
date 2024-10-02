package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	gok "github.com/gokrazy/tools/gok"
)

type Client struct {
	InstanceName string
	ParentDir    string
	stdin        io.Reader
	stdout       io.Writer
	stderr       io.Writer
	conf         configuration
}

func NewClient(instanceName string, stdin io.Reader, stdout, stderr io.Writer) (*Client, error) {
	parentDir, err := os.MkdirTemp("", instanceName)
	if err != nil {
		return nil, err
	}

	c := Client{InstanceName: instanceName, ParentDir: parentDir}

	if stdin != nil {
		c.stdin = stdin
	}

	if stdout != nil {
		c.stdout = stdout
	}

	if stderr != nil {
		c.stderr = stderr
	}

	return &c, nil
}

func (c *Client) DoRequest(args []string) error {
	ctx := gok.Context{
		Stdin:  c.stdin,
		Stdout: c.stdout,
		Stderr: c.stderr,
		Args:   append(args, "--parent_dir", c.ParentDir),
	}
	err := ctx.Execute(context.Background())

	return err
}

func (c *Client) Update() error {
	return c.DoRequest([]string{"update"})
}

func (c *Client) CreateInstance(conf configuration) error {
	err := os.Mkdir(filepath.Join(c.ParentDir, c.InstanceName), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(c.ParentDir, c.InstanceName, "config.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	byt, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return err
	}

	f.Write(byt)

	return nil
}

func (c *Client) Cleanup() error {
	return os.RemoveAll(c.ParentDir)
}

type update struct {
	HttpPassword string `json:"HTTPPassword"`
}

type configuration struct {
	Hostname      string                            `json:"Hostname"`
	Update        update                            `json:"Update"`
	Packages      []string                          `json:"Packages"`
	Config        map[string]map[string]interface{} `json:"PackageConfig"`
	SerialConsole string                            `json:"SerialConsole"`
}

func main() {
	byt := []byte(`
{
    "Hostname": "example",
    "Update": {
        "HTTPPassword": "password1"
    },
    "Packages": [
        "github.com/gokrazy/podman"
    ],
    "PackageConfig": {
        "github.com/gokrazy/breakglass": {
            "CommandLineFlags": [
                "-authorized_keys=/etc/breakglass.authorized_keys"
            ],
            "ExtraFilePaths": {
                "/etc/breakglass.authorized_keys": "/some/path"
            }
        }
    },
    "SerialConsole": "disabled"
}
`)
	res := configuration{}
	if err := json.Unmarshal(byt, &res); err != nil {
		panic(err)
	}

	c, err := NewClient("hello", os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		panic(err)
	}

	err = c.CreateInstance(res)
	if err != nil {
		panic(err)
	}

	fmt.Println(c.ParentDir)

	timer := time.NewTimer(10 * time.Second)
	<-timer.C

	err = c.Cleanup()
	if err != nil {
		panic(err)
	}
}
