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

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	common "github.com/iotbzh/xds-common/golib"
)

// getProjects returns all projects configuration
func (s *APIService) getProjects(c *gin.Context) {
	c.JSON(http.StatusOK, s.projects.GetProjectArr())
}

// getProject returns a specific project configuration
func (s *APIService) getProject(c *gin.Context) {
	id, err := s.projects.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	prj := s.projects.Get(id)
	if prj == nil {
		common.APIError(c, "Invalid id")
		return
	}

	c.JSON(http.StatusOK, (*prj).GetProject())
}

// addProject adds a new project to server config
func (s *APIService) addProject(c *gin.Context) {
	var cfgArg xaapiv1.ProjectConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	s.Log.Debugln("Add project config: ", cfgArg)

	newFld, err := s.projects.Add(cfgArg, s.sessions.GetID(c))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, newFld)
}

// syncProject force synchronization of project files
func (s *APIService) syncProject(c *gin.Context) {
	id, err := s.projects.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	s.Log.Debugln("Sync project id: ", id)

	err = s.projects.ForceSync(id)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, nil)
}

// delProject deletes project from server config
func (s *APIService) delProject(c *gin.Context) {
	id, err := s.projects.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	s.Log.Debugln("Delete project id ", id)

	delEntry, err := s.projects.Delete(id, s.sessions.GetID(c))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, delEntry)
}

// updateProject Update some field of a specific project
func (s *APIService) updateProject(c *gin.Context) {
	id, err := s.projects.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	var cfgArg xaapiv1.ProjectConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	s.Log.Debugln("Update project id ", id)

	upPrj, err := s.projects.Update(id, cfgArg, s.sessions.GetID(c))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, upPrj)
}
