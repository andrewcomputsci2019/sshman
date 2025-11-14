package main

import (
	"andrew/sshman/internal/buildInfo"
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/flags"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type optionFlags []string

func (o *optionFlags) String() string {
	return fmt.Sprintf("%v", *o)
}

func (o *optionFlags) Set(value string) error {
	*o = append(*o, value)
	return nil
}

func main() {
	// log level setting
	logLevel := flag.String("log", "None", "logging level, [Debug, Info, Warning, Error]")

	// quick action commands
	quickAdd := flag.Bool("qa", false, "quick add")
	quickDelete := flag.Bool("qd", false, "quick delete")
	quickEdit := flag.Bool("qe", false, "quick edit")
	quickConnect := flag.Bool("qc", false, "quick connect")
	quickSync := flag.Bool("qs", false, "quick sync")

	// validate config flag
	validateConfig := flag.Bool("validate", false, "validate configuration")
	printConfig := flag.Bool("parse-config", false, "print configuration")

	// build info
	versionFlag := flag.Bool("version", false, "print version and exit")

	// run config parameter flags
	host := flags.NewStringSettableFlag("h", "", "hostname")
	port := flags.NewUintSettableFlag("p", 0, "ssh port")
	identityFile := flags.NewStringSettableFlag("i", "", "identity file")
	sshConfigFile := flags.NewStringSettableFlag("f", "", "ssh config file")

	var optionFlags optionFlags
	flag.Var(&optionFlags, "o",
		"option flags, these are ssh options that you want passed, works for quick commands such as qe, and qa, in tui, "+
			"these are passed when invoking ssh, these should be passed identically to how they are written in a config file")
	flag.Parse()
	if logLevel != nil && *logLevel != "None" {
		// set up slog here
		lowerCase := strings.ToLower(*logLevel)
		switch lowerCase {
		case "debug":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		case "info":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})))
		case "warning":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelWarn,
			})))
		case "error":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelError,
			})))
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unkown log level: %s, defualting to info\n", *logLevel)
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})))
		}
	}

	if *versionFlag {
		fmt.Printf("ssh-man:\n\tversion: %v\n\tbuild date: %v\n\tbuild os: %v\n\tbuild architecture: %v\n",
			fmt.Sprintf("%v.%v.%v", buildInfo.BuildMajor, buildInfo.BuildMinor, buildInfo.BuildPatch),
			buildInfo.BuildDate, buildInfo.BUILD_OS, buildInfo.BUILD_ARC)
		os.Exit(0)
	}

	/*
		Load Config here
	*/
	_ = config.LoadConfig()

	if *validateConfig {
		os.Exit(0)
	}
	if *printConfig {
		// todo print config out to console
	}
	/*
		LoadDatabase here
	*/

	/*
		QUICK ACTION PARSING
	*/
	if *quickConnect {
		// handle quick connect here
		if !host.SetByUser {
			_, _ = fmt.Fprintf(os.Stderr, "You must set host when using quick connect\n")
			slog.Error("host not set in quick connect")
			os.Exit(1)
		}
		if port.SetByUser && port.Value == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "Proivded port number: %v, is not valid\n", port.Value)
			slog.Error("port provided is invalid", "port", port)
			os.Exit(1)
		}
		_ = port
		_ = host
		_ = identityFile

		os.Exit(0)
	}

	if *quickSync {
		_ = sshConfigFile
		os.Exit(0)
	}

	if *quickEdit {
		// handle quick edit here

		os.Exit(0)
	}
	if *quickAdd {
		// handle quick add here

		os.Exit(0)
	}
	if *quickDelete {
		// handle quick delete here

		os.Exit(0)
	}

	/*
		OtherWise start tui
	*/

}
