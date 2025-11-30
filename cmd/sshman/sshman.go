package main

import (
	"andrew/sshman/internal/buildInfo"
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/flags"
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/sshParser"
	"andrew/sshman/internal/utils"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
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
	// first run flag -- init
	init := flag.Bool("init", false, "initialize sshman")
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
	host := flags.NewStringSettableFlag("h", "", "host alias")
	hostname := flags.NewStringSettableFlag("hostname", "", "new hostname for given host alias")
	port := flags.NewUintSettableFlag("p", 22, "ssh port")
	identityFile := flags.NewStringSettableFlag("i", "", "identity file")
	sshConfigFile := flags.NewStringSettableFlag("f", "", "ssh config file")
	forceSync := flag.Bool("fs", false, "force sync, ignores checksum and attempts to sync database with provided config file")

	var sshConfigOptions optionFlags
	flag.Var(&sshConfigOptions, "o",
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

	if *init {
		// run init procedure then exit
		err := utils.InitProjectStructure()
		if err != nil {
			slog.Error("Failed to initialize project structure", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	/*
		Load Config here
	*/
	cfg := config.LoadConfig()
	err := config.ValidateConfig(&cfg)
	if err != nil {
		slog.Error("Error validating config", "error", err)
		os.Exit(1)
	}

	if *validateConfig {
		os.Exit(0)
	}
	if *printConfig {
		config.PrintConfig(cfg)
		os.Exit(0)
	}
	/*
		LoadDatabase here
	*/
	var dbAO *sqlite.HostDao // get database access object
	if cfg.StorageConf.StoragePath != "" {
		conn, err := sqlite.CreateAndLoadDB(cfg.StorageConf.StoragePath)
		if err != nil {
			slog.Error("Error loading storage", "error", err, "path", cfg.StorageConf.StoragePath)
			_, _ = fmt.Fprintf(os.Stderr, "Failed to load storage file, please verify path is valid %s", cfg.StorageConf.StoragePath)
			os.Exit(1)
		}
		dbAO = sqlite.NewHostDao(conn)
	} else {
		// use default path
		storagePath := path.Join(xdg.DataHome, config.DefaultAppStorePath, config.DatabaseDir)
		conn, err := sqlite.CreateAndLoadDB(storagePath)
		if err != nil {
			slog.Error("Error loading storage", "error", err, "path", storagePath)
			_, _ = fmt.Fprintf(os.Stderr, "Failed to load storage file, please verify user has permission to access %s", storagePath)
			os.Exit(1)
		}
		dbAO = sqlite.NewHostDao(conn)
	}

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
		if !sshConfigFile.SetByUser {
			_, _ = fmt.Fprintf(os.Stderr, "You must speficy a config file to sync against if using quickSync\n")
			os.Exit(1)
		}
		filePath := sshConfigFile.String()
		if forceSync != nil && *forceSync {

		}
		// check checksum and see if file has already been checked against
		isSame, err := sshParser.IsSame(filePath)
		if err != nil {
			slog.Error("Error getting checksum of config file", "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Verfiy config file exist and is readable by current login user\nFile given: %s\nerr: %v", filePath, err)
			os.Exit(1)
		}

		if isSame {
			slog.Info("Config file already has been synced before, skipping")
			os.Exit(0)
		}

		hostsFromConfig, err := sshParser.ReadConfig(filePath) // get host defs from config
		if err != nil {
			slog.Error("Error reading config file", "error", err)
			os.Exit(1)
		}

		conflictPolicy := cfg.StorageConf.ConflictPolicy
		switch conflictPolicy {
		case string(config.ConflictAlwaysError):
			err := dbAO.InsertMany(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database, please see error %v.\n", err)
				os.Exit(1)
			}
		case string(config.ConflictIgnore):
			err := dbAO.InsertManyIgnoreConflict(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database due to internal error", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database due to internal error, please see error and try again: %v.\n", err)
				os.Exit(1)
			}
		case string(config.ConflictFavorConfig):
			// upsert hosts into database
			err := dbAO.InsertOrUpdateMany(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database, please see error %v.\n", err)
				os.Exit(1)
			}
		default:
			err := dbAO.InsertMany(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database, please see error %v.\n", err)
				os.Exit(1)
			}
		}
		err = sshParser.DumpCheckSum(filePath)
		if err != nil {
			slog.Error("Failed to dump checksum of config file", "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Sync finished but errored when dumping checksum of given config file")
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *quickEdit {
		// handle quick edit here
		if !host.SetByUser {
			_, _ = fmt.Fprintf(os.Stderr, "You must set host when using quick Edit\n")
		}
		// check for host existence in database
		dbHost, err := dbAO.Get(host.Value)
		if err != nil {
			slog.Error("Error getting host", "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Failed to get host, please verify host is valid\n")
		}
		//convert hostOpts to a map of host-ops, note its a list due to a couple of options that can muti values
		optMap := make(map[string][]sqlite.HostOptions)
		if hostname.SetByUser {

		}

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
