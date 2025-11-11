//go:build !windows

package config

const (
	DefaultEnvConfigLookup = "$HOME"
	DefaultConfigPath      = "/.ssh/config"
)
