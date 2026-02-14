//go:build !windows

package config

const (
	// DefaultAppConfigPath should be prefixed with XDG_CONFIG_HOME
	DefaultAppConfigPath = "ssh_man/config.yaml"
	// DefaultAppStorePath should be prefixed with XDG_DATA_HOME
	DefaultAppStorePath = "ssh_man"
)
const (
	KeyStoreDir   = "ssh/keystore"
	SshConfigPath = "ssh/config"
	DatabaseDir   = "db"
	DatabaseName  = "hosts.db"
)
