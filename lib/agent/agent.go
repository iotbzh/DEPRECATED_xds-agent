package agent

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/iotbzh/xds-agent/lib/syncthing"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	"github.com/iotbzh/xds-agent/lib/xdsserver"
)

// Context holds the Agent context structure
type Context struct {
	ProgName    string
	Config      *xdsconfig.Config
	Log         *logrus.Logger
	SThg        *st.SyncThing
	SThgCmd     *exec.Cmd
	SThgInotCmd *exec.Cmd
	WWWServer   *xdsserver.ServerService
	Exit        chan os.Signal
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

	// Define default configuration
	ctx := Context{
		ProgName: cliCtx.App.Name,
		Log:      log,
		Exit:     make(chan os.Signal, 1),
	}

	// register handler on SIGTERM / exit
	signal.Notify(ctx.Exit, os.Interrupt, syscall.SIGTERM)
	go handlerSigTerm(&ctx)

	return &ctx
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
	if ctx.WWWServer != nil {
		ctx.Log.Infof("Stoping Web server...")
		ctx.WWWServer.Stop()
	}
	os.Exit(1)
}
