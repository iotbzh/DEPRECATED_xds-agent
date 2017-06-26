package xdsconfig

import (
	"fmt"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	common "github.com/iotbzh/xds-common/golib"
)

// Config parameters (json format) of /config command
type Config struct {
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	VersionGitTag string `json:"gitTag"`

	// Private / un-exported fields
	HTTPPort string         `json:"-"`
	FileConf *FileConfig    `json:"-"`
	Log      *logrus.Logger `json:"-"`
}

// Config default values
const (
	DefaultAPIVersion = "1"
	DefaultPort       = "8010"
	DefaultLogLevel   = "error"
)

// Init loads the configuration on start-up
func Init(ctx *cli.Context, log *logrus.Logger) (*Config, error) {
	var err error

	// Define default configuration
	c := Config{
		Version:       ctx.App.Metadata["version"].(string),
		APIVersion:    DefaultAPIVersion,
		VersionGitTag: ctx.App.Metadata["git-tag"].(string),

		HTTPPort: DefaultPort,
		Log:      log,
	}

	// config file settings overwrite default config
	c.FileConf, err = updateConfigFromFile(&c, ctx.GlobalString("config"))
	if err != nil {
		return nil, err
	}

	if c.FileConf.LogsDir != "" && !common.Exists(c.FileConf.LogsDir) {
		if err := os.MkdirAll(c.FileConf.LogsDir, 0770); err != nil {
			return nil, fmt.Errorf("Cannot create logs dir: %v", err)
		}
	}
	c.Log.Infoln("Logs directory: ", c.FileConf.LogsDir)

	return &c, nil
}

// UpdateAll Update the current configuration
func (c *Config) UpdateAll(newCfg Config) error {
	return fmt.Errorf("Not Supported")
}
