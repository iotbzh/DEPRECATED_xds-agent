package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

// Only define useful fields
type ExecArgs struct {
	ID string `json:"id" binding:"required"`
}

// ExecCmd executes remotely a command
func (s *APIService) execCmd(c *gin.Context) {
	s._execRequest("/exec", c)
}

// execSignalCmd executes remotely a command
func (s *APIService) execSignalCmd(c *gin.Context) {
	s._execRequest("/signal", c)
}

func (s *APIService) _execRequest(url string, c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		common.APIError(c, err.Error())
	}

	// First get Project ID to retrieve Server ID and send command to right server
	id := c.Param("id")
	if id == "" {
		args := ExecArgs{}
		// XXX - we cannot use c.BindJSON, so directly unmarshall it
		// (see https://github.com/gin-gonic/gin/issues/1078)
		if err := json.Unmarshal(data, &args); err != nil {
			common.APIError(c, "Invalid arguments")
			return
		}
		id = args.ID
	}
	prj := s.projects.Get(id)
	if prj == nil {
		common.APIError(c, "Unknown id")
		return
	}

	svr := (*prj).GetServer()
	if svr == nil {
		common.APIError(c, "Cannot identify XDS Server")
		return
	}

	// Retrieve session info
	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}
	sock := sess.IOSocket
	if sock == nil {
		common.APIError(c, "Websocket not established")
		return
	}

	// Forward XDSServer WS events to client WS
	// TODO removed static event name list and get it from XDSServer
	for _, evName := range []string{
		"exec:input",
		"exec:output",
		"exec:exit",
		"exec:inferior-input",
		"exec:inferior-output",
	} {
		evN := evName
		svr.EventOn(evN, func(evData interface{}) {
			(*sock).Emit(evN, evData)
		})
	}

	// Forward back command to right server
	response, err := svr.HTTPPostBody(url, string(data))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Decode response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		common.APIError(c, "Cannot read response body")
		return
	}
	c.JSON(http.StatusOK, string(body))

}