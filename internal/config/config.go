package config

type Config struct {
	SshConfig SshConfig `yaml:"ssh_config"`
}

type SshConfig struct {
	SshConfigEnabled bool   `yaml:"enabled"`
	SshPath          string `yaml:"ssh_path"`
}
