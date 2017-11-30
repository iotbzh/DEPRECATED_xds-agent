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

package agent

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	common "github.com/iotbzh/xds-common/golib"
)

var confMut sync.Mutex

// GetConfig returns the configuration
func (s *APIService) getConfig(c *gin.Context) {
	confMut.Lock()
	defer confMut.Unlock()

	cfg := s._getConfig()

	c.JSON(http.StatusOK, cfg)
}

// SetConfig sets configuration
func (s *APIService) setConfig(c *gin.Context) {
	var cfgArg xaapiv1.APIConfig
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

func (s *APIService) _getConfig() xaapiv1.APIConfig {
	cfg := xaapiv1.APIConfig{
		Version:       s.Config.Version,
		APIVersion:    s.Config.APIVersion,
		VersionGitTag: s.Config.VersionGitTag,
		Servers:       []xaapiv1.ServerCfg{},
	}

	for _, svr := range s.xdsServers {
		cfg.Servers = append(cfg.Servers, xaapiv1.ServerCfg{
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
