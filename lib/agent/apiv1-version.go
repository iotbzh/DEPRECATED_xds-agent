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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
)

// getInfo : return various information about server
func (s *APIService) getVersion(c *gin.Context) {
	response := xaapiv1.XDSVersion{
		Client: xaapiv1.VersionData{
			ID:            "",
			Version:       s.Config.Version,
			APIVersion:    s.Config.APIVersion,
			VersionGitTag: s.Config.VersionGitTag,
		},
	}

	svrVer := []xaapiv1.VersionData{}
	for _, svr := range s.xdsServers {
		res := xaapiv1.VersionData{}
		if err := svr.GetVersion(&res); err != nil {
			errMsg := fmt.Sprintf("Cannot retrieve version of XDS server ID %s : %v", svr.ID, err.Error())
			s.Log.Warning(errMsg)
			res.ID = svr.ID
			res.Version = errMsg
		}
		svrVer = append(svrVer, res)
	}
	response.Server = svrVer

	c.JSON(http.StatusOK, response)
}
