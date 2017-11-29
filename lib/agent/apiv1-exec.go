package agent

import (
	"net/http"

	"github.com/franciscocpg/reflectme"
	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	uuid "github.com/satori/go.uuid"
)

// ExecCmd executes remotely a command
func (s *APIService) execCmd(c *gin.Context) {

	args := xaapiv1.ExecArgs{}
	if err := c.BindJSON(&args); err != nil {
		s.Log.Warningf("/exec invalid args, err=%v", err)
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

	// Forward input events from client to XDSServer through WS
	// TODO use XDSServer events names definition
	evtInList := []string{
		xaapiv1.ExecInEvent,
		xaapiv1.ExecInferiorInEvent,
	}
	for _, evName := range evtInList {
		evN := evName
		err := (*sock).On(evN, func(stdin string) {
			if s.LogLevelSilly {
				s.Log.Debugf("EXEC EVENT IN (%s) <<%v>>", evN, stdin)
			}
			svr.EventEmit(evN, stdin)
		})
		if err != nil {
			msgErr := "Error while registering WS for " + evN + " event"
			s.Log.Errorf(msgErr, ", err: %v", err)
			common.APIError(c, msgErr)
			return
		}
	}

	// Forward output events from XDSServer to client through WS
	// TODO use XDSServer events names definition
	var fwdFuncID []uuid.UUID
	evtOutList := []string{
		xaapiv1.ExecOutEvent,
		xaapiv1.ExecInferiorOutEvent,
	}
	for _, evName := range evtOutList {
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

			if s.LogLevelSilly {
				s.Log.Debugf("EXEC EVENT OUT (%s) <<%v>>", evN, evData)
			}

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
	exitFunc := func(privD interface{}, evData interface{}) error {
		evN := xaapiv1.ExecExitEvent

		pData := privD.(map[string]string)
		sid := pData["sessID"]
		prjID := pData["prjID"]

		// Add sessionID to event Data
		reflectme.SetField(evData, "sessionID", sid)

		// IO socket can be nil when disconnected
		so := s.sessions.IOSocketGet(sid)
		if so != nil {
			(*so).Emit(evN, evData)
		} else {
			s.Log.Infof("%s not emitted: WS closed (sid:%s)", evN, sid)
		}

		prj := s.projects.Get(prjID)
		if prj != nil {
			evD := evData.(map[string]interface{})
			cmdIDData, cmdIDExist := evD["cmdID"]
			svr := (*prj).GetServer()
			if svr != nil && cmdIDExist {
				svr.CommandDelete(cmdIDData.(string))
			} else {
				s.Log.Infof("%s: cannot retrieve server for sid=%s, prjID=%s, evD=%v", evN, sid, prjID, evD)
			}
		} else {
			s.Log.Infof("%s: cannot retrieve project for sid=%s, prjID=%s", evN, sid, prjID)
		}

		// cleanup listener
		for i, evName := range evtOutList {
			svr.EventOff(evName, fwdFuncID[i])
		}
		svr.EventOff(evN, exitFuncID)

		return nil
	}

	prjCfg := (*prj).GetProject()
	privData := map[string]string{"sessID": sess.ID, "prjID": prjCfg.ID}
	exitFuncID, err = svr.EventOn(xaapiv1.ExecExitEvent, privData, exitFunc)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Forward back command to right server
	res := xsapiv1.ExecResult{}
	xsArgs := &xsapiv1.ExecArgs{
		ID:              args.ID,
		SdkID:           args.SdkID,
		CmdID:           args.CmdID,
		Cmd:             args.Cmd,
		Args:            args.Args,
		Env:             args.Env,
		RPath:           args.RPath,
		TTY:             args.TTY,
		TTYGdbserverFix: args.TTYGdbserverFix,
		ExitImmediate:   args.ExitImmediate,
		CmdTimeout:      args.CmdTimeout,
	}
	if err := svr.CommandExec(xsArgs, &res); err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Add command to running commands list
	if err := svr.CommandAdd(res.CmdID, xsArgs); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, xaapiv1.ExecResult{Status: res.Status, CmdID: res.CmdID})
}

// execSignalCmd executes remotely the signal command
func (s *APIService) execSignalCmd(c *gin.Context) {

	args := xaapiv1.ExecSignalArgs{}
	if err := c.BindJSON(&args); err != nil {
		s.Log.Warningf("/signal invalid args, err=%v", err)
		common.APIError(c, "Invalid arguments")
		return
	}

	// Retrieve on which xds-server the command is running
	var svr *XdsServer
	var dataCmd interface{}
	for _, svr = range s.xdsServers {
		dataCmd = svr.CommandGet(args.CmdID)
		if dataCmd != nil {
			break
		}
	}
	if dataCmd == nil {
		common.APIError(c, "Cannot retrieve XDS Server for this cmdID")
		return
	}

	// Forward back command to right server
	res := xsapiv1.ExecSigResult{}
	xsArgs := &xsapiv1.ExecSignalArgs{
		CmdID:  args.CmdID,
		Signal: args.Signal,
	}
	if err := svr.CommandSignal(xsArgs, &res); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, xaapiv1.ExecSignalResult{Status: res.Status, CmdID: res.CmdID})
}
