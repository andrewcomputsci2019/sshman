// Package sshParser provides generic utilities for managing ssh host config files
package sshParser

import (
	"andrew/sshman/internal/sqlite"
	"fmt"
	"os"

	"github.com/kevinburke/ssh_config"
)

// todo create a generic way to validate that this function indeed reads config file and gets all opts

func ReadConfig(file string) ([]sqlite.Host, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg, _ := ssh_config.Decode(f)
	for _, h := range cfg.Hosts {
		fmt.Print("Host: [")
		for i, p := range h.Patterns {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Print(p.String())
		}
		fmt.Println("]")
		fmt.Print("Opts: {\n")
		for _, n := range h.Nodes {
			fmt.Print("\t")
			fmt.Print(n.String() + "\n")
		}
		fmt.Println("}")
	}
	return nil, nil
}

func DeleteHostFromFile(file string, host string) error {

	return nil
}

func AddHostToFile(file string, host *sqlite.Host) error {
	return nil
}

// UpdateHostToFile is a helper function that calls
func UpdateHostToFile(file string, host string) error {

	return nil
}
