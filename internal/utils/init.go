// Package utils provides helpful utils for managing program directories and other useful generic misc helpers
package utils

import (
	"andrew/sshman/internal/config"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/goccy/go-yaml"
)

const (
	checksums = "checksums"
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
	err = createSQLiteDataStorePath()
	// todo dump default config into file
	cfg := config.GetDefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		slog.Error("Failed to generate yaml definition", "error", err)
	}
	cfgPath := filepath.Join(xdg.ConfigHome, config.DefaultAppConfigPath)
	cfgFile, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_RDWR, 0760)
	if err != nil {
		slog.Error("failed to create config file", "Path", cfgPath, "error", err)
		return err
	}
	_, err = cfgFile.Write(data)
	if err != nil {
		slog.Error("Failed to write to config file", "Path", cfgPath, "error", err)
		return err
	}
	return nil
}

// todo test init function and verify that it correctly creates necessary program directories

// createKeyStorageIfNotExist creates the dir of $XDG_CONFIG_HOME/ssh_man/ssh/keystore
func createKeyStorageIfNotExist() error {
	keystoreDir := filepath.Join(xdg.ConfigHome, config.AppName, config.KeyStoreDir)
	err := os.MkdirAll(keystoreDir, 0760)
	if err != nil {
		slog.Error("Failed to create keystore directory", "path", keystoreDir, "error", err)
		return err
	}
	return nil
}

// createChecksumDirIfNotExist creates the dir of $XDG_DATA_HOME/ssh_man/checksums/
func createChecksumDirIfNotExist() error {
	checkSumPath := filepath.Join(xdg.DataHome, config.AppName, checksums)
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

func createSQLiteDataStorePath() error {
	filePath := filepath.Join(xdg.DataHome, config.DefaultAppStorePath, config.DatabaseDir, config.DatabaseName)
	err := os.MkdirAll(filepath.Dir(filePath), 0760)
	if err != nil {
		slog.Error("Failed to create database storage directory", "path", filePath, "error", err)
		return err
	}
	return nil
}

func DeInitProjectStructure() error {
	if err := deleteConfigLocations(); err != nil {
		return err
	}
	if err := deleteDataLocations(); err != nil {
		return err
	}
	return nil
}

func deleteConfigLocations() error {
	configRoot := filepath.Dir(filepath.Join(xdg.ConfigHome, config.AppName, config.SshConfigPath))
	return os.RemoveAll(configRoot)
}

func deleteDataLocations() error {
	dataRoot := filepath.Join(xdg.DataHome, config.DefaultAppStorePath)
	return os.RemoveAll(dataRoot)
}
