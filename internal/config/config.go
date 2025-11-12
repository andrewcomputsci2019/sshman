package config

type ConflictPolicy string

const (
	ConflictIgnore      ConflictPolicy = "ignore"       // ignore conflicts dont sync data stores
	ConflictFavorDB     ConflictPolicy = "favor_db"     // conflicts are resolved with using db version
	ConflictFavorConfig ConflictPolicy = "favor_config" // conflicts are resolved with using config version
	ConflictAlwaysError ConflictPolicy = "always_error" // if there's conflict error the program
)

type Config struct {
	SshConf     SshConfig     `yaml:"ssh_config"`
	StorageConf StorageConfig `yaml:"storage_config"`
	Ssh         SSH           `yaml:"ssh"`
}

type SshConfig struct {
	SshConfigEnabled bool   `yaml:"enabled"`
	SshConfigPath    string `yaml:"config_path,omitempty"`
	Sync             bool   `yaml:"sync,omitempty"` // syncs backing storage to this config
}

type StorageConfig struct {
	StoragePath    string `yaml:"storage_path,omitempty"`
	WriteThrough   bool   `yaml:"write_through,omitempty"`
	ConflictPolicy string `yaml:"conflict_policy,omitempty"`
}

type SSH struct {
	ExcPath                    string   `yaml:"executable_path,omitempty"`
	KeyOnly                    bool     `yaml:"key_only,omitempty"`
	KeyPath                    string   `yaml:"key_path,omitempty"`                  // where to store generated keys
	AcceptableKeyGenAlgorithms []string `yaml:"acceptable_key_algorithms,omitempty"` // Note this will reject DSA if provided
}
