/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agent

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-agent/lib/syncthing"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	"github.com/urfave/cli"
)

const cookieMaxAge = "3600"

// Context holds the Agent context structure
type Context struct {
	ProgName      string
	Config        *xdsconfig.Config
	Log           *logrus.Logger
	LogLevelSilly bool
	LogSillyf     func(format string, args ...interface{})
	SThg          *st.SyncThing
	SThgCmd       *exec.Cmd
	SThgInotCmd   *exec.Cmd

	webServer  *WebServer
	xdsServers map[string]*XdsServer
	sessions   *Sessions
	events     *Events
	projects   *Projects

	Exit chan os.Signal
}

// NewAgent Create a new instance of Agent
func NewAgent(cliCtx *cli.Context) *Context {
	var err error

	// Set logger level and formatter
	log := cliCtx.App.Metadata["logger"].(*logrus.Logger)

	logLevel := cliCtx.GlobalString("log")
	if logLevel == "" {
		logLevel = "error" // FIXME get from Config DefaultLogLevel
	}
	if log.Level, err = logrus.ParseLevel(logLevel); err != nil {
		fmt.Printf("Invalid log level : \"%v\"\n", logLevel)
		os.Exit(1)
	}
	log.Formatter = &logrus.TextFormatter{}

	// Support silly logging (printed on log.debug)
	sillyVal, sillyLog := os.LookupEnv("XDS_LOG_SILLY")
	logSilly := sillyLog && sillyVal == "1"
	sillyFunc := func(format string, args ...interface{}) {
		if logSilly {
			log.Debugf("SILLY: "+format, args...)
		}
	}

	// Define default configuration
	ctx := Context{
		ProgName:      cliCtx.App.Name,
		Log:           log,
		LogLevelSilly: logSilly,
		LogSillyf:     sillyFunc,
		Exit:          make(chan os.Signal, 1),

		webServer:  nil,
		xdsServers: make(map[string]*XdsServer),
		events:     nil,
	}

	// register handler on SIGTERM / exit
	signal.Notify(ctx.Exit, os.Interrupt, syscall.SIGTERM)
	go handlerSigTerm(&ctx)

	return &ctx
}

// Run Main function called to run agent
func (ctx *Context) Run() (int, error) {
	var err error

	// Logs redirected into a file when logfile option or logsDir config is set
	ctx.Config.LogVerboseOut = os.Stderr
	if ctx.Config.FileConf.LogsDir != "" {
		if ctx.Config.Options.LogFile != "stdout" {
			logFile := ctx.Config.Options.LogFile

			fdL, err := os.OpenFile(logFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			if err != nil {
				msgErr := fmt.Errorf("Cannot create log file %s", logFile)
				return int(syscall.EPERM), msgErr
			}
			ctx.Log.Out = fdL

			ctx._logPrint("Logging file: %s\n", logFile)
		}

		logFileHTTPReq := filepath.Join(ctx.Config.FileConf.LogsDir, "xds-agent-verbose.log")
		fdLH, err := os.OpenFile(logFileHTTPReq, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			msgErr := fmt.Errorf("Cannot create log file %s", logFileHTTPReq)
			return int(syscall.EPERM), msgErr
		}
		ctx.Config.LogVerboseOut = fdLH

		ctx._logPrint("Logging file for HTTP requests: %s\n", logFileHTTPReq)
	}

	// Create syncthing instance when section "syncthing" is present in agent-config.json
	if ctx.Config.FileConf.SThgConf != nil {
		ctx.SThg = st.NewSyncThing(ctx.Config, ctx.Log)
	}

	// Start local instance of Syncthing and Syncthing-notify
	if ctx.SThg != nil {
		ctx.Log.Infof("Starting Syncthing...")
		ctx.SThgCmd, err = ctx.SThg.Start()
		if err != nil {
			return 2, err
		}
		fmt.Printf("Syncthing started (PID %d)\n", ctx.SThgCmd.Process.Pid)

		ctx.Log.Infof("Starting Syncthing-inotify...")
		ctx.SThgInotCmd, err = ctx.SThg.StartInotify()
		if err != nil {
			return 2, err
		}
		fmt.Printf("Syncthing-inotify started (PID %d)\n", ctx.SThgInotCmd.Process.Pid)

		// Establish connection with local Syncthing (retry if connection fail)
		time.Sleep(3 * time.Second)
		maxRetry := 30
		retry := maxRetry
		for retry > 0 {
			if err := ctx.SThg.Connect(); err == nil {
				break
			}
			ctx.Log.Infof("Establishing connection to Syncthing (retry %d/%d)", retry, maxRetry)
			time.Sleep(time.Second)
			retry--
		}
		if err != nil || retry == 0 {
			return 2, err
		}

		// Retrieve Syncthing config
		id, err := ctx.SThg.IDGet()
		if err != nil {
			return 2, err
		}
		ctx.Log.Infof("Local Syncthing ID: %s", id)

	} else {
		ctx.Log.Infof("Cloud Sync / Syncthing not supported")
	}

	// Create Web Server
	ctx.webServer = NewWebServer(ctx)

	// Sessions manager
	ctx.sessions = NewClientSessions(ctx, cookieMaxAge)

	// Create events management
	ctx.events = NewEvents(ctx)

	// Create projects management
	ctx.projects = NewProjects(ctx, ctx.SThg)

	// Run Web Server until exit requested (blocking call)
	if err = ctx.webServer.Serve(); err != nil {
		log.Println(err)
		return 3, err
	}

	return 4, fmt.Errorf("Program exited")
}

// Helper function to log message on both stdout and logger
func (ctx *Context) _logPrint(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	if ctx.Log.Out != os.Stdout {
		ctx.Log.Infof(format, args...)
	}
}

// Handle exit and properly stop/close all stuff
func handlerSigTerm(ctx *Context) {
	<-ctx.Exit
	if ctx.SThg != nil {
		ctx.Log.Infof("Stoping Syncthing... (PID %d)",
			ctx.SThgCmd.Process.Pid)
		ctx.Log.Infof("Stoping Syncthing-inotify... (PID %d)",
			ctx.SThgInotCmd.Process.Pid)
		ctx.SThg.Stop()
		ctx.SThg.StopInotify()
	}
	if ctx.webServer != nil {
		ctx.Log.Infof("Stoping Web server...")
		ctx.webServer.Stop()
	}
	os.Exit(1)
}
