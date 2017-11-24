package agent

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/apiv1"
)

// getInfo : return various information about server
func (s *APIService) getVersion(c *gin.Context) {
	response := apiv1.XDSVersion{
		Client: apiv1.VersionData{
			ID:            "",
			Version:       s.Config.Version,
			APIVersion:    s.Config.APIVersion,
			VersionGitTag: s.Config.VersionGitTag,
		},
	}

	svrVer := []apiv1.VersionData{}
	for _, svr := range s.xdsServers {
		res := apiv1.VersionData{}
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
