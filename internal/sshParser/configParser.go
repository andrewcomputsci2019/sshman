// Package sshParser provides generic utilities for managing ssh host config files
package sshParser

import (
	"andrew/sshman/internal/sqlite"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
)

var (
	ErrInvalidHost = errors.New("invalid host given as ssh host")
)

// todo create a generic way to validate that this function indeed reads config file and gets all opts

func ReadConfig(file string) ([]sqlite.Host, error) {
	slog.Debug("Reading config", "file", file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}
	hosts := make([]sqlite.Host, 0)
	for idx, h := range cfg.Hosts {
		if idx == 0 { // first values is always a host * pattern
			continue
		}
		if len(h.Patterns) == 0 {
			continue
		}
		// ignore these types of host patterns
		if len(h.Patterns) == 1 && h.Patterns[0].String() == "*" {
			continue
		}
		matchedHosts := make([]*sqlite.Host, 0)
		// take all valid hostnames and parse them
		for _, pattern := range h.Patterns {
			if isFullQName(pattern.String()) {
				host := &sqlite.Host{
					Host:           pattern.String(),
					CreatedAt:      time.Now(),
					UpdatedAt:      nil,
					LastConnection: nil,
					Notes:          "",
				}
				matchedHosts = append(matchedHosts, host)
			}
		}
		notes := make([]string, 0)
		options := make([]sqlite.HostOptions, 0)
		// extract nodes here
		for _, node := range h.Nodes {
			// couple of types of Nodes KV, Empty and include (include we will support but not parse, basically treat as just a kv)
			switch bType := node.(type) {
			case *ssh_config.Empty:
				if bType.Comment != "" {
					notes = append(notes, bType.Comment)
				}
			case *ssh_config.Include:
				opt := sqlite.HostOptions{}
				opt.Host = ""
				includeString := bType.String()
				key := ""
				value := ""
				if includeString != "" {
					includeString = strings.TrimLeft(includeString, " ")
					key = includeString[:7]
					value = includeString[8:] // there will be either a space or = at pos 7, either way must ignore it
				} else {
					slog.Warn("Empty Include string was given", "file", file, "pos", bType.Pos().String())
					continue
				}
				opt.Key = key
				opt.Value = value
				options = append(options, opt)
				if bType.Comment != "" {
					notes = append(notes, key+": "+bType.Comment)
				}
			case *ssh_config.KV:
				// parse key value pair
				opt := sqlite.HostOptions{}
				opt.Host = ""
				opt.Key = bType.Key
				opt.Value = bType.Value
				options = append(options, opt)
				if bType.Comment != "" {
					notes = append(notes, bType.Key+": "+bType.Comment)
				}
			}
		}
		for _, host := range matchedHosts {
			host.Notes = strings.Join(notes, "\n")
			host.Options = make([]sqlite.HostOptions, len(options))
			copy(host.Options, options)
			for i := range host.Options {
				host.Options[i].Host = host.Host
			}
			hosts = append(hosts, *host)
		}
	}
	return hosts, nil
}

// isFullQName returns true if host string
func isFullQName(pattern string) bool {
	if len(pattern) == 0 || strings.ContainsAny(pattern, "*?!") || strings.Contains(pattern, "[]") {
		return false
	}
	return true
}

func serializeHostToSshHost(host *sqlite.Host) (ssh_config.Host, error) {
	if host == nil {
		return ssh_config.Host{}, errors.New("nil host")
	}
	sshHost := ssh_config.Host{
		Patterns: nil,
		Nodes:    nil,
	}
	if p, err := ssh_config.NewPattern(host.Host); err != nil {
		slog.Error("Failed to construct ssh host object with given host", "host", host, "err", err)
		return ssh_config.Host{}, ErrInvalidHost
	} else {
		sshHost.Patterns = append(sshHost.Patterns, p)
	}
	for _, opt := range host.Options {
		sshHost.Nodes = append(sshHost.Nodes, &ssh_config.KV{Key: opt.Key, Value: opt.Value})
	}
	for _, line := range strings.Split(host.Notes, "\n") {
		sshHost.Nodes = append(sshHost.Nodes, &ssh_config.Empty{Comment: line})
	}
	return sshHost, nil
}

// AddHostToFile is for append only operations to elevate blocking io required for full serialization
func AddHostToFile(file string, host sqlite.Host) error {
	slog.Debug("AddHostToFile", "host", host)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer f.Close()
	sshHost, err := serializeHostToSshHost(&host)
	if err != nil {
		return err
	}
	// // old approach to handle serialization before migrating to helper func
	//if p, err := ssh_config.NewPattern(host.Host); err != nil {
	//	slog.Error("Failed to construct ssh host object with given host", "host", host, "err", err)
	//	return ErrInvalidHost
	//} else {
	//	sshHost.Patterns = append(sshHost.Patterns, p)
	//}
	//for _, opt := range host.Options {
	//	sshHost.Nodes = append(sshHost.Nodes, &ssh_config.KV{Key: opt.Key, Value: opt.Value})
	//}
	//for _, line := range strings.Split(host.Notes, "\n") {
	//	sshHost.Nodes = append(sshHost.Nodes, &ssh_config.Empty{Comment: line})
	//}
	slog.Debug("AddHostToFile", "serialized Host", sshHost)
	_, err = f.WriteString(sshHost.String())
	if err != nil {
		return err
	}
	return nil
}

// SerializeHostToFile should be used to update and delete the config file as these actually save overhead
// This function will make a backup of the last working config before writing the new one
func SerializeHostToFile(file string, hosts []sqlite.Host) error {
	slog.Debug("SerializeHostToFile", "hosts", hosts)
	_, err := os.Stat(file)
	// file does  exist and backup needs to be made
	if err == nil {
		err := func() error {
			originalF, err := os.Open(file)
			if err != nil {

				return err
			}
			defer originalF.Close()
			dir, fileName := filepath.Split(file)
			copyFileName := filepath.Join(dir, fileName+".old")
			copyF, err := os.Create(copyFileName)
			if err != nil {
				return err
			}
			defer copyF.Close()
			_, err = io.Copy(copyF, originalF)
			return err
		}()
		if err != nil {
			return err
		}
	}
	serializedHosts := make([]ssh_config.Host, 0)
	for _, host := range hosts {
		sshHost, err := serializeHostToSshHost(&host)
		if err != nil {
			return err
		}
		serializedHosts = append(serializedHosts, sshHost)
	}
	f, err := os.OpenFile(file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	for _, host := range serializedHosts {
		_, err = f.WriteString(host.String())
		if err != nil {
			return err
		}
	}
	return nil
}
