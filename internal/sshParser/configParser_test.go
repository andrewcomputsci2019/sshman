package sshParser

import (
	"andrew/sshman/internal/sqlite"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kevinburke/ssh_config"
)

func TestReadConfig_SingleHost(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config")

	cfg := `
Host example.com
  User test
  Port 2222
`
	if err := os.WriteFile(file, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	hosts, err := ReadConfig(file)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}

	if hosts[0].Host != "example.com" {
		t.Fatalf("expected host example.com, got %q", hosts[0].Host)
	}

	if len(hosts[0].Options) != 2 {
		t.Fatalf("expected 2 hosts options but got %v", len(hosts[0].Options))
	}
}

func TestReadConfig_MultiHost(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config")

	cfg := `
Host example.com
  User test
  Port 2222

Host staging.example.com
  User deploy
  Hostname tester.local
  Port 2200
`
	if err := os.WriteFile(file, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	hosts, err := ReadConfig(file)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	if hosts[0].Host != "example.com" {
		t.Fatalf("expected host example.com, got %q", hosts[0].Host)
	}

	if hosts[1].Host != "staging.example.com" {
		t.Fatalf("expected host staging.example.com, got %q", hosts[1].Host)
	}

	if hosts[1].Options[1].Key != "Hostname" || hosts[1].Options[1].Value != "tester.local" {
		t.Fatalf("Parsed values are not correct")
	}

	if len(hosts[0].Options) != 2 {
		t.Fatalf("expected 2 hosts options but got %v", len(hosts[0].Options))
	}

	if len(hosts[1].Options) != 3 {
		t.Fatalf("expected 3 hosts options but got %v", len(hosts[1].Options))
	}
}

func TestSerializeHostToSshHost(t *testing.T) {
	host := sqlite.Host{
		Host:      "test.local",
		CreatedAt: time.Now(),
		Options: []sqlite.HostOptions{{
			Key:   "User",
			Value: "MyUser",
		}},
		Notes: "These should be at the bottom of the host",
	}
	hostStr, err := serializeHostToSshHost(&host)
	if err != nil {
		t.Fatalf("Failed to serialize ssh host object to ssh parsed form")
	}
	for _, node := range hostStr.Nodes {
		if comment, ok := node.(*ssh_config.Empty); ok {
			if comment.Comment != "These should be at the bottom of the host" {
				t.Fatalf("Comment should only be the string %s", "These should be at the bottom of the host")
			}
		}
	}
	t.Logf("converted host to string: %v", hostStr)
}

// todo add test for adding files into config file
func TestAddHostToFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "new_config")
	host := sqlite.Host{
		Host:      "test.local",
		CreatedAt: time.Now(),
		Options: []sqlite.HostOptions{{
			Key:   "User",
			Value: "MyUser",
		}},
		Notes: "These should be at the bottom of the host",
	}
	err := AddHostToFile(file, host)
	if err != nil {
		t.Fatalf("failed to add host into config file. Error %v", err)
	}
	// now we need to check if host exist in file
	hosts, err := ReadConfig(file)
	if err != nil {
		t.Fatalf("Failed to parse file after writing host to file. Error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("Should only be 1 host inside of the file")
	}
	if hosts[0].Host != "test.local" {
		t.Fatalf("Hostname is not correct: Expected %s but got %s", "test.local", hosts[0].Host)
	}
	if len(hosts[0].Options) != 1 {
		t.Fatalf("Hosts options should be of length 1 but was %v", len(hosts[0].Options))
	}
	if hosts[0].Options[0].Key != "User" || hosts[0].Options[0].Value != "MyUser" {
		t.Fatalf("Host option was not parsed correctly. %v,%v", hosts[0].Options[0].Key, hosts[0].Options[0].Value)
	}
	if hosts[0].Notes != "These should be at the bottom of the host" {
		t.Fatalf("Host Notes are not correct")
	}
}

func TestSerializeHostToFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "dump_config")
	host := sqlite.Host{
		Host:      "test.local",
		CreatedAt: time.Now(),
		Options: []sqlite.HostOptions{{
			Key:   "User",
			Value: "MyUser",
		}},
		Notes: "These should be at the bottom of the host",
	}
	host2 := sqlite.Host{
		Host:      "test2.local",
		CreatedAt: time.Now(),
		Options: []sqlite.HostOptions{{
			Key:   "User",
			Value: "MyUser2",
		}},
		Notes: "These should be at the bottom of host2",
	}
	err := SerializeHostToFile(file, []sqlite.Host{host, host2})
	if err != nil {
		t.Fatalf("Failed to write out ssh host objects to config file. Error %v", err)
	}
	hosts, err := ReadConfig(file)
	if err != nil {
		t.Fatalf("Failed to parse file after writing host to file. Error: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("Should only be 2 host inside of the file")
	}
	if hosts[0].Host != "test.local" {
		t.Fatalf("Hostname is not correct: Expected %s but got %s", "test.local", hosts[0].Host)
	}
	if len(hosts[0].Options) != 1 {
		t.Fatalf("Hosts options should be of length 1 but was %v", len(hosts[0].Options))
	}
	if hosts[0].Options[0].Key != "User" || hosts[0].Options[0].Value != "MyUser" {
		t.Fatalf("Host option was not parsed correctly. %v,%v", hosts[0].Options[0].Key, hosts[0].Options[0].Value)
	}
	if hosts[0].Notes != "These should be at the bottom of the host" {
		t.Fatalf("Host Notes are not correct. Notes %v", hosts[0].Notes)
	}

	if hosts[1].Host != "test2.local" {
		t.Fatalf("Hostname is not correct: Expected %s but got %s", "test.local", hosts[0].Host)
	}
	if len(hosts[1].Options) != 1 {
		t.Fatalf("Hosts options should be of length 1 but was %v", len(hosts[1].Options))
	}
	if hosts[1].Options[0].Key != "User" || hosts[1].Options[0].Value != "MyUser2" {
		t.Fatalf("Host option was not parsed correctly. %v,%v", hosts[1].Options[0].Key, hosts[1].Options[0].Value)
	}
	if hosts[1].Notes != "These should be at the bottom of host2" {
		t.Fatalf("Host Notes are not correct. Notes %v", hosts[1].Notes)
	}
}
