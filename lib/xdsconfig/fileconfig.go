package xdsconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
)

type SyncThingConf struct {
	BinDir     string `json:"binDir"`
	Home       string `json:"home"`
	GuiAddress string `json:"gui-address"`
	GuiAPIKey  string `json:"gui-apikey"`
}

type FileConfig struct {
	HTTPPort string         `json:"httpPort"`
	SThgConf *SyncThingConf `json:"syncthing"`
}

// getConfigFromFile reads configuration from a config file.
// Order to determine which config file is used:
//  1/ from command line option: "--config myConfig.json"
//  2/ $HOME/.xds/agent-config.json file
//  3/ <current_dir>/agent-config.json file
//  4/ <executable dir>/agent-config.json file

func updateConfigFromFile(c *Config, confFile string) (*FileConfig, error) {

	searchIn := make([]string, 0, 3)
	if confFile != "" {
		searchIn = append(searchIn, confFile)
	}
	if usr, err := user.Current(); err == nil {
		searchIn = append(searchIn, path.Join(usr.HomeDir, ".xds", "agent-config.json"))
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
	fCfg := FileConfig{}
	if cFile == nil {
		// No config file found
		return &fCfg, nil
	}

	c.log.Infof("Use config file: %s", *cFile)

	// TODO move on viper package to support comments in JSON and also
	// bind with flags (command line options)
	// see https://github.com/spf13/viper#working-with-flags

	fd, _ := os.Open(*cFile)
	defer fd.Close()
	if err := json.NewDecoder(fd).Decode(&fCfg); err != nil {
		return nil, err
	}

	// Support environment variables (IOW ${MY_ENV_VAR} syntax) in agent-config.json
	// TODO: better to use reflect package to iterate on fields and be more generic
	var rep string

	if rep, err = resolveEnvVar(fCfg.SThgConf.BinDir); err != nil {
		return nil, err
	}
	fCfg.SThgConf.BinDir = path.Clean(rep)

	if rep, err = resolveEnvVar(fCfg.SThgConf.Home); err != nil {
		return nil, err
	}
	fCfg.SThgConf.Home = path.Clean(rep)

	return &fCfg, nil
}

// resolveEnvVar Resolved environment variable regarding the syntax ${MYVAR}
func resolveEnvVar(s string) (string, error) {
	re := regexp.MustCompile("\\${(.*)}")
	vars := re.FindAllStringSubmatch(s, -1)
	res := s
	for _, v := range vars {
		val := os.Getenv(v[1])
		if val == "" {
			return res, fmt.Errorf("ERROR: %s env variable not defined", v[1])
		}

		rer := regexp.MustCompile("\\${" + v[1] + "}")
		res = rer.ReplaceAllString(res, val)
	}

	return res, nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
