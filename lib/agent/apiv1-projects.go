package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/apiv1"
	common "github.com/iotbzh/xds-common/golib"
)

// getProjects returns all projects configuration
func (s *APIService) getProjects(c *gin.Context) {
	c.JSON(http.StatusOK, s.projects.GetProjectArr())
}

// getProject returns a specific project configuration
func (s *APIService) getProject(c *gin.Context) {
	prj := s.projects.Get(c.Param("id"))
	if prj == nil {
		common.APIError(c, "Invalid id")
		return
	}

	c.JSON(http.StatusOK, (*prj).GetProject())
}

// addProject adds a new project to server config
func (s *APIService) addProject(c *gin.Context) {
	var cfgArg apiv1.ProjectConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	s.Log.Debugln("Add project config: ", cfgArg)

	newFld, err := s.projects.Add(cfgArg)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, newFld)
}

// syncProject force synchronization of project files
func (s *APIService) syncProject(c *gin.Context) {
	id := c.Param("id")

	s.Log.Debugln("Sync project id: ", id)

	err := s.projects.ForceSync(id)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

// delProject deletes project from server config
func (s *APIService) delProject(c *gin.Context) {
	id := c.Param("id")

	s.Log.Debugln("Delete project id ", id)

	delEntry, err := s.projects.Delete(id)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, delEntry)
}
