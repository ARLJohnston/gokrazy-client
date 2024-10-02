package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient(t *testing.T) {
	clientName := "test"
	c, err := NewClient(clientName, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer c.Cleanup()

	if c.InstanceName != clientName {
		t.Errorf("Instance name was mangled in creation: wanted %q got %q", clientName, c.InstanceName)
	}
}

func TestDoRequest(t *testing.T) {
	fmt.Println("unimplemented")
}

func TestUpdate(t *testing.T) {
	fmt.Println("unimplemented")
}

func TestWifi(t *testing.T) {
	clientName := "test"
	c, err := NewClient(clientName, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer c.Cleanup()

	wifi := wifiConfiguration{
		SSID: "ssid",
		PSK:  "psk",
	}
	c.ApplyConfiguration(wifi)

	_, err = os.Stat(filepath.Join(c.ParentDir, c.InstanceName, "wifi.json"))
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf(err.Error())
	}
}

func TestCreateInstance(t *testing.T) {
	clientName := "test"
	c, err := NewClient(clientName, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer c.Cleanup()

	conf := gokrazyConfiguration{
		Hostname:      "test",
		Update:        update{HttpPassword: "password1"},
		Packages:      []string{"github.com/gokrazy/breakglass"},
		Config:        nil,
		SerialConsole: "disabled",
	}

	err = c.ApplyConfiguration(conf)
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = os.Stat(filepath.Join(c.ParentDir, c.InstanceName, "config.json"))
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf(err.Error())
	}
}

func TestEndToEnd(t *testing.T) {
	byt := []byte(`
{
    "Hostname": "test",
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

	res := gokrazyConfiguration{}
	if err := json.Unmarshal(byt, &res); err != nil {
		t.Errorf(err.Error())
	}

	c, err := NewClient("test", os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer c.Cleanup()

	err = c.ApplyConfiguration(res)
	if err != nil {
		t.Errorf(err.Error())
	}

	_, err = os.Stat(c.ParentDir)
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf(err.Error())
	}

	_, err = os.Stat(filepath.Join(c.ParentDir, c.InstanceName))
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf(err.Error())
	}

	file, err := os.Open(filepath.Join(c.ParentDir, c.InstanceName, "config.json"))
	if err != nil {
		t.Errorf(err.Error())
	}

	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf(err.Error())
	}

	gotConf := gokrazyConfiguration{}
	if err = json.Unmarshal(data, &gotConf); err != nil {
		t.Errorf(err.Error())
	}
}

func TestCleanup(t *testing.T) {
	c, err := NewClient("test", os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		t.Errorf(err.Error())
	}

	parentDir := c.ParentDir

	c.Cleanup()

	_, err = os.Stat(parentDir)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Unable to delete" + parentDir)
	}
}
