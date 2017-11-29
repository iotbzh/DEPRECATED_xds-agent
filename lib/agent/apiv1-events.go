package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	common "github.com/iotbzh/xds-common/golib"
)

// eventsList Registering for events that will be send over a WS
func (s *APIService) eventsList(c *gin.Context) {
	c.JSON(http.StatusOK, s.events.GetList())
}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsRegister(c *gin.Context) {
	var args xaapiv1.EventRegisterArgs

	if c.BindJSON(&args) != nil || args.Name == "" {
		common.APIError(c, "Invalid arguments")
		return
	}

	sess := s.webServer.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	// Register to all or to a specific events
	if err := s.events.Register(args.Name, sess.ID); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsUnRegister(c *gin.Context) {
	var args xaapiv1.EventUnRegisterArgs

	if c.BindJSON(&args) != nil || args.Name == "" {
		common.APIError(c, "Invalid arguments")
		return
	}

	sess := s.webServer.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	// Register to all or to a specific events
	if err := s.events.UnRegister(args.Name, sess.ID); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
