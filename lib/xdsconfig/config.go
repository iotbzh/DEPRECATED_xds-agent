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

package xdsconfig

import (
	"fmt"
	"io"
	"path/filepath"

	"os"

	"github.com/Sirupsen/logrus"
	common "github.com/iotbzh/xds-common/golib"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli"
)

// Config parameters (json format) of /config command
type Config struct {
	AgentUID      string
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

	// TODO: allocate uuid only the first time and save+reuse it later
	uuid := uuid.NewV1().String()

	// Define default configuration
	c := Config{
		AgentUID:      uuid,
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

	c.Log.Infoln("Agent UUID:     ", uuid)

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
