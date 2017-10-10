package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

type version struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	VersionGitTag string `json:"gitTag"`
}

type apiVersion struct {
	Client version   `json:"client"`
	Server []version `json:"servers"`
}

// getInfo : return various information about server
func (s *APIService) getVersion(c *gin.Context) {
	response := apiVersion{
		Client: version{
			ID:            "",
			Version:       s.Config.Version,
			APIVersion:    s.Config.APIVersion,
			VersionGitTag: s.Config.VersionGitTag,
		},
	}

	svrVer := []version{}
	for _, svr := range s.xdsServers {
		res := version{}
		if err := svr.GetVersion(&res); err != nil {
			common.APIError(c, "Cannot retrieve version of XDS server ID %s : %v", svr.ID, err.Error())
			return
		}
		svrVer = append(svrVer, res)
	}
	response.Server = svrVer

	c.JSON(http.StatusOK, response)
}
