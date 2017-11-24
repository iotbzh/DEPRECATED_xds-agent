package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/franciscocpg/reflectme"
	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/apiv1"
	common "github.com/iotbzh/xds-common/golib"
	uuid "github.com/satori/go.uuid"
)

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
	iid := c.Param("id")
	if iid == "" {
		iid = args.ID
	}
	id, err := s.projects.ResolveID(iid)
	if err != nil {
		common.APIError(c, err.Error())
		return
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

	var fwdFuncID []uuid.UUID
	for _, evName := range evtList {
		evN := evName
		fwdFunc := func(pData interface{}, evData interface{}) error {
			sid := pData.(string)
			// IO socket can be nil when disconnected
			so := s.sessions.IOSocketGet(sid)
			if so == nil {
				s.Log.Infof("%s not emitted: WS closed (sid:%s)", evN, sid)
				return nil
			}

			// Add sessionID to event Data
			reflectme.SetField(evData, "sessionID", sid)

			// Forward event to Client/Dashboard
			(*so).Emit(evN, evData)
			return nil
		}
		id, err := svr.EventOn(evN, sess.ID, fwdFunc)
		if err != nil {
			common.APIError(c, err.Error())
			return
		}
		fwdFuncID = append(fwdFuncID, id)
	}

	// Handle Exit event separately to cleanup registered listener
	var exitFuncID uuid.UUID
	exitFunc := func(pData interface{}, evData interface{}) error {
		evN := apiv1.ExecExitEvent
		sid := pData.(string)

		// Add sessionID to event Data
		reflectme.SetField(evData, "sessionID", sid)

		// IO socket can be nil when disconnected
		so := s.sessions.IOSocketGet(sid)
		if so != nil {
			(*so).Emit(evN, evData)
		} else {
			s.Log.Infof("%s not emitted: WS closed (sid:%s)", evN, sid)
		}

		// cleanup listener
		for i, evName := range evtList {
			svr.EventOff(evName, fwdFuncID[i])
		}
		svr.EventOff(evN, exitFuncID)

		return nil
	}
	exitFuncID, err = svr.EventOn(apiv1.ExecExitEvent, sess.ID, exitFunc)
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
