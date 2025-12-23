package sshUtils

import (
	"andrew/sshman/internal/config"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/adrg/xdg"
	"github.com/google/uuid"
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
	random_uuid, err := uuid.NewV7()
	if err != nil {
		slog.Warn("Failed to generate a v7 uuid generating a v1 instead", "Error", err)
		random_uuid = uuid.New()
	}
	comment := config.AppName + ":" + host + ":" + random_uuid.String()
	filename := "rsa_" + safeHost + "_" + time.Now().Format("20060102")
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
	random_uuid, err := uuid.NewV7()
	if err != nil {
		slog.Warn("Failed to generate a v7 uuid generating a v1 instead", "Error", err)
		random_uuid = uuid.New()
	}
	comment := config.AppName + ":" + host + ":" + random_uuid.String()
	filename := "ecdsa_" + safeHost + "_" + time.Now().Format("20060102")
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
	random_uuid, err := uuid.NewV7()
	if err != nil {
		slog.Warn("Failed to generate a v7 uuid generating a v1 instead", "Error", err)
		random_uuid = uuid.New()
	}
	comment := config.AppName + ":" + host + ":" + random_uuid.String()
	filename := "ed25519_" + safeHost + "_" + time.Now().Format("20060102")
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
