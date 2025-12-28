package sshUtils

import (
	"andrew/sshman/internal/config"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenRsaWithoutPassword(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyGenType := "rsa_"
	keyPair, err := genRSAKey(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair without password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenRsaWithPassword(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyGenType := "rsa_"
	keyPair, err := genRSAKey(hoststr, "password", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenRsaWithSanitizedName(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "fd00:1234:5678"
	keyGenType := "rsa_"
	keyPair, err := genRSAKey(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenECDSAWithoutPassword(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyGenType := "ecdsa_"
	keyPair, err := genECDSAKey(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate ecdsa key pair without password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenECDSAWithPassword(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyGenType := "ecdsa_"
	keyPair, err := genECDSAKey(hoststr, "password", cfg)
	if err != nil {
		t.Fatalf("Failed to generate ecdsa key pair without password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}
func TestGenECDSAWithSanitizeName(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "fd00:1234:5678"
	keyGenType := "ecdsa_"
	keyPair, err := genECDSAKey(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenED25519WithoutPassword(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyGenType := "ed25519_"
	keyPair, err := genED25519Key(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenED25519WithPassword(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyGenType := "ed25519_"
	keyPair, err := genED25519Key(hoststr, "password", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGenED25519WithSanitizeName(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "fd00:1234:5678"
	keyGenType := "ed25519_"
	keyPair, err := genED25519Key(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	t.Logf("Private Key %v, PubKey %v", keyPair.PrivateKey, keyPair.PubKey)
	privateKey := filepath.Base(keyPair.PrivateKey)
	if !hasCorrectPrefix(privateKey, keyGenType) {
		t.Fatalf("Key files did not have correct prefix of %v", keyGenType)
	}
	privateKey = strings.TrimPrefix(privateKey, keyGenType)
	if !usesSanitizedName(privateKey, hoststr) {
		t.Fatalf("Key pair did not sanitize name correctly")
	}
	privateKey = strings.TrimPrefix(privateKey, sanitizeName(hoststr))
	if !usesMarkerTag(privateKey) {
		t.Fatalf("Key did not use correct marker tag of 6 random hex chars and yyyymmdd timestamp")
	}
}

func TestGetKeyComment(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{}
	cfg.Ssh.KeyPath = dir
	hoststr := "example.com"
	keyPair, err := genED25519Key(hoststr, "", cfg)
	if err != nil {
		t.Fatalf("Failed to generate rsa key pair with password: Error %v", err)
	}
	comment, err := getKeyComment(keyPair.PrivateKey, sanitizeName(hoststr))
	if err != nil {
		t.Fatalf("Failed to extract a comment from given private key. Private key: %v , Error: %v", keyPair.PrivateKey, err)
	}
	sections := strings.Split(comment, ":")
	if len(sections) != 3 {
		t.Fatalf("Should be 3 parts to the comment")
	}
	if sections[0] != "ssh_man" {
		t.Fatalf("Section 0 should be the string ssh_man but was %v", sections[0])
	}
	if sections[1] != sanitizeName(hoststr) {
		t.Fatalf("Section 1 should be the sanitized host string")
	}
	if !usesMarkerTag(sections[2]) {
		t.Fatalf("Marker tag should be in an identical format as file")
	}
}

/*
	Utilities to check key strings are valid
*/

func hasCorrectPrefix(key, keyType string) bool {
	return strings.HasPrefix(key, keyType)
}

func usesSanitizedName(key, host string) bool {
	if key[0] == '_' {
		key = key[1:]
	}
	return strings.HasPrefix(key, sanitizeName(host))
}

func usesMarkerTag(key string) bool {
	if key[0] == '_' {
		key = key[1:]
	}
	parts := strings.Split(key, "_")
	if len(parts) < 2 {
		return false
	}
	hexPart := parts[0]
	timePart := parts[1]
	_, err := hex.DecodeString(hexPart)
	if err != nil {
		return false
	}
	return time.Now().Format("20060102") == timePart
}
