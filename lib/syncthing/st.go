package st

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"fmt"

	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-agent/lib/common"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	"github.com/syncthing/syncthing/lib/config"
)

// SyncThing .
type SyncThing struct {
	BaseURL string
	ApiKey  string
	Home    string
	STCmd   *exec.Cmd
	STICmd  *exec.Cmd

	// Private fields
	binDir      string
	exitSTChan  chan ExitChan
	exitSTIChan chan ExitChan
	client      *common.HTTPClient
	log         *logrus.Logger
}

// Monitor process exit
type ExitChan struct {
	status int
	err    error
}

// NewSyncThing creates a new instance of Syncthing
//func NewSyncThing(url string, apiKey string, home string, log *logrus.Logger) *SyncThing {
func NewSyncThing(conf *xdsconfig.SyncThingConf, log *logrus.Logger) *SyncThing {
	url := conf.GuiAddress
	apiKey := conf.GuiAPIKey
	home := conf.Home

	s := SyncThing{
		BaseURL: url,
		ApiKey:  apiKey,
		Home:    home,
		binDir:  conf.BinDir,
		log:     log,
	}

	if s.BaseURL == "" {
		s.BaseURL = "http://localhost:8384"
	}
	if s.BaseURL[0:7] != "http://" {
		s.BaseURL = "http://" + s.BaseURL
	}

	return &s
}

// Start Starts syncthing process
func (s *SyncThing) startProc(exeName string, args []string, env []string, eChan *chan ExitChan) (*exec.Cmd, error) {

	// Kill existing process (useful for debug ;-) )
	if os.Getenv("DEBUG_MODE") != "" {
		exec.Command("bash", "-c", "pkill -9 "+exeName).Output()
	}

	path, err := exec.LookPath(path.Join(s.binDir, exeName))
	if err != nil {
		return nil, fmt.Errorf("Cannot find %s executable in %s", exeName, s.binDir)
	}
	cmd := exec.Command(path, args...)
	cmd.Env = os.Environ()
	for _, ev := range env {
		cmd.Env = append(cmd.Env, ev)
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	*eChan = make(chan ExitChan, 1)
	go func(c *exec.Cmd) {
		status := 0
		cmdOut, err := c.StdoutPipe()
		if err == nil {
			s.log.Errorf("Pipe stdout error for : %s", err)
		} else if cmdOut != nil {
			stdOutput, _ := ioutil.ReadAll(cmdOut)
			fmt.Printf("STDOUT: %s\n", stdOutput)
		}
		sts, err := c.Process.Wait()
		if !sts.Success() {
			s := sts.Sys().(syscall.WaitStatus)
			status = s.ExitStatus()
		}
		*eChan <- ExitChan{status, err}
	}(cmd)

	return cmd, nil
}

// Start Starts syncthing process
func (s *SyncThing) Start() (*exec.Cmd, error) {
	var err error
	args := []string{
		"--home=" + s.Home,
		"-no-browser",
		"--gui-address=" + s.BaseURL,
	}

	if s.ApiKey != "" {
		args = append(args, "-gui-apikey=\""+s.ApiKey+"\"")
	}
	if s.log.Level == logrus.DebugLevel {
		args = append(args, "-verbose")
	}

	env := []string{
		"STNODEFAULTFOLDER=1",
	}

	s.STCmd, err = s.startProc("syncthing", args, env, &s.exitSTChan)

	return s.STCmd, err
}

// StartInotify Starts syncthing-inotify process
func (s *SyncThing) StartInotify() (*exec.Cmd, error) {
	var err error

	args := []string{
		"--home=" + s.Home,
		"-target=" + s.BaseURL,
	}
	if s.log.Level == logrus.DebugLevel {
		args = append(args, "-verbosity=4")
	}

	env := []string{}

	s.STICmd, err = s.startProc("syncthing-inotify", args, env, &s.exitSTIChan)

	return s.STICmd, err
}

func (s *SyncThing) stopProc(pname string, proc *os.Process, exit chan ExitChan) {
	if err := proc.Signal(os.Interrupt); err != nil {
		s.log.Errorf("Proc interrupt %s error: %s", pname, err.Error())

		select {
		case <-exit:
		case <-time.After(time.Second):
			// A bigger bonk on the head.
			if err := proc.Signal(os.Kill); err != nil {
				s.log.Errorf("Proc term %s error: %s", pname, err.Error())
			}
			<-exit
		}
	}
	s.log.Infof("%s stopped (PID %d)", pname, proc.Pid)
}

// Stop Stops syncthing process
func (s *SyncThing) Stop() {
	if s.STCmd == nil {
		return
	}
	s.stopProc("syncthing", s.STCmd.Process, s.exitSTChan)
	s.STCmd = nil
}

// StopInotify Stops syncthing process
func (s *SyncThing) StopInotify() {
	if s.STICmd == nil {
		return
	}
	s.stopProc("syncthing-inotify", s.STICmd.Process, s.exitSTIChan)
	s.STICmd = nil
}

// Connect Establish HTTP connection with Syncthing
func (s *SyncThing) Connect() error {
	var err error
	s.client, err = common.HTTPNewClient(s.BaseURL,
		common.HTTPClientConfig{
			URLPrefix:           "/rest",
			HeaderClientKeyName: "X-Syncthing-ID",
		})
	if err != nil {
		msg := ": " + err.Error()
		if strings.Contains(err.Error(), "connection refused") {
			msg = fmt.Sprintf("(url: %s)", s.BaseURL)
		}
		return fmt.Errorf("ERROR: cannot connect to Syncthing %s", msg)
	}
	if s.client == nil {
		return fmt.Errorf("ERROR: cannot connect to Syncthing (null client)")
	}
	return nil
}

// IDGet returns the Syncthing ID of Syncthing instance running locally
func (s *SyncThing) IDGet() (string, error) {
	var data []byte
	if err := s.client.HTTPGet("system/status", &data); err != nil {
		return "", err
	}
	status := make(map[string]interface{})
	json.Unmarshal(data, &status)
	return status["myID"].(string), nil
}

// ConfigGet returns the current Syncthing configuration
func (s *SyncThing) ConfigGet() (config.Configuration, error) {
	var data []byte
	config := config.Configuration{}
	if err := s.client.HTTPGet("system/config", &data); err != nil {
		return config, err
	}
	err := json.Unmarshal(data, &config)
	return config, err
}

// ConfigSet set Syncthing configuration
func (s *SyncThing) ConfigSet(cfg config.Configuration) error {
	body, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.client.HTTPPost("system/config", string(body))
}
