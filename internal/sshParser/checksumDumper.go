package sshParser

import (
	"andrew/sshman/internal/config"
	"bytes"
	"crypto/sha256"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
)

const (
	checksumDir = "checksums"
)

// todo add test

// IsSame checks if checksum is identical to previous know version
func IsSame(file string) (bool, error) {
	dataDir := xdg.DataHome
	filename := path.Base(file)
	ext := path.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	dumpLoc := path.Join(dataDir, config.AppName, checksumDir, filename)
	f, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		slog.Error("Failed to get checksum for file", "file", file, "err", err)
		return false, err
	}
	checksum := hash.Sum(nil)
	checkFile, err := os.Open(dumpLoc)
	if err != nil {
		slog.Warn("Checksum file does not exist", "checksum path", dumpLoc, "file", file, "error", err)
		return false, nil
	}
	defer checkFile.Close()
	data, err := io.ReadAll(checkFile)
	if err != nil {
		slog.Error("Failed to read checksum file", "file", file, "err", err)
		return false, nil
	}

	return bytes.Equal(checksum, data), nil
}

// DumpCheckSum dumps file hash checksum to file
func DumpCheckSum(file string) error {
	filename := path.Base(file)
	ext := path.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)
	fileDumpPath := path.Join(xdg.DataHome, config.AppName, checksumDir, filename)
	f, err := os.Open(file)
	if err != nil {
		slog.Error("Failed to open file to compute checksum for", "file", file, "err", err)
		return err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		slog.Error("Failed to get checksum for file", "file", file, "err", err)
		return err
	}
	checksum := hash.Sum(nil)
	err = os.WriteFile(fileDumpPath, checksum, 0644)
	if err != nil {
		slog.Error("Failed to write checksum for file", "file", file, "err", err)
		return err
	}
	return nil
}
