package config

const (
	// DefaultAppConfigPath Should be prefixed with AppData
	DefaultAppConfigPath = "ssh_man\\config.yml"
	// DefaultAppStorePath points to the path where app storage should go, note this should be prefixed with %APPDATA%
	DefaultAppStorePath = "ssh_man\\"
)

const (
	KeyStoreDir   = "ssh\\keystore"
	SshConfigPath = "ssh\\config"
	DatabaseDir   = "db"
	DatabaseName  = "hosts.db"
)
