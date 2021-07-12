package config

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/apache/trafficcontrol/cache-config/t3cutil"
	"github.com/apache/trafficcontrol/lib/go-log"
	"github.com/pborman/getopt/v2"
)

const AppName = "t3c-request"
const Version = "0.1"
const UserAgent = AppName + "/" + Version

type Cfg struct {
	CommandArgs      []string
	LogLocationDebug string
	LogLocationWarn  string
	LogLocationError string
	LogLocationInfo  string
	LoginDispersion  time.Duration
	t3cutil.TCCfg
}

func (cfg Cfg) DebugLog() log.LogLocation   { return log.LogLocation(cfg.LogLocationDebug) }
func (cfg Cfg) ErrorLog() log.LogLocation   { return log.LogLocation(cfg.LogLocationError) }
func (cfg Cfg) InfoLog() log.LogLocation    { return log.LogLocation(cfg.LogLocationInfo) }
func (cfg Cfg) WarningLog() log.LogLocation { return log.LogLocation(cfg.LogLocationWarn) }
func (cfg Cfg) EventLog() log.LogLocation   { return log.LogLocation(log.LogLocationNull) } // event logging is not used.

// Usage() writes command line options and usage to 'stderr'
func Usage() {
	getopt.PrintUsage(os.Stderr)
	os.Exit(0)
}

// InitConfig() intializes the configuration variables and loggers.
func InitConfig() (Cfg, error) {
	dispersionPtr := getopt.IntLong("login-dispersion", 'l', 0, "[seconds] wait a random number of seconds between 0 and [seconds] before login to traffic ops, default 0")
	cacheHostNamePtr := getopt.StringLong("cache-host-name", 'H', "", "Host name of the cache to generate config for. Must be the server host name in Traffic Ops, not a URL, and not the FQDN")
	getDataPtr := getopt.StringLong("get-data", 'D', "system-info", "non-config-file Traffic Ops Data to get. Valid values are update-status, packages, chkconfig, system-info, and statuses")
	toInsecurePtr := getopt.BoolLong("traffic-ops-insecure", 'I', "[true | false] ignore certificate errors from Traffic Ops")
	toTimeoutMSPtr := getopt.IntLong("traffic-ops-timeout-milliseconds", 't', 30000, "Timeout in milli-seconds for Traffic Ops requests, default is 30000")
	toURLPtr := getopt.StringLong("traffic-ops-url", 'u', "", "Traffic Ops URL. Must be the full URL, including the scheme. Required. May also be set with     the environment variable TO_URL")
	toUserPtr := getopt.StringLong("traffic-ops-user", 'U', "", "Traffic Ops username. Required. May also be set with the environment variable TO_USER")
	revalOnlyPtr := getopt.BoolLong("reval-only", 'r', "[true | false] whether to only fetch data needed to revalidate, versus all config data. Only used if get-data is config")
	disableProxyPtr := getopt.BoolLong("traffic-ops-disable-proxy", 'p', "[true | false] whether to not use any configure Traffic Ops proxy parameter. Only used if get-data is config")
	toPassPtr := getopt.StringLong("traffic-ops-password", 'P', "", "Traffic Ops password. Required. May also be set with the environment variable TO_PASS    ")
	helpPtr := getopt.BoolLong("help", 'h', "Print usage information and exit")
	versionPtr := getopt.BoolLong("version", 'V', "Print the app version")
	verbosePtr := getopt.CounterLong("verbose", 'v', `Log verbosity. Logging is output to stderr. By default, errors are logged. To log warnings, pass '-v'. To log info, pass '-vv'. To omit error logging, see '-s'`)
	silentPtr := getopt.BoolLong("silent", 's', `Silent. Errors are not logged, and the 'verbose' flag is ignored. If a fatal error occurs, the return code will be non-zero but no text will be output to stderr`)

	getopt.Parse()

	if *helpPtr == true {
		Usage()
	}
	if *versionPtr == true {
		fmt.Println(AppName + " v" + Version)
		os.Exit(0)
	}

	logLocationError := log.LogLocationStderr
	logLocationWarn := log.LogLocationNull
	logLocationInfo := log.LogLocationNull
	logLocationDebug := log.LogLocationNull
	if *silentPtr {
		logLocationError = log.LogLocationNull
	} else {
		if *verbosePtr >= 1 {
			logLocationWarn = log.LogLocationStderr
		}
		if *verbosePtr >= 2 {
			logLocationInfo = log.LogLocationStderr
			logLocationDebug = log.LogLocationStderr // t3c only has 3 verbosity options: none (-s), error (default or --verbose=0), warning (-v), and info (-vv). Any code calling log.Debug is treated as Info.
		}
	}

	if *verbosePtr > 2 {
		return Cfg{}, errors.New("Too many verbose options. The maximum log verbosity level is 2 (-vv or --verbose=2) for errors (0), warnings (1), and info (2)")
	}

	dispersion := time.Second * time.Duration(*dispersionPtr)
	toTimeoutMS := time.Millisecond * time.Duration(*toTimeoutMSPtr)
	toURL := *toURLPtr
	toUser := *toUserPtr
	toPass := *toPassPtr

	urlSourceStr := "argument" // for error messages
	if toURL == "" {
		urlSourceStr = "environment variable"
		toURL = os.Getenv("TO_URL")
	}
	if toUser == "" {
		toUser = os.Getenv("TO_USER")
	}
	if *toPassPtr == "" {
		toPass = os.Getenv("TO_PASS")
	}

	toURLParsed, err := url.Parse(toURL)
	if err != nil {
		return Cfg{}, errors.New("parsing Traffic Ops URL from " + urlSourceStr + " '" + toURL + "': " + err.Error())
	} else if err := t3cutil.ValidateURL(toURLParsed); err != nil {
		return Cfg{}, errors.New("invalid Traffic Ops URL from " + urlSourceStr + " '" + toURL + "': " + err.Error())
	}

	var cacheHostName string
	if len(*cacheHostNamePtr) > 0 {
		cacheHostName = *cacheHostNamePtr
	} else {
		cacheHostName, err = os.Hostname()
		if err != nil {
			return Cfg{}, errors.New("could not get the OS hostname, please supply a hostname: " + err.Error())
		}
	}

	cfg := Cfg{
		CommandArgs:      getopt.Args(),
		LogLocationDebug: logLocationDebug,
		LogLocationError: logLocationError,
		LogLocationInfo:  logLocationInfo,
		LogLocationWarn:  logLocationWarn,
		LoginDispersion:  dispersion,
		TCCfg: t3cutil.TCCfg{
			CacheHostName:  cacheHostName,
			GetData:        *getDataPtr,
			TOInsecure:     *toInsecurePtr,
			TOTimeoutMS:    toTimeoutMS,
			TOUser:         toUser,
			TOPass:         toPass,
			TOURL:          toURLParsed,
			UserAgent:      UserAgent,
			RevalOnly:      *revalOnlyPtr,
			TODisableProxy: *disableProxyPtr,
		},
	}

	if err := log.InitCfg(cfg); err != nil {
		return Cfg{}, errors.New("initializing loggers: " + err.Error())
	}

	return cfg, nil
}

func (cfg Cfg) PrintConfig() {
	fmt.Printf("CommandArgs: %s\n", cfg.CommandArgs)
	fmt.Printf("LogLocationDebug: %s\n", cfg.LogLocationDebug)
	fmt.Printf("LogLocationError: %s\n", cfg.LogLocationError)
	fmt.Printf("LogLocationInfo: %s\n", cfg.LogLocationInfo)
	fmt.Printf("LogLocationWarn: %s\n", cfg.LogLocationWarn)
	fmt.Printf("LoginDispersion : %s\n", cfg.LoginDispersion)
	fmt.Printf("CacheHostName: %s\n", cfg.CacheHostName)
	fmt.Printf("TOInsecure: %v\n", cfg.TOInsecure)
	fmt.Printf("TOTimeoutMS: %s\n", cfg.TOTimeoutMS)
	fmt.Printf("TOUser: %s\n", cfg.TOUser)
	fmt.Printf("TOPass: xxxxxx\n")
	fmt.Printf("TOURL: %s\n", cfg.TOURL)
}
