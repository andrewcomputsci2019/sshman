package main

import (
	"andrew/sshman/internal/buildInfo"
	"andrew/sshman/internal/config"
	"andrew/sshman/internal/flags"
	"andrew/sshman/internal/sqlite"
	"andrew/sshman/internal/sshParser"
	"andrew/sshman/internal/sshUtils"
	"andrew/sshman/internal/tui"
	"andrew/sshman/internal/utils"
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
)

type optionFlags []string

func (o *optionFlags) String() string {
	return fmt.Sprintf("%v", *o)
}

func (o *optionFlags) Set(value string) error {
	*o = append(*o, strings.TrimSpace(value))
	return nil
}

type noopHandler struct{}

func (h *noopHandler) Enabled(_ context.Context, level slog.Level) bool   { return false }
func (h *noopHandler) Handle(_ context.Context, record slog.Record) error { return nil }
func (h *noopHandler) WithAttrs(attrs []slog.Attr) slog.Handler           { return h }
func (h *noopHandler) WithGroup(name string) slog.Handler                 { return h }

func main() {
	// log level setting
	logLevel := flag.String("log", "None", "logging level, [Debug, Info, Warning, Error]")
	logFile := flag.Bool("logFile", true, "set output of logger to log file instead of terminal")
	// first run flag -- init
	init := flag.Bool("init", false, "initialize ssh-man")
	de_init := flag.Bool("uninstall", false, "delete ssh-man owned resources")
	// quick action commands
	quickAdd := flag.Bool("qa", false, "quick add")
	quickDelete := flag.Bool("qd", false, "quick delete")
	quickEdit := flag.Bool("qe", false, "quick edit")
	quickConnect := flag.Bool("qc", false, "quick connect")
	quickSync := flag.Bool("qs", false, "quick sync")

	// debug flags
	// get host relies on user setting host alias flag
	getHost := flag.Bool("gh", false, "get a host config definition and print it")
	createConfigFlag := flag.Bool("cc", false, "create ssh config using sqlite database")
	_ = flag.Bool("update", false, "checks for an available update, on unix may prompt for auto update")
	// validate config flag
	validateConfig := flag.Bool("validate", false, "validate configuration")
	printConfig := flag.Bool("parse-config", false, "print configuration")

	// build info
	versionFlag := flag.Bool("version", false, "print version and exit")

	// run config parameter flags
	host := flags.NewStringSettableFlag("host", "", "host alias")
	// used for quickadd and quickedit
	hostname := flags.NewStringSettableFlag("hostname", "", "new hostname for given host alias")
	// these are only used for quick connect
	port := flags.NewUintSettableFlag("p", 22, "ssh port")
	identityFile := flags.NewStringSettableFlag("i", "", "identity file")
	// used for both quick connect and quick sync
	sshConfigFile := flags.NewStringSettableFlag("f", "", "ssh config file")

	forceSync := flag.Bool("fs", false, "force sync, ignores checksum and attempts to sync database with provided config file")
	// these are forwarded to the tui
	var sshConfigOptions optionFlags
	flag.Var(&sshConfigOptions, "o",
		"option flags, these are ssh options that you want passed, works for quick commands such as qe, and qa, in tui, "+
			"these are passed when invoking ssh, these should be passed as -o Key=Value")
	flag.Parse()
	if logLevel != nil && *logLevel != "None" {
		// set up slog here
		var writer io.Writer
		lowerCase := strings.ToLower(*logLevel)
		if *logFile {
			logFile, err := os.Create(path.Join(xdg.DataHome, config.DefaultAppStorePath, "app.log"))
			if err != nil {
				fmt.Printf("Failed to initialize logger")
				return
			}
			defer logFile.Close()
			writer = logFile
		} else {
			writer = os.Stdout
		}
		switch lowerCase {
		case "debug":
			slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		case "info":
			slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})))
		case "warn":
			slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelWarn,
			})))
		case "error":
			slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelError,
			})))
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unknown log level: %s, defaulting to info\n", *logLevel)
			slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})))
		}
	} else { // discard log message
		slog.SetDefault(slog.New(&noopHandler{}))
	}

	if *versionFlag {
		fmt.Printf("ssh-man:\n\tversion: %v\n\tbuild date: %v\n\tbuild os: %v\n\tbuild architecture: %v\n",
			fmt.Sprintf("%v.%v.%v", buildInfo.BuildMajor, buildInfo.BuildMinor, buildInfo.BuildPatch),
			buildInfo.BuildDate, buildInfo.BUILD_OS, buildInfo.BUILD_ARC)
		return
	}

	if *init {
		// run init procedure then exit
		err := utils.InitProjectStructure()
		if err != nil {
			slog.Error("Failed to initialize project structure", "error", err)
			os.Exit(1)
		}
		return
	}

	if *de_init {
		fmt.Printf("Do you want to delete/uninstall ssh-man [Y/N]\n")
		ioReader := bufio.NewReader(os.Stdin)
		if input, err := ioReader.ReadString('\n'); err != nil {
			slog.Error("failed to get user input from stdin", "error", err)
			os.Exit(1)
		} else {
			if strings.ToLower(input) != "y" {
				fmt.Printf("Aborting uninstall event\n")
				return
			}
		}

		// delete owned resources then quit
		err := utils.DeInitProjectStructure()
		if err != nil {
			fmt.Printf("Failed to delete owned resources. Error: %v\n", err)
			os.Exit(1)
		}
		return
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
		return
	}
	if *printConfig {
		config.PrintConfig(cfg)
		return
	}
	var closeResource func() // db conn closure function
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
		closeResource = func() {
			conn.Close()
		}
	} else {
		// use default path
		storagePath := filepath.Join(xdg.DataHome, config.DefaultAppStorePath, config.DatabaseDir, config.DatabaseName)
		conn, err := sqlite.CreateAndLoadDB(storagePath)
		if err != nil {
			slog.Error("Error loading storage", "error", err, "path", storagePath)
			_, _ = fmt.Fprintf(os.Stderr, "Failed to load storage file, please verify user has permission to access %s", storagePath)
			os.Exit(1)
		}
		dbAO = sqlite.NewHostDao(conn)
		closeResource = func() {
			conn.Close()
		}
	}
	defer closeResource()
	if *getHost {
		if !host.SetByUser {
			slog.Error("host must be set in order to print stored definition")
			_, _ = fmt.Fprintf(os.Stderr, "You must set host when getting stored definition")
			// make sure to close database object
			closeResource()
			os.Exit(1)
		}
		storedHost, err := dbAO.Get(host.Value)
		if err != nil {
			slog.Error("error getting host from database", "host", host.Value, "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Error getting host from database")
			closeResource()
			os.Exit(1)
		}
		fmt.Printf("Database compliant definition:\n %v\n", storedHost)
		hStr, err := sshParser.ConvertSQLiteHostToString(&storedHost)
		if err != nil {
			slog.Error("Failed to convert host into ssh compliant string", "error", err)
			closeResource()
			os.Exit(1)
		}
		fmt.Printf("SSH Config compliant definition:\n %v\n", hStr)
		return
	}

	if *createConfigFlag {
		if createSSHConfigFile(dbAO, cfg.GetSshConfigFilePath()) != nil {
			slog.Error("could not write ssh config file out")
			_, _ = fmt.Fprint(os.Stderr, "Failed to write ssh config file out\n")
			closeResource()
			os.Exit(1)
		}
		return
	}

	/*
		QUICK ACTION PARSING
	*/
	if *quickConnect {
		// handle quick connect here
		if !host.SetByUser {
			_, _ = fmt.Fprintf(os.Stderr, "You must set host when using quick connect\n")
			slog.Error("host not set in quick connect")
			closeResource()
			os.Exit(1)
		}
		if port.SetByUser && port.Value == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "Provided port number: %v, is not valid\n", port.Value)
			slog.Error("port provided is invalid", "port", port)
			closeResource()
			os.Exit(1)
		}

		_, err := dbAO.Get(host.Value)
		if !sshConfigFile.SetByUser && err != nil {
			slog.Error("Host does not exist in table exiting", "host", host.Value)
			closeResource()
			os.Exit(1)
		}
		var configPath string
		if sshConfigFile.SetByUser {
			configPath = sshConfigFile.Value
		} else {
			configPath = cfg.GetSshConfigFilePath()
		}

		// compile options together
		options := make([]string, 0)
		if port.SetByUser {
			options = append(options, "-p")
			options = append(options, strconv.FormatUint(uint64(port.Value), 10))
		}
		if identityFile.SetByUser {
			options = append(options, "-i")
			options = append(options, identityFile.Value)
		}
		if sshConfigFile.SetByUser {
			options = append(options, "-F")
			options = append(options, sshConfigFile.Value)
		}
		for _, opt := range sshConfigOptions {
			options = append(options, "-o")
			options = append(options, opt)
		}
		cmd := createSSHCommand(host.Value, cfg.Ssh.ExcPath, configPath, options...)
		// slave current tty to cmd
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Run()
		return
	}

	if *quickSync {
		if !sshConfigFile.SetByUser {
			_, _ = fmt.Fprintf(os.Stderr, "You must specify a config file to sync against if using quickSync\n")
			closeResource()
			os.Exit(1)
		}
		filePath := sshConfigFile.String()
		if forceSync != nil && *forceSync {

		}
		// check checksum and see if file has already been checked against
		isSame, err := sshParser.IsSame(filePath)
		if err != nil {
			slog.Error("Error getting checksum of config file", "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Verify config file exist and is readable by current login user\nFile given: %s\nerr: %v", filePath, err)
			closeResource()
			os.Exit(1)
		}

		if isSame {
			slog.Info("Config file already has been synced before, skipping")
			return
		}

		hostsFromConfig, err := sshParser.ReadConfig(filePath) // get host defs from config
		if err != nil {
			slog.Error("Error reading config file", "error", err)
			closeResource()
			os.Exit(1)
		}

		conflictPolicy := cfg.StorageConf.ConflictPolicy
		switch conflictPolicy {
		case string(config.ConflictAlwaysError):
			err := dbAO.InsertMany(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database, please see error %v.\n", err)
				closeResource()
				os.Exit(1)
			}
		case string(config.ConflictIgnore):
			err := dbAO.InsertManyIgnoreConflict(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database due to internal error", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database due to internal error, please see error and try again: %v.\n", err)
				closeResource()
				os.Exit(1)
			}
		case string(config.ConflictFavorConfig):
			// upsert hosts into database
			err := dbAO.InsertOrUpdateMany(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database, please see error %v.\n", err)
				closeResource()
				os.Exit(1)
			}
		default:
			err := dbAO.InsertMany(hostsFromConfig...)
			if err != nil {
				slog.Error("failed to sync host into database", "error", err)
				_, _ = fmt.Fprintf(os.Stderr, "Failed to sync host into database, please see error %v.\n", err)
				closeResource()
				os.Exit(1)
			}
		}
		err = sshParser.DumpCheckSum(filePath)
		if err != nil {
			slog.Error("Failed to dump checksum of config file", "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Sync finished but errored when dumping checksum of given config file\n")
			closeResource()
			os.Exit(1)
		}
		if createSSHConfigFile(dbAO, cfg.GetSshConfigFilePath()) != nil {
			slog.Error("could not write ssh config file out")
			_, _ = fmt.Fprint(os.Stderr, "Failed to write ssh config file out\n")
			closeResource()
			os.Exit(1)
		}
		return
	}

	if *quickEdit {
		// handle quick edit here
		if !host.SetByUser {
			_, _ = fmt.Fprintf(os.Stderr, "You must set host when using quick Edit\n")
			return
		}
		// check for host existence in database
		dbHost, err := dbAO.Get(host.Value)
		if err != nil {
			slog.Error("Error getting host", "error", err)
			_, _ = fmt.Fprintf(os.Stderr, "Failed to get host, please verify host is valid\n")
			closeResource()
			os.Exit(1)
		}
		//convert hostOpts to a map of host-ops, note its a list due to a couple of options that can muti values
		optMap := make(map[string][]sqlite.HostOptions)
		for _, opt := range dbHost.Options {
			// iterate over option and plug into the map
			optMap[opt.Key] = append(optMap[opt.Key], opt)
		}
		if hostname.SetByUser {
			optMap["HostName"][0] = sqlite.HostOptions{
				Host:  dbHost.Host,
				Key:   "HostName",
				Value: hostname.Value,
			}
		}
		for _, option := range sshConfigOptions {
			split := strings.Split(option, "=")
			key, value := strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
			if !sshUtils.IsAcceptableOption(key) {
				slog.Warn("Skipping unknown option from user provided list", "key", key)
				continue
			}
			if sshUtils.IsOptionYesNo(key) && !sshUtils.YesNoOptionValid(strings.ToLower(value)) {
				slog.Warn("Skipping config option as it was detected as a ssh yes no option yet was not provided a valid yes no value", "key", key, "value", value)
				continue
			}
			if sshUtils.IsOptionYesNo(key) {
				value = strings.ToLower(value)
			}
			if _, ok := optMap[key]; ok {
				if sshUtils.OptionIsOfMutiType(key) {
					optMap[key] = append(optMap[key], sqlite.HostOptions{
						Host:  dbHost.Host,
						Key:   key,
						Value: value,
					})
				} else {
					optMap[key][0] = sqlite.HostOptions{
						Host:  dbHost.Host,
						Key:   key,
						Value: value,
					}
				}
			} else {
				optMap[key] = append(optMap[key], sqlite.HostOptions{
					Host:  dbHost.Host,
					Key:   key,
					Value: value,
				})
			}
		}
		replacementList := make([]sqlite.HostOptions, 0)
		for _, optArrays := range optMap {
			for _, vals := range optArrays {
				replacementList = append(replacementList, vals)
			}
		}
		dbHost.UpdatedAt = new(time.Time)
		*dbHost.UpdatedAt = time.Now()
		dbHost.Options = replacementList
		err = dbAO.Update(dbHost)
		if err != nil {
			slog.Error("failed to update host from quick edit command", "error", err, "updated-host", dbHost)
			closeResource()
			os.Exit(1)
		}
		if createSSHConfigFile(dbAO, cfg.GetSshConfigFilePath()) != nil {
			slog.Error("could not write ssh config file out")
			_, _ = fmt.Fprint(os.Stderr, "Failed to write ssh config file out\n")
			closeResource()
			os.Exit(1)
		}
		return
	}
	if *quickAdd {
		// handle quick add here
		if !host.SetByUser {
			_, _ = fmt.Fprint(os.Stderr, "Host must be set when adding a host")
			closeResource()
			os.Exit(1)
		}
		if !hostname.SetByUser {
			_, _ = fmt.Fprint(os.Stderr, "HostName must be set when adding a host")
			closeResource()
			os.Exit(1)
		}
		hOptions := make([]sqlite.HostOptions, 0)
		for _, option := range sshConfigOptions {
			split := strings.Split(option, "=")
			key, value := strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
			if !sshUtils.IsAcceptableOption(key) {
				slog.Warn("Skipping unknown option from user provided list", "key", key)
				continue
			}
			if sshUtils.IsOptionYesNo(key) && !sshUtils.YesNoOptionValid(strings.ToLower(value)) {
				slog.Warn("Skipping config option as it was detected as a ssh yes no option yet was not provided a valid yes no value", "key", key, "value", value)
				continue
			}
			if sshUtils.IsOptionYesNo(key) {
				value = strings.ToLower(value)
			}
			hOptions = append(hOptions, sqlite.HostOptions{
				Host:  host.Value,
				Key:   key,
				Value: value,
			})
		}
		sqHost := sqlite.Host{
			Host:      host.Value,
			CreatedAt: time.Now(),
			Notes:     "",
			Options:   hOptions,
		}
		err := dbAO.Insert(sqHost)
		if err != nil {
			slog.Error("Failed to add host to host table", "error", err)
			_, _ = fmt.Fprint(os.Stderr, "Failed to add host to host table\n")
			closeResource()
			os.Exit(1)
		}
		if sshParser.AddHostToFile(cfg.GetSshConfigFilePath(), sqHost) != nil {
			slog.Error("Failed to add host to ssh config file", "error", err)
			_, _ = fmt.Fprint(os.Stderr, "Failed to write ssh config file out\n")
			closeResource()
			os.Exit(1)
		}
		return
	}
	if *quickDelete {
		// handle quick delete here
		if !host.SetByUser {
			_, _ = fmt.Fprint(os.Stderr, "Host needs to be set when using quick delete")
			closeResource()
			os.Exit(1)
		}
		err := dbAO.Delete(sqlite.Host{
			Host: host.Value,
		})
		if err != nil {
			slog.Warn("Failed to delete host from table, or host does not exist in the table", "error", err)
			closeResource()
			os.Exit(1)
		}
		if createSSHConfigFile(dbAO, cfg.GetSshConfigFilePath()) != nil {
			slog.Error("could not write ssh config file out")
			_, _ = fmt.Fprint(os.Stderr, "Failed to write ssh config file out\n")
			closeResource()
			os.Exit(1)
		}
		return
	}

	/*
		OtherWise start tui
	*/
	hosts, err := dbAO.GetAll()
	if err != nil {
		slog.Error("Failed to fetch all hosts during startup")
		_, _ = fmt.Fprint(os.Stderr, "Failed to fetch all hosts during startup")
		return
	}
	prefixedOptions := make([]string, 0)
	for _, opt := range sshConfigOptions {
		prefixedOptions = append(prefixedOptions, "-o")
		prefixedOptions = append(prefixedOptions, opt)
	}
	app := tui.NewAppModel(hosts, dbAO, cfg, prefixedOptions...)
	program := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Printf("err: %s", err)
	}
}

func createSSHCommand(host string, sshPath string, configPath string, options ...string) *exec.Cmd {
	// note options need to be passed with a prefix of -o
	var c *exec.Cmd
	if sshPath == "" {
		args := []string{"-F", configPath}
		args = append(args, options...)
		args = append(args, host)
		c = exec.Command("ssh", args...)
	} else {
		args := []string{"-F", configPath}
		args = append(args, options...)
		args = append(args, host)
		c = exec.Command(sshPath, args...)
	}
	return c
}

func createSSHConfigFile(db *sqlite.HostDao, filePath string) error {
	allHosts, err := db.GetAll()
	if err != nil {
		return err
	}
	return sshParser.SerializeHostToFile(filePath, allHosts)
}
