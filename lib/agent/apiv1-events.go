package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

// EventRegisterArgs is the parameters (json format) of /events/register command
type EventRegisterArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"filterProjectID"`
}

// EventUnRegisterArgs is the parameters (json format) of /events/unregister command
type EventUnRegisterArgs struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// eventsList Registering for events that will be send over a WS
func (s *APIService) eventsList(c *gin.Context) {
	c.JSON(http.StatusOK, s.events.GetList())
}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsRegister(c *gin.Context) {
	var args EventRegisterArgs

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
	var args EventUnRegisterArgs

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
