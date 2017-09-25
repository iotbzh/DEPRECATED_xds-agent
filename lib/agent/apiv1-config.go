package agent

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	common "github.com/iotbzh/xds-common/golib"
)

var confMut sync.Mutex

// APIConfig parameters (json format) of /config command
type APIConfig struct {
	Servers []ServerCfg `json:"servers"`

	// Not exposed outside in JSON
	Version       string `json:"-"`
	APIVersion    string `json:"-"`
	VersionGitTag string `json:"-"`
}

// ServerCfg .
type ServerCfg struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	APIURL     string `json:"apiUrl"`
	PartialURL string `json:"partialUrl"`
	ConnRetry  int    `json:"connRetry"`
	Connected  bool   `json:"connected"`
	Disabled   bool   `json:"disabled"`
}

// GetConfig returns the configuration
func (s *APIService) getConfig(c *gin.Context) {
	confMut.Lock()
	defer confMut.Unlock()

	cfg := s._getConfig()

	c.JSON(http.StatusOK, cfg)
}

// SetConfig sets configuration
func (s *APIService) setConfig(c *gin.Context) {
	var cfgArg APIConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	s.Log.Debugln("SET config: ", cfgArg)

	// First delete/disable XDS Server that are no longer listed
	for _, svr := range s.xdsServers {
		found := false
		for _, svrArg := range cfgArg.Servers {
			if svr.ID == svrArg.ID {
				found = true
				break
			}
		}
		if !found {
			s.DelXdsServer(svr.ID)
		}
	}

	// Add new XDS Server
	for _, svr := range cfgArg.Servers {
		cfg := xdsconfig.XDSServerConf{
			ID:        svr.ID,
			URL:       svr.URL,
			ConnRetry: svr.ConnRetry,
		}
		if _, err := s.AddXdsServer(cfg); err != nil {
			common.APIError(c, err.Error())
			return
		}
	}

	c.JSON(http.StatusOK, s._getConfig())
}

func (s *APIService) _getConfig() APIConfig {
	cfg := APIConfig{
		Version:       s.Config.Version,
		APIVersion:    s.Config.APIVersion,
		VersionGitTag: s.Config.VersionGitTag,
		Servers:       []ServerCfg{},
	}

	for _, svr := range s.xdsServers {
		cfg.Servers = append(cfg.Servers, ServerCfg{
			ID:         svr.ID,
			URL:        svr.BaseURL,
			APIURL:     svr.APIURL,
			PartialURL: svr.PartialURL,
			ConnRetry:  svr.ConnRetry,
			Connected:  svr.Connected,
			Disabled:   svr.Disabled,
		})
	}
	return cfg
}
