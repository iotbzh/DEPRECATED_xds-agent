package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/apiv1"
	common "github.com/iotbzh/xds-common/golib"
	uuid "github.com/satori/go.uuid"
)

var execCmdID = 1

// ExecCmd executes remotely a command
func (s *APIService) execCmd(c *gin.Context) {
	s._execRequest("/exec", c)
}

// execSignalCmd executes remotely a command
func (s *APIService) execSignalCmd(c *gin.Context) {
	s._execRequest("/signal", c)
}

func (s *APIService) _execRequest(cmd string, c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		common.APIError(c, err.Error())
	}

	args := apiv1.ExecArgs{}
	// XXX - we cannot use c.BindJSON, so directly unmarshall it
	// (see https://github.com/gin-gonic/gin/issues/1078)
	if err := json.Unmarshal(data, &args); err != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	// First get Project ID to retrieve Server ID and send command to right server
	id := c.Param("id")
	if id == "" {
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
	evtList := []string{
		apiv1.ExecInEvent,
		apiv1.ExecOutEvent,
		apiv1.ExecInferiorInEvent,
		apiv1.ExecInferiorOutEvent,
	}
	so := *sock
	fwdFuncID := []uuid.UUID{}
	for _, evName := range evtList {
		evN := evName
		fwdFunc := func(evData interface{}) {
			// Forward event to Client/Dashboard
			so.Emit(evN, evData)
		}
		id, err := svr.EventOn(evN, fwdFunc)
		if err != nil {
			common.APIError(c, err.Error())
			return
		}
		fwdFuncID = append(fwdFuncID, id)
	}

	// Handle Exit event separately to cleanup registered listener
	var exitFuncID uuid.UUID
	exitFunc := func(evData interface{}) {
		so.Emit(apiv1.ExecExitEvent, evData)

		// cleanup listener
		for i, evName := range evtList {
			svr.EventOff(evName, fwdFuncID[i])
		}
		svr.EventOff(apiv1.ExecExitEvent, exitFuncID)
	}
	exitFuncID, err = svr.EventOn(apiv1.ExecExitEvent, exitFunc)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Forward back command to right server
	response, err := svr.SendCommand(cmd, data)
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
