// Package utils provides helpful utils for managing program directories and other useful generic misc helpers
package utils

import (
	"andrew/sshman/internal/config"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// InitProjectStructure creates all necessary data directories need for the program to function correctly
func InitProjectStructure() error {
	err := createKeyStorageIfNotExist()
	if err != nil {
		return err
	}
	err = createChecksumDirIfNotExist()
	if err != nil {
		return err
	}
	err = createSshConfigDirIfNotExist()
	if err != nil {
		return err
	}
	return nil
}

// todo test init function and verify that it correctly creates necessary program directories

// createKeyStorageIfNotExist creates the dir of $XDG_DATA_HOME/ssh_man/keystore
func createKeyStorageIfNotExist() error {
	checkSumPath := filepath.Join(xdg.ConfigHome, config.AppName, config.KeyStoreDir)
	err := os.MkdirAll(checkSumPath, 0760)
	if err != nil {
		slog.Error("Failed to create keystore directory", "path", checkSumPath, "error", err)
		return err
	}
	return nil
}

// createChecksumDirIfNotExist creates the dir of $XDG_DATA_HOME/ssh_man/checksums/
func createChecksumDirIfNotExist() error {
	checkSumPath := filepath.Join(xdg.DataHome, config.AppName, "checksums")
	err := os.MkdirAll(checkSumPath, 0760)
	if err != nil {
		slog.Error("Failed to create checksums directory", "path", checkSumPath, "error", err)
		return err
	}
	return nil
}

// createSshConfigDirIfNotExist create the dir of $XDG_CONFIG_HOME/ssh_man/ssh/
func createSshConfigDirIfNotExist() error {
	configPath := filepath.Join(xdg.ConfigHome, config.AppName, config.SshConfigPath)
	err := os.MkdirAll(filepath.Dir(configPath), 0760)
	if err != nil {
		slog.Error("Failed to create ssh config dir", "path", configPath, "error", err)
		return err
	}
	return nil
}
