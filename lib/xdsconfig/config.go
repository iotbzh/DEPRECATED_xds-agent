package xdsconfig

import (
	"fmt"
	"io"
	"path/filepath"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	common "github.com/iotbzh/xds-common/golib"
)

// Config parameters (json format) of /config command
type Config struct {
	Version       string
	APIVersion    string
	VersionGitTag string
	Options       Options
	FileConf      FileConfig
	Log           *logrus.Logger
	LogVerboseOut io.Writer
}

// Options set at the command line
type Options struct {
	ConfigFile string
	LogLevel   string
	LogFile    string
}

// Config default values
const (
	DefaultAPIVersion = "1"
	DefaultLogLevel   = "error"
)

// Init loads the configuration on start-up
func Init(ctx *cli.Context, log *logrus.Logger) (*Config, error) {
	var err error

	defaultWebAppDir := "${EXEPATH}/www"
	defaultSTHomeDir := "${HOME}/.xds/agent/syncthing-config"

	// Define default configuration
	c := Config{
		Version:       ctx.App.Metadata["version"].(string),
		APIVersion:    DefaultAPIVersion,
		VersionGitTag: ctx.App.Metadata["git-tag"].(string),

		Options: Options{
			ConfigFile: ctx.GlobalString("config"),
			LogLevel:   ctx.GlobalString("log"),
			LogFile:    ctx.GlobalString("logfile"),
		},

		FileConf: FileConfig{
			HTTPPort:  "8800",
			WebAppDir: defaultWebAppDir,
			LogsDir:   "/tmp/logs",
			// SEB XDSAPIKey: "1234abcezam",
			ServersConf: []XDSServerConf{
				XDSServerConf{
					URL:       "http://localhost:8000",
					ConnRetry: 10,
				},
			},
			SThgConf: &SyncThingConf{
				Home: defaultSTHomeDir,
			},
		},
		Log: log,
	}

	// config file settings overwrite default config
	err = readGlobalConfig(&c, c.Options.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Handle where Logs are redirected:
	//  default 'stdout' (logfile option default value)
	//  else use file (or filepath) set by --logfile option
	//  that may be overwritten by LogsDir field of config file
	logF := c.Options.LogFile
	logD := c.FileConf.LogsDir
	if logF != "stdout" {
		if logD != "" {
			lf := filepath.Base(logF)
			if lf == "" || lf == "." {
				lf = "xds-agent.log"
			}
			logF = filepath.Join(logD, lf)
		} else {
			logD = filepath.Dir(logF)
		}
	}
	if logD == "" || logD == "." {
		logD = "/tmp/xds/logs"
	}
	c.Options.LogFile = logF
	c.FileConf.LogsDir = logD

	if c.FileConf.LogsDir != "" && !common.Exists(c.FileConf.LogsDir) {
		if err := os.MkdirAll(c.FileConf.LogsDir, 0770); err != nil {
			return nil, fmt.Errorf("Cannot create logs dir: %v", err)
		}
	}
	c.Log.Infoln("Logs file:      ", c.Options.LogFile)
	c.Log.Infoln("Logs directory: ", c.FileConf.LogsDir)

	return &c, nil
}
