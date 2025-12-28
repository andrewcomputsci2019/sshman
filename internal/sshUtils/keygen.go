package sshUtils

import (
	"andrew/sshman/internal/config"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/adrg/xdg"
)

const (
	randomTagSize = 3 // 6 hex random characters
)

type KeyPair struct {
	PubKey     string // PubKey path
	PrivateKey string // Private Key Path
}

// todo add code here to automatically gen keys etc

// GenKey takes host that the key is generated for a keyType find under config package RSA ECDSA ED25519
// returns the generated key path (should be relative if cfg uses default path), returns non nill error if incorrect key type
// or process error during the generation of the key. Note this function is blocking and should be run outside the ui thread
func GenKey(host, keyType, password string, cfg config.Config) (KeyPair, error) {

	switch keyType {
	case config.RSA:
		return genRSAKey(host, password, cfg)
	case config.ECDSA:
		return genECDSAKey(host, password, cfg)
	case config.ED25519:
		return genED25519Key(host, password, cfg)
	default:
		return KeyPair{}, errors.New("Key type given is not recognized")
	}
}

// genRSA key function
func genRSAKey(host string, password string, cfg config.Config) (KeyPair, error) {
	safeHost := sanitizeName(host)
	// if slice not provided all keys valid
	if len(cfg.Ssh.AcceptableKeyGenAlgorithms) > 0 {
		if !slices.ContainsFunc(cfg.Ssh.AcceptableKeyGenAlgorithms, func(e string) bool {
			return strings.EqualFold(e, config.RSA)
		}) {
			return KeyPair{}, errors.New("RSA key generation is disabled in config")
		}
	}
	// call ssh key gen function add host string as comment and name key after host
	var keyGenPath string
	if cfg.Ssh.KeyPath == "" {
		keyGenPath = filepath.Join(xdg.ConfigHome, config.DefaultAppStorePath, config.KeyStoreDir)
	} else {
		keyGenPath = cfg.Ssh.KeyPath
	}
	marker, err := newKeyMarker()
	if err != nil {
		return KeyPair{}, err
	}
	comment := config.AppName + ":" + safeHost + ":" + marker
	filename := "rsa_" + safeHost + "_" + marker
	full_path := filepath.Join(keyGenPath, filename)
	if exist, err := doesFileExist(full_path); exist || err != nil {
		if err != nil {
			return KeyPair{}, err
		}
		full_path += "_1"
	} else if exist, err := doesFileExist(full_path + ".pub"); exist || err != nil {
		if err != nil {
			return KeyPair{}, err
		}
		full_path += "_1"
	}
	args := []string{
		"-t", "rsa",
		"-f", full_path,
		"-C", comment,
		"-N", password,
		"-b", "4096",
	}
	exe := exec.Command("ssh-keygen", args...)
	err = exe.Run()
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{
		PubKey:     full_path + ".pub",
		PrivateKey: full_path,
	}, nil
}

// genECDSAKey function
func genECDSAKey(host string, password string, cfg config.Config) (KeyPair, error) {
	safeHost := sanitizeName(host)
	// if slice not provided all keys valid
	if len(cfg.Ssh.AcceptableKeyGenAlgorithms) > 0 {
		if !slices.ContainsFunc(cfg.Ssh.AcceptableKeyGenAlgorithms, func(e string) bool {
			return strings.EqualFold(e, config.ECDSA)
		}) {
			return KeyPair{}, errors.New("ECDSA key generation is disabled in config")
		}
	}
	// call ssh key gen function add host string as comment and name key after host
	var keyGenPath string
	if cfg.Ssh.KeyPath == "" {
		keyGenPath = filepath.Join(xdg.ConfigHome, config.DefaultAppStorePath, config.KeyStoreDir)
	} else {
		keyGenPath = cfg.Ssh.KeyPath
	}
	marker, err := newKeyMarker()
	if err != nil {
		return KeyPair{}, err
	}
	comment := config.AppName + ":" + safeHost + ":" + marker
	filename := "ecdsa_" + safeHost + "_" + marker
	full_path := filepath.Join(keyGenPath, filename)
	if exist, err := doesFileExist(full_path); exist || err != nil {
		if err != nil {
			return KeyPair{}, err
		}
		full_path += "_1"
	} else if exist, err := doesFileExist(full_path + ".pub"); exist || err != nil {
		if err != nil {
			return KeyPair{}, err
		}
		full_path += "_1"
	}
	args := []string{
		"-t", "ecdsa",
		"-f", full_path,
		"-C", comment,
		"-N", password,
		"-b", "521",
	}
	exe := exec.Command("ssh-keygen", args...)
	err = exe.Run()
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{
		PubKey:     full_path + ".pub",
		PrivateKey: full_path,
	}, nil
}

// genED25519Key function
func genED25519Key(host string, password string, cfg config.Config) (KeyPair, error) {
	safeHost := sanitizeName(host)
	// if slice not provided all keys valid
	if len(cfg.Ssh.AcceptableKeyGenAlgorithms) > 0 {
		if !slices.ContainsFunc(cfg.Ssh.AcceptableKeyGenAlgorithms, func(e string) bool {
			return strings.EqualFold(e, config.ED25519)
		}) {
			return KeyPair{}, errors.New("ED25519 key generation is disabled in config")
		}
	}
	// call ssh key gen function add host string as comment and name key after host
	var keyGenPath string
	if cfg.Ssh.KeyPath == "" {
		keyGenPath = filepath.Join(xdg.ConfigHome, config.DefaultAppStorePath, config.KeyStoreDir)
	} else {
		keyGenPath = cfg.Ssh.KeyPath
	}
	marker, err := newKeyMarker()
	if err != nil {
		return KeyPair{}, err
	}
	comment := config.AppName + ":" + safeHost + ":" + marker
	filename := "ed25519_" + safeHost + "_" + marker
	full_path := filepath.Join(keyGenPath, filename)
	if exist, err := doesFileExist(full_path); exist || err != nil {
		if err != nil {
			return KeyPair{}, err
		}
		full_path += "_1"
	} else if exist, err := doesFileExist(full_path + ".pub"); exist || err != nil {
		if err != nil {
			return KeyPair{}, err
		}
		full_path += "_1"
	}
	args := []string{
		"-t", "ed25519",
		"-f", full_path,
		"-C", comment,
		"-N", password,
	}
	exe := exec.Command("ssh-keygen", args...)
	err = exe.Run()
	if err != nil {
		return KeyPair{}, err
	}
	return KeyPair{
		PubKey:     full_path + ".pub",
		PrivateKey: full_path,
	}, nil
}

func doesFileExist(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func sanitizeName(host string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		switch r {
		case '-', '_', '.':
			return r
		default:
			return '_' // replace char with legal char
		}
	}, host)
}

func newKeyMarker() (string, error) {
	var b [randomTagSize]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	hexPart := hex.EncodeToString(b[:])
	datePart := time.Now().Format("20060102")
	return hexPart + "_" + datePart, nil
}

func getKeyComment(keyPath string, cleanedHost string) (string, error) {
	// todo write code to take key string and parse it to get key comment
	filename := filepath.Base(keyPath)
	if strings.HasSuffix(filename, ".pub") {
		return "", fmt.Errorf("expected private key, got public key")
	}
	idx := strings.IndexByte(filename, '_')
	if idx == -1 {
		return "", fmt.Errorf("invalid key filename: %q", filename)
	}
	keyPrefix := filename[:idx]
	keyWithOutGenType := strings.TrimPrefix(filename, keyPrefix+"_")
	if !strings.HasPrefix(keyWithOutGenType, cleanedHost) {
		return "", fmt.Errorf("Key is not in correct format")
	}
	tag := strings.TrimPrefix(keyWithOutGenType, cleanedHost+"_")
	return strings.Join([]string{config.AppName, cleanedHost, tag}, ":"), nil
}

// CopyKey returns the exce cmd necessary to copy the public key to the ssh host
//
// keyToCopy is the path to the public key to copy
//
// host is the host target
//
// cfg is the config of the program
// options are options the user passed in when calling the program, note each of these need to options need to be
// prefixed with -o, ie "-o Port=22" would be a valid string
func CopyKey(keyToCopy, host string, cfg config.Config, options ...string) *exec.Cmd {
	configFilePath := cfg.GetSshConfigFilePath()
	args := []string{
		"-F", configFilePath,
		"i", keyToCopy,
	}
	args = append(args, options...)
	args = append(args, host)
	return exec.Command("ssh-copy-id", args...)
}

func generateShellScriptToRemoveOldKey(comment string) string {
	// script first checks for posix compliance
	// then inline the tag as a sh var
	// check for ssh key file existence
	// backup current key file in case an issue occurs
	// greps with invent match and returns all lines that do not have the comment from the key
	// overwrites the key file with the new one and chmods it to be user owned
	// then returns
	script := strings.TrimSpace(`command -v sh >/dev/null 2>&1 || exit 100
	set -e
	TAG="` + comment + `"
	AUTH_KEYS="$HOME/.ssh/authorized_keys"
	[ -f "$AUTH_KEYS" ] || exit 2
	cp -p "$AUTH_KEYS" "$AUTH_KEYS.bak"
	grep -vF "$TAG" "$AUTH_KEYS" > "$AUTH_KEYS.tmp"
	mv "$AUTH_KEYS.tmp" "$AUTH_KEYS"
	chmod 600 "$AUTH_KEYS"`)
	return script
}

func RemoveOldKeyFromRemoteServer(keyToRemove, host string, cfg config.Config, options ...string) (*exec.Cmd, error) {
	sanitizedHost := sanitizeName(host)
	comment, err := getKeyComment(keyToRemove, sanitizedHost)
	if err != nil {
		return nil, fmt.Errorf("Failed to get comment from key")
	}
	scriptString := generateShellScriptToRemoveOldKey(comment)
	configFilePath := cfg.GetSshConfigFilePath()
	args := []string{
		"-F", configFilePath,
	}
	args = append(args, options...)
	args = append(args, host)
	args = append(args, "sh", "-c", scriptString)

	return exec.Command("ssh", args...), nil

}
