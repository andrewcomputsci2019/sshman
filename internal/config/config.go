package config

import (
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
)

type ConflictPolicy string

const AppName = "ssh_man"

const (
	ConflictIgnore      ConflictPolicy = "ignore"       // ignore conflicts dont sync data stores
	ConflictFavorConfig ConflictPolicy = "favor_config" // conflicts are resolved with using config version
	ConflictAlwaysError ConflictPolicy = "always_error" // if there's conflict error the program
)

type AcceptableKeyGenType string

const (
	RSA     = "RSA"
	ECDSA   = "ECDSA"
	ED25519 = "ED25519"
)

type NamedDirectories string

const (
	KeyStoreDir = "keystore"
	DatabaseDir = "db"
)

var KeyGenTypeSet = map[string]struct{}{
	RSA: {}, ECDSA: {}, ED25519: {},
}

type Config struct {
	StorageConf StorageConfig `yaml:"storage_config"`
	Ssh         SSH           `yaml:"ssh"`
	EnablePing  bool          `yaml:"enable_ping"` // trys to ping host to see if they are up and reports their ping
}

type StorageConfig struct {
	StoragePath    string `yaml:"storage_path,omitempty"`
	WriteThrough   *bool  `yaml:"write_through,omitempty"`   // by default WriteThrough is considered True
	ConflictPolicy string `yaml:"conflict_policy,omitempty"` // if not provided or illegal type defaults to ignore
}

type SSH struct {
	ExcPath                    string   `yaml:"executable_path,omitempty"`
	KeyOnly                    bool     `yaml:"key_only,omitempty"`
	KeyPath                    string   `yaml:"key_path,omitempty"`                  // where to store generated keys
	AcceptableKeyGenAlgorithms []string `yaml:"acceptable_key_algorithms,omitempty"` // Note this will reject DSA if provided
}

func (cfg *Config) String() string {
	if cfg == nil {
		return "<nil>"
	}
	builder := strings.Builder{}
	builder.WriteString("config:\n")
	builder.WriteString("Ping Enabled: ")
	builder.WriteString(strconv.FormatBool(cfg.EnablePing) + "\n")
	builder.WriteString("STORAGE_CONFIG:\n")
	builder.WriteString("\tStorage Path: ")
	if cfg.StorageConf.StoragePath == "" {
		pathBase := xdg.DataHome
		pathBase = path.Join(pathBase, DefaultAppStorePath)
		builder.WriteString(pathBase + "\n")
	} else {
		builder.WriteString(cfg.StorageConf.StoragePath + "\n")
	}
	builder.WriteString("\tWrite Through: ")
	if cfg.StorageConf.WriteThrough == nil {
		builder.WriteString("true\n")
	} else {
		builder.WriteString(strconv.FormatBool(*cfg.StorageConf.WriteThrough) + "\n")
	}
	builder.WriteString("\tConflict Policy: ")
	if cfg.StorageConf.ConflictPolicy == "" {
		builder.WriteString(string(ConflictAlwaysError) + "\n")
	} else {
		builder.WriteString(string(cfg.StorageConf.ConflictPolicy) + "\n")
	}
	builder.WriteString("SSH:\n")
	builder.WriteString("\tExecutable Path: ")
	if cfg.Ssh.ExcPath == "" {
		sshPath, err := exec.LookPath("ssh")
		if err != nil {
			builder.WriteString("\n")
		} else {
			builder.WriteString(sshPath + "\n")
		}
	} else {
		builder.WriteString(cfg.Ssh.ExcPath + "\n")
	}
	builder.WriteString("\tKey Only: ")
	builder.WriteString(strconv.FormatBool(cfg.Ssh.KeyOnly) + "\n")
	builder.WriteString("\tKey Path: ")
	if cfg.Ssh.KeyPath == "" {
		pathBase := xdg.ConfigHome
		pathBase = path.Join(pathBase, DefaultAppStorePath, KeyStoreDir)
		builder.WriteString(pathBase + "\n")
	} else {
		builder.WriteString(cfg.Ssh.KeyPath + "\n")
	}
	builder.WriteString("\tAcceptable KeyGenAlgorithms: ")
	if cfg.Ssh.AcceptableKeyGenAlgorithms == nil {
		builder.WriteString("RSA,ECDSA,ED25519\n")
	} else {
		for idx, al := range cfg.Ssh.AcceptableKeyGenAlgorithms {
			builder.WriteString(al)
			if idx < len(cfg.Ssh.AcceptableKeyGenAlgorithms)-1 {
				builder.WriteString(",")
			}
		}
	}
	return builder.String()
}

func (c Config) GetSshConfigFilePath() string {
	return path.Join(xdg.ConfigHome, DefaultAppStorePath, "ssh/config")
}
