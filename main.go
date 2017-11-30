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
 *
 *
 * xds-agent: X(cross) Development System client running on developer/local host.
 */

package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-agent/lib/agent"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	"github.com/urfave/cli"
)

const (
	appName        = "xds-agent"
	appDescription = "X(cross) Development System Agent is a web server that allows to remotely cross build applications."
	appCopyright   = "Copyright (C) 2017 IoT.bzh - Apache-2.0"
	appUsage       = "X(cross) Development System Agent"
)

var appAuthors = []cli.Author{
	cli.Author{Name: "Sebastien Douheret", Email: "sebastien@iot.bzh"},
}

// AppVersion is the version of this application
var AppVersion = "?.?.?"

// AppSubVersion is the git tag id added to version string
// Should be set by compilation -ldflags "-X main.AppSubVersion=xxx"
var AppSubVersion = "unknown-dev"

// xdsAgent main routine
func xdsAgent(cliCtx *cli.Context) error {
	var err error

	// Create Agent context
	ctxAgent := agent.NewAgent(cliCtx)

	// Load config
	ctxAgent.Config, err = xdsconfig.Init(cliCtx, ctxAgent.Log)
	if err != nil {
		return cli.NewExitError(err, 2)
	}

	// Run Agent (main loop)
	errCode, err := ctxAgent.Run()

	return cli.NewExitError(err, errCode)
}

// main
func main() {

	// Create a new instance of the logger
	log := logrus.New()

	// Create a new App instance
	app := cli.NewApp()
	app.Name = appName
	app.Description = appDescription
	app.Usage = appUsage
	app.Version = AppVersion + " (" + AppSubVersion + ")"
	app.Authors = appAuthors
	app.Copyright = appCopyright
	app.Metadata = make(map[string]interface{})
	app.Metadata["version"] = AppVersion
	app.Metadata["git-tag"] = AppSubVersion
	app.Metadata["logger"] = log

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "JSON config file to use\n\t",
			EnvVar: "XDS_CONFIGFILE",
		},
		cli.StringFlag{
			Name:   "log, l",
			Value:  "error",
			Usage:  "logging level (supported levels: panic, fatal, error, warn, info, debug)\n\t",
			EnvVar: "XDS_LOGLEVEL",
		},
		cli.StringFlag{
			Name:   "logfile",
			Value:  "stdout",
			Usage:  "filename where logs will be redirected (default stdout)\n\t",
			EnvVar: "XDS_LOGFILE",
		},
	}

	// only one action
	app.Action = xdsAgent

	app.Run(os.Args)
}
