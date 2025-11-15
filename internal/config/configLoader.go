package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

var ymlString []byte

func LoadConfig() Config {
	// todo
	cfg := Config{}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to get user home dir: %v", err)
		slog.Error("failed to get user home dir", "err", err)
		panic(err)
	}
	path := fmt.Sprintf("%s/%s", homeDir, DefaultAppConfigPath)
	file, err := os.ReadFile(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to read config file: %v", err)
		_, _ = fmt.Fprintln(os.Stderr, "verify config file exist at: "+path)
		slog.Error("Failed to read config file", "err", err, "file", path)
		panic(err)
	}
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to unmarshal config file: %v", err)
		slog.Error("failed to unmarshal config file", "err", err)
		panic(err)
	}
	ymlString = file
	return cfg
}

func ValidateConfig(config *Config) error {
	if config.StorageConf.StoragePath != "" {
		parentDir := filepath.Dir(config.StorageConf.StoragePath)
		parentDir = filepath.Clean(parentDir)
		_, err := os.Stat(parentDir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Storage path does not exist: %v", err)
			source, err := yaml.PathString("$.storage_config.storage_path")
			if err != nil {
				return err
			}
			annotation, err := source.AnnotateSource(ymlString, true)
			if err != nil {
				return err
			}
			fmt.Printf("expected valid storage location but given %s\n%s\n", config.StorageConf.StoragePath, string(annotation))
			return err
		}
		_, err = os.Stat(config.StorageConf.StoragePath)
		if err != nil {
			slog.Warn("storage database does not exist, disregard on first launch or after just setting location")
		}
	}
	if config.StorageConf.ConflictPolicy != "" {
		switch config.StorageConf.ConflictPolicy {
		case string(ConflictFavorConfig):
			break
		case string(ConflictIgnore):
			break
		case string(ConflictAlwaysError):
			break
		default:
			err := fmt.Errorf("unknown conflict_policy: %s", config.StorageConf.ConflictPolicy)
			source, errorYml := yaml.PathString("$.storage_config.conflict_policy")
			if errorYml != nil {
				return err
			}
			annotation, errorYml := source.AnnotateSource(ymlString, true)
			if errorYml != nil {
				return err
			}
			fmt.Printf("expected valid conflict policy but given %s\n%s\n", config.StorageConf.ConflictPolicy, string(annotation))
			return err
		}
	}

	if config.SshConf.SshConfigEnabled {
		sshFilePath := config.SshConf.SshConfigPath
		if sshFilePath != "" {
			_, err := os.Stat(sshFilePath)
			if err != nil {
				source, errorYml := yaml.PathString("$.ssh_config.config_path")
				if errorYml != nil {
					return err
				}
				annotation, errorYml := source.AnnotateSource(ymlString, true)
				if errorYml != nil {
					return err
				}
				fmt.Printf("ssh config file doesnt exist at: %s\n%s\n", sshFilePath, string(annotation))
				return err
			}
		}
	} else {
		if config.SshConf.SshConfigPath != "" {
			slog.Warn("ssh config file path given but config disables ssh config files. Enable ssh_conf.enabled in yaml file if you want config to load ssh config file")
		}
	}

	if config.Ssh.ExcPath != "" {
		fileInfo, err := os.Stat(config.Ssh.ExcPath)
		if err != nil {
			source, errorYml := yaml.PathString("$.ssh.executable_path")
			if errorYml != nil {
				return err
			}
			annotation, errorYml := source.AnnotateSource(ymlString, true)
			if errorYml != nil {
				return err
			}
			fmt.Printf("ssh executable does not exist at path: %s\n%s\n", config.Ssh.ExcPath, string(annotation))
			return err
		}
		if fileInfo.Mode().Perm()&0111 == 0 {
			err = fmt.Errorf("ssh executable path given is not an executable")
			source, errorYml := yaml.PathString("$.ssh.executable_path")
			if errorYml != nil {
				return err
			}
			annotation, errorYml := source.AnnotateSource(ymlString, true)
			if errorYml != nil {
				return err
			}
			fmt.Printf("ssh executable path given does not point to executable: %s\n%s\n", config.Ssh.ExcPath, string(annotation))
			return err
		}

	} else {
		_, err := exec.LookPath("ssh")
		if err != nil {
			return fmt.Errorf("ssh executable path does not exist: %v", err)
		}
	}

	if config.Ssh.AcceptableKeyGenAlgorithms != nil {
		for _, algorithm := range config.Ssh.AcceptableKeyGenAlgorithms {
			_, ok := KeyGenTypeSet[strings.ToUpper(algorithm)]
			if !ok {
				err := fmt.Errorf("unknown algorithm: %s", algorithm)
				source, errorYml := yaml.PathString("$.ssh.acceptable_key_algorithms")
				if errorYml != nil {
					return err
				}
				annotation, errorYml := source.AnnotateSource(ymlString, true)
				if errorYml != nil {
					return err
				}
				fmt.Printf("ssh executable path given does not point to executable: %s\n%s\n", algorithm, string(annotation))
				return err
			}
		}
	}
	return nil
}

func PrintConfig(cfg Config) {
	fmt.Printf("ssh_manager config:\n%s", cfg.String())
}
