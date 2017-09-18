package xdsconfig

import (
	"encoding/json"
	"os"
	"os/user"
	"path"
	"path/filepath"

	common "github.com/iotbzh/xds-common/golib"
)

type SyncThingConf struct {
	BinDir     string `json:"binDir"`
	Home       string `json:"home"`
	GuiAddress string `json:"gui-address"`
	GuiAPIKey  string `json:"gui-apikey"`
}

type FileConfig struct {
	HTTPPort  string         `json:"httpPort"`
	LogsDir   string         `json:"logsDir"`
	XDSAPIKey string         `json:"xds-apikey"`
	SThgConf  *SyncThingConf `json:"syncthing"`
}

// getConfigFromFile reads configuration from a config file.
// Order to determine which config file is used:
//  1/ from command line option: "--config myConfig.json"
//  2/ $HOME/.xds/agent/agent-config.json file
//  3/ <current_dir>/agent-config.json file
//  4/ <executable dir>/agent-config.json file

func updateConfigFromFile(c *Config, confFile string) (*FileConfig, error) {

	searchIn := make([]string, 0, 3)
	if confFile != "" {
		searchIn = append(searchIn, confFile)
	}
	if usr, err := user.Current(); err == nil {
		searchIn = append(searchIn, path.Join(usr.HomeDir, ".xds", "agent", "agent-config.json"))
	}
	cwd, err := os.Getwd()
	if err == nil {
		searchIn = append(searchIn, path.Join(cwd, "agent-config.json"))
	}
	exePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		searchIn = append(searchIn, path.Join(exePath, "agent-config.json"))
	}

	var cFile *string
	for _, p := range searchIn {
		if _, err := os.Stat(p); err == nil {
			cFile = &p
			break
		}
	}
	// Use default settings
	fCfg := *c.FileConf

	// Read config file when existing
	if cFile != nil {
		c.Log.Infof("Use config file: %s", *cFile)

		// TODO move on viper package to support comments in JSON and also
		// bind with flags (command line options)
		// see https://github.com/spf13/viper#working-with-flags

		fd, _ := os.Open(*cFile)
		defer fd.Close()
		if err := json.NewDecoder(fd).Decode(&fCfg); err != nil {
			return nil, err
		}
	}

	// Support environment variables (IOW ${MY_ENV_VAR} syntax) in agent-config.json
	vars := []*string{
		&fCfg.LogsDir,
	}
	if fCfg.SThgConf != nil {
		vars = append(vars, &fCfg.SThgConf.Home, &fCfg.SThgConf.BinDir)
	}
	for _, field := range vars {
		var err error
		*field, err = common.ResolveEnvVar(*field)
		if err != nil {
			return nil, err
		}
	}

	// Config file settings overwrite default config
	if fCfg.HTTPPort != "" {
		c.HTTPPort = fCfg.HTTPPort
	}

	// Set default apikey
	// FIXME - rework with dynamic key
	if fCfg.XDSAPIKey == "" {
		fCfg.XDSAPIKey = "1234abcezam"
	}

	return &fCfg, nil
}
