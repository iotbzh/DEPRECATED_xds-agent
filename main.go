// TODO add Doc
//
package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/iotbzh/xds-agent/lib/agent"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
)

const (
	appName        = "xds-agent"
	appDescription = "X(cross) Development System Agent is a web server that allows to remotely cross build applications."
	appCopyright   = "Apache-2.0"
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
