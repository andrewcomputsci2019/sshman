package sshParser

import (
	"os"
	"path/filepath"
	"testing"
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
