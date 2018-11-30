// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/btcsuite/btclog"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrwallet/version"
	flags "github.com/jessevdk/go-flags"
	"github.com/raedahgroup/dcrchainanalysis/v1/networkconfig"
)

const (
	defaultLogLevel       = "info"
	defaultLogDirname     = "logs"
	defaultDataDirname    = "data"
	defaultDcrdHost       = "127.0.0.1"
	defaultDCAHost        = "127.0.0.1" // dcrchainanalysis tool default host
	defaultDCAPort        = "8476"      // dcrchainanalysis tool default port
	defaultConfigFilename = "dcrchainanalyser.conf"
	defaultLogFilename    = "dcrchainanalyser.log"
)

var (
	defaultAppDataDir        = dcrutil.AppDataDir("dcrchainanalyser", false)
	defaultConfigFile        = filepath.Join(defaultAppDataDir, defaultConfigFilename)
	defaultLogDir            = filepath.Join(defaultAppDataDir, defaultLogDirname)
	defaultDataDir           = filepath.Join(defaultAppDataDir, defaultDataDirname)
	dcrdHomeDir              = dcrutil.AppDataDir("dcrd", false)
	defaultDaemonRPCCertFile = filepath.Join(dcrdHomeDir, "rpc.cert")
)

type config struct {
	// General application behavior
	LogDir      string `long:"logdir" description:"Directory to log output."`
	AppDataDir  string `short:"A" long:"appdata" description:"Application data directory for wallet config, databases and logs"`
	ConfigFile  string `short:"C" long:"configfile" description:"Path to configuration file"`
	DebugLevel  string `short:"d" long:"debuglevel" description:"Logging level {trace, debug, info, warn, error, critical}"`
	ShowVersion bool   `short:"V" long:"version" description:"Display version information and exit"`
	TestNet     bool   `long:"testnet" description:"Use the test network (default mainnet)"`
	SimNet      bool   `long:"simnet" description:"Use the simulation test network (default mainnet)"`

	// DCA server configuration
	DCAHost string `long:"dcahost" description:"Chain analysis tool server host (default localhost)"`
	DCAPort string `long:"dcaport" description:"Chain analysis tool server host (default 8476)"`

	// RPC client options
	DcrdUser         string `long:"dcrduser" description:"Daemon RPC user name"`
	DcrdPass         string `long:"dcrdpass" description:"Daemon RPC password"`
	DcrdServ         string `long:"dcrdserv" description:"Hostname/IP and port of dcrd RPC server to connect to (default localhost:9109, testnet: localhost:19109, simnet: localhost:19556)"`
	DcrdCert         string `long:"dcrdcert" description:"File containing the dcrd certificate file"`
	DisableDaemonTLS bool   `long:"nodaemontls" description:"Disable TLS for the daemon RPC client -- NOTE: This is only allowed if the RPC client is connecting to localhost"`
}

// extraParams defines the extra parameters that could not be added to the
// config struct.
type extraParams struct {
	RemainingArgs []string
	ActiveNet     networkconfig.NetworkType
}

// cleanAndExpandPath expands environement variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// NOTE: The os.ExpandEnv doesn't work with Windows cmd.exe-style
	// %VARIABLE%, but they variables can still be expanded via POSIX-style
	// $VARIABLE.
	path = os.ExpandEnv(path)
	if !strings.HasPrefix(path, "~") {
		return filepath.Clean(path)
	}

	// Expand initial ~ to the current user's home directory, or ~otheruser
	// to otheruser's home directory.  On Windows, both forward and backward
	// slashes can be used.
	path = path[1:]
	var pathSeparators = string(os.PathSeparator)

	if runtime.GOOS == "windows" {
		pathSeparators = pathSeparators + "/"
	}

	userName := ""
	if i := strings.IndexAny(path, pathSeparators); i != -1 {
		userName = path[:i]
		path = path[i:]
	}

	homeDir := ""
	var u *user.User
	var err error
	if userName == "" {
		u, err = user.Current()
	} else {
		u, err = user.Lookup(userName)
	}

	if err == nil {
		homeDir = u.HomeDir
	}

	// Fallback to CWD if user lookup fails or user has no home directory.
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, path)
}

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	_, ok := btclog.LevelFromString(logLevel)
	return ok
}

// supportedSubsystems returns a sorted slice of the supported subsystems for
// logging purposes.
func supportedSubsystems() []string {
	// Convert the subsystemLoggers map keys to a slice.
	subsystems := make([]string, 0, len(subsystemLoggers))
	for subsysID := range subsystemLoggers {
		subsystems = append(subsystems, subsysID)
	}
	// Sort the subsytems for stable display.
	sort.Strings(subsystems)
	return subsystems
}

// parseAndSetDebugLevels attempts to parse the specified debug level and set
// the levels accordingly.  An appropriate error is returned if anything is
// invalid.
func parseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		if !validLogLevel(debugLevel) {
			str := "The specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}
		// Change the logging level for all subsystems.
		setLogLevels(debugLevel)
		return nil
	}

	// Split the specified string into subsystem/level pairs while detecting
	// issues and update the log levels accordingly.
	for _, logLevelPair := range strings.Split(debugLevel, ",") {
		if !strings.Contains(logLevelPair, "=") {
			str := "The specified debug level contains an invalid " +
				"subsystem/level pair [%v]"
			return fmt.Errorf(str, logLevelPair)
		}
		// Extract the specified subsystem and log level.
		fields := strings.Split(logLevelPair, "=")
		subsysID, logLevel := fields[0], fields[1]
		// Validate subsystem.
		if _, exists := subsystemLoggers[subsysID]; !exists {
			str := "The specified subsystem [%v] is invalid -- " +
				"supported subsytems %v"
			return fmt.Errorf(str, subsysID, supportedSubsystems())
		}
		// Validate log level.
		if !validLogLevel(logLevel) {
			str := "The specified debug level [%v] is invalid"
			return fmt.Errorf(str, logLevel)
		}
		setLogLevel(subsysID, logLevel)
	}
	return nil
}

// loadConfig initializes and fetches the configuration from a config file
// and command line options.
//
// The configuration proceeds as follows:
//      1) Start with a default config with sane settings
//      2) Pre-parse the command line to check for an alternative config file
//      3) Load configuration file overwriting defaults with any specified options
//      4) Parse CLI options and overwrite/add any specified options
//
// The above results in dcrwallet functioning properly without any config
// settings while still allowing the user to override settings with config files
// and command line options.  Command line options always take precedence.
// The bool returned indicates whether or not the wallet was recreated from a
// seed and needs to perform the initial resync. The []byte is the private
// passphrase required to do the sync for this special case.
func loadConfig() (*config, *extraParams, error) {
	loadConfigError := func(err error) (*config, *extraParams, error) {
		return nil, nil, err
	}
	// Default config.
	cfg := config{
		DebugLevel: defaultLogLevel,
		ConfigFile: defaultConfigFile,
		AppDataDir: defaultAppDataDir,
		LogDir:     defaultLogDir,
		DcrdCert:   defaultDaemonRPCCertFile,
	}

	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified.
	preCfg := cfg
	preParser := flags.NewParser(&preCfg, flags.Default)
	_, err := preParser.Parse()
	if err != nil {
		e, ok := err.(*flags.Error)
		if ok && e.Type == flags.ErrHelp {
			os.Exit(0)
		}
		preParser.WriteHelp(os.Stderr)
		return loadConfigError(err)
	}

	// Show the version and exit if the version flag was specified.
	//funcName := "loadConfig"
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	//usageMessage := fmt.Sprintf("Use %s -h to show usage", appName)
	if preCfg.ShowVersion {
		fmt.Printf("%s version %s (Go version %s)\n", appName, version.String(), runtime.Version())
		os.Exit(0)
	}

	// Load additional config from file.
	var configFileError error
	parser := flags.NewParser(&cfg, flags.Default)
	configFilePath := preCfg.ConfigFile
	if preCfg.ConfigFile != "" {
		configFilePath = cleanAndExpandPath(configFilePath)
	}
	err = flags.NewIniParser(parser).ParseFile(configFilePath)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Fprintln(os.Stderr, err)
			parser.WriteHelp(os.Stderr)
			return loadConfigError(err)
		}
		configFileError = err
	}

	// Parse command line options again to ensure they take precedence.
	remainingArgs, err := parser.Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		}
		return loadConfigError(err)
	}

	params := &extraParams{
		RemainingArgs: remainingArgs,
	}

	// Fetch the active network configured.
	switch {
	case cfg.TestNet && cfg.SimNet:
		fmt.Println("Testnet and Simnet should not be set simultaneously.")
		os.Exit(0)
	case cfg.TestNet:
		params.ActiveNet = networkconfig.TestNet
	case cfg.SimNet:
		params.ActiveNet = networkconfig.SimNet
	default:
		params.ActiveNet = networkconfig.MainNet
	}

	if cfg.DcrdServ == "" {
		cfg.DcrdServ = defaultDcrdHost + ":" + params.ActiveNet.RPCPort()
	}

	if cfg.DCAHost == "" {
		cfg.DCAHost = defaultDCAHost + ":" + defaultDCAPort
	}

	// Append the network type to the log directory so it is "namespaced"
	// per network.
	cfg.LogDir = cleanAndExpandPath(cfg.LogDir)
	// Special show command to list supported subsystems and exit.
	if cfg.DebugLevel == "show" {
		fmt.Println("Supported subsystems", supportedSubsystems())
		os.Exit(0)
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	initLogRotator(filepath.Join(cfg.LogDir, defaultLogFilename))
	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(cfg.DebugLevel); err != nil {
		err := fmt.Errorf("%s: %v", "loadConfig", err.Error())
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		return loadConfigError(err)
	}

	// Error and shutdown if config file is specified on the command line
	// but cannot be found.
	if (configFileError != nil) && cfg.ConfigFile == "" {
		if preCfg.ConfigFile == "" || cfg.ConfigFile == "" {
			log.Errorf("%v", configFileError)
			return loadConfigError(configFileError)
		}
	}

	// Warn about missing config file after the final command line parse
	// succeeds.  This prevents the warning on help messages and invalid
	// options.
	if configFileError != nil {
		log.Warnf("%v", configFileError)
	}
	return &cfg, params, nil
}
