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
	"encoding/json"
	"os"
	"path"

	common "github.com/iotbzh/xds-common/golib"
)

type SyncThingConf struct {
	BinDir     string `json:"binDir"`
	Home       string `json:"home"`
	GuiAddress string `json:"gui-address"`
	GuiAPIKey  string `json:"gui-apikey"`
}

type XDSServerConf struct {
	URL       string `json:"url"`
	ConnRetry int    `json:"connRetry"`

	// private/not exported fields
	ID            string `json:"-"`
	APIBaseURL    string `json:"-"`
	APIPartialURL string `json:"-"`
}

type FileConfig struct {
	HTTPPort    string          `json:"httpPort"`
	WebAppDir   string          `json:"webAppDir"`
	LogsDir     string          `json:"logsDir"`
	XDSAPIKey   string          `json:"xds-apikey"`
	ServersConf []XDSServerConf `json:"xdsServers"`
	SThgConf    *SyncThingConf  `json:"syncthing"`
}

// readGlobalConfig reads configuration from a config file.
// Order to determine which config file is used:
//  1/ from command line option: "--config myConfig.json"
//  2/ $HOME/.xds/agent/agent-config.json file
//  3/ /etc/xds-agent/config.json file

func readGlobalConfig(c *Config, confFile string) error {

	searchIn := make([]string, 0, 3)
	if confFile != "" {
		searchIn = append(searchIn, confFile)
	}
	if homeDir := common.GetUserHome(); homeDir != "" {
		searchIn = append(searchIn, path.Join(homeDir, ".xds", "agent", "agent-config.json"))
	}

	searchIn = append(searchIn, "/etc/xds-agent/agent-config.json")

	var cFile *string
	for _, p := range searchIn {
		if _, err := os.Stat(p); err == nil {
			cFile = &p
			break
		}
	}
	if cFile == nil {
		c.Log.Infof("No config file found")
		return nil
	}

	c.Log.Infof("Use config file: %s", *cFile)

	// TODO move on viper package to support comments in JSON and also
	// bind with flags (command line options)
	// see https://github.com/spf13/viper#working-with-flags

	fd, _ := os.Open(*cFile)
	defer fd.Close()

	// Decode config file content and save it in a first variable
	fCfg := FileConfig{}
	if err := json.NewDecoder(fd).Decode(&fCfg); err != nil {
		return err
	}

	// Decode config file content and overwrite default settings
	fd.Seek(0, 0)
	json.NewDecoder(fd).Decode(&c.FileConf)

	// Disable Syncthing support when there is no syncthing field in config
	if fCfg.SThgConf == nil {
		c.FileConf.SThgConf = nil
	}

	// Support environment variables (IOW ${MY_ENV_VAR} syntax) in agent-config.json
	vars := []*string{
		&c.FileConf.LogsDir,
		&c.FileConf.WebAppDir,
	}
	if c.FileConf.SThgConf != nil {
		vars = append(vars, &c.FileConf.SThgConf.Home,
			&c.FileConf.SThgConf.BinDir)
	}
	for _, field := range vars {
		var err error
		*field, err = common.ResolveEnvVar(*field)
		if err != nil {
			return err
		}
	}

	return nil
}
