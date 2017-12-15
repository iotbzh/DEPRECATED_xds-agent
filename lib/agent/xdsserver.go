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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	uuid "github.com/satori/go.uuid"
	sio_client "github.com/sebd71/go-socket.io-client"
)

// XdsServer .
type XdsServer struct {
	*Context
	ID           string
	BaseURL      string
	APIURL       string
	PartialURL   string
	ConnRetry    int
	Connected    bool
	Disabled     bool
	ServerConfig *xsapiv1.APIConfig

	// Events management
	CBOnError      func(error)
	CBOnDisconnect func(error)
	sockEvents     map[string][]*caller
	sockEventsLock *sync.Mutex

	// Private fields
	client    *common.HTTPClient
	ioSock    *sio_client.Client
	logOut    io.Writer
	apiRouter *gin.RouterGroup
	cmdList   map[string]interface{}
}

// EventCB Event emitter callback
type EventCB func(privData interface{}, evtData interface{}) error

// caller Used to chain event listeners
type caller struct {
	id          uuid.UUID
	EventName   string
	Func        EventCB
	PrivateData interface{}
}

const _IDTempoPrefix = "tempo-"

// NewXdsServer creates an instance of XdsServer
func NewXdsServer(ctx *Context, conf xdsconfig.XDSServerConf) *XdsServer {
	return &XdsServer{
		Context:    ctx,
		ID:         _IDTempoPrefix + uuid.NewV1().String(),
		BaseURL:    conf.URL,
		APIURL:     conf.APIBaseURL + conf.APIPartialURL,
		PartialURL: conf.APIPartialURL,
		ConnRetry:  conf.ConnRetry,
		Connected:  false,
		Disabled:   false,

		sockEvents:     make(map[string][]*caller),
		sockEventsLock: &sync.Mutex{},
		logOut:         ctx.Log.Out,
		cmdList:        make(map[string]interface{}),
	}
}

// Close Free and close XDS Server connection
func (xs *XdsServer) Close() error {
	err := xs._Disconnected()
	xs.Disabled = true
	return err
}

// Connect Establish HTTP connection with XDS Server
func (xs *XdsServer) Connect() error {
	var err error
	var retry int

	xs.Disabled = false
	xs.Connected = false

	err = nil
	for retry = xs.ConnRetry; retry > 0; retry-- {
		if err = xs._CreateConnectHTTP(); err == nil {
			break
		}
		if retry == xs.ConnRetry {
			// Notify only on the first conn error
			// doing that avoid 2 notifs (conn false; conn true) on startup
			xs._NotifyState()
		}
		xs.Log.Infof("Establishing connection to XDS Server (retry %d/%d)", retry, xs.ConnRetry)
		time.Sleep(time.Second)
	}
	if retry == 0 {
		// FIXME: re-use _Reconnect to wait longer in background
		return fmt.Errorf("Connection to XDS Server failure")
	}
	if err != nil {
		return err
	}

	// Check HTTP connection and establish WS connection
	err = xs._Connect(false)

	return err
}

// IsTempoID returns true when server as a temporary id
func (xs *XdsServer) IsTempoID() bool {
	return strings.HasPrefix(xs.ID, _IDTempoPrefix)
}

// SetLoggerOutput Set logger ou
func (xs *XdsServer) SetLoggerOutput(out io.Writer) {
	xs.logOut = out
}

// SendCommand Send a command to XDS Server
func (xs *XdsServer) SendCommand(cmd string, body []byte, res interface{}) error {
	url := cmd
	if !strings.HasPrefix("/", cmd) {
		url = "/" + cmd
	}
	return xs.client.Post(url, string(body), res)
}

// GetVersion Send Get request to retrieve XDS Server version
func (xs *XdsServer) GetVersion(res interface{}) error {
	return xs.client.Get("/version", &res)
}

// GetFolders Send GET request to get current folder configuration
func (xs *XdsServer) GetFolders(folders *[]xsapiv1.FolderConfig) error {
	return xs.client.Get("/folders", folders)
}

// FolderAdd Send POST request to add a folder
func (xs *XdsServer) FolderAdd(fld *xsapiv1.FolderConfig, res interface{}) error {
	err := xs.client.Post("/folders", fld, res)
	if err != nil {
		return fmt.Errorf("FolderAdd error: %s", err.Error())
	}
	return err
}

// FolderDelete Send DELETE request to delete a folder
func (xs *XdsServer) FolderDelete(id string) error {
	return xs.client.HTTPDelete("/folders/" + id)
}

// FolderSync Send POST request to force synchronization of a folder
func (xs *XdsServer) FolderSync(id string) error {
	return xs.client.HTTPPost("/folders/sync/"+id, "")
}

// FolderUpdate Send PUT request to update a folder
func (xs *XdsServer) FolderUpdate(fld *xsapiv1.FolderConfig, resFld *xsapiv1.FolderConfig) error {
	return xs.client.Put("/folders/"+fld.ID, fld, resFld)
}

// CommandExec Send POST request to execute a command
func (xs *XdsServer) CommandExec(args *xsapiv1.ExecArgs, res *xsapiv1.ExecResult) error {
	return xs.client.Post("/exec", args, res)
}

// CommandSignal Send POST request to send a signal to a command
func (xs *XdsServer) CommandSignal(args *xsapiv1.ExecSignalArgs, res *xsapiv1.ExecSigResult) error {
	return xs.client.Post("/signal", args, res)
}

// SetAPIRouterGroup .
func (xs *XdsServer) SetAPIRouterGroup(r *gin.RouterGroup) {
	xs.apiRouter = r
}

// PassthroughGet Used to declare a route that sends directly a GET request to XDS Server
func (xs *XdsServer) PassthroughGet(url string) {
	if xs.apiRouter == nil {
		xs.Log.Errorf("apiRouter not set !")
		return
	}

	xs.apiRouter.GET(url, func(c *gin.Context) {
		var data interface{}
		// Take care of param (eg. id in /projects/:id)
		nURL := url
		if strings.Contains(url, ":") {
			nURL = strings.TrimPrefix(c.Request.URL.Path, xs.APIURL)
		}
		// Send Get request
		if err := xs.client.Get(nURL, &data); err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				xs._Disconnected()
			}
			common.APIError(c, err.Error())
			return
		}

		c.JSON(http.StatusOK, data)
	})
}

// PassthroughPost Used to declare a route that sends directly a POST request to XDS Server
func (xs *XdsServer) PassthroughPost(url string) {
	if xs.apiRouter == nil {
		xs.Log.Errorf("apiRouter not set !")
		return
	}

	xs.apiRouter.POST(url, func(c *gin.Context) {
		bodyReq := []byte{}
		n, err := c.Request.Body.Read(bodyReq)
		if err != nil {
			common.APIError(c, err.Error())
			return
		}

		// Take care of param (eg. id in /projects/:id)
		nURL := url
		if strings.Contains(url, ":") {
			nURL = strings.TrimPrefix(c.Request.URL.Path, xs.APIURL)
		}

		// Send Post request
		body, err := json.Marshal(bodyReq[:n])
		if err != nil {
			common.APIError(c, err.Error())
			return
		}

		response, err := xs.client.HTTPPostWithRes(nURL, string(body))
		if err != nil {
			common.APIError(c, err.Error())
			return
		}

		bodyRes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			common.APIError(c, "Cannot read response body")
			return
		}
		c.JSON(http.StatusOK, string(bodyRes))
	})
}

// EventRegister Post a request to register to an XdsServer event
func (xs *XdsServer) EventRegister(evName string, id string) error {
	return xs.client.Post(
		"/events/register",
		xsapiv1.EventRegisterArgs{
			Name:      evName,
			ProjectID: id,
		},
		nil)
}

// EventEmit Emit a event to XDS Server through WS
func (xs *XdsServer) EventEmit(message string, args ...interface{}) error {
	if xs.ioSock == nil {
		return fmt.Errorf("Io.Socket not initialized")
	}

	return xs.ioSock.Emit(message, args...)
}

// EventOn Register a callback on events reception
func (xs *XdsServer) EventOn(evName string, privData interface{}, f EventCB) (uuid.UUID, error) {
	if xs.ioSock == nil {
		return uuid.Nil, fmt.Errorf("Io.Socket not initialized")
	}

	xs.sockEventsLock.Lock()
	defer xs.sockEventsLock.Unlock()

	if _, exist := xs.sockEvents[evName]; !exist {
		// Register listener only the first time
		evn := evName

		err := xs.ioSock.On(evn, func(data interface{}) error {
				xs.sockEventsLock.Lock()
				sEvts := make([]*caller, len(xs.sockEvents[evn]))
				copy(sEvts, xs.sockEvents[evn])
				xs.sockEventsLock.Unlock()
				for _, c := range sEvts {
					c.Func(c.PrivateData, data)
				}
				return nil
			})
		if err != nil {
			return uuid.Nil, err
		}
	}

	c := &caller{
		id:          uuid.NewV1(),
		EventName:   evName,
		Func:        f,
		PrivateData: privData,
	}

	xs.sockEvents[evName] = append(xs.sockEvents[evName], c)
	xs.LogSillyf("XS EventOn: sockEvents[\"%s\"]: len %d", evName, len(xs.sockEvents[evName]))
	return c.id, nil
}

// EventOff Un-register a (or all) callbacks associated to an event
func (xs *XdsServer) EventOff(evName string, id uuid.UUID) error {
	xs.sockEventsLock.Lock()
	defer xs.sockEventsLock.Unlock()
	if _, exist := xs.sockEvents[evName]; exist {
		if id == uuid.Nil {
			// Un-register all
			xs.sockEvents[evName] = []*caller{}
		} else {
			// Un-register only the specified callback
			for i, ff := range xs.sockEvents[evName] {
				if uuid.Equal(ff.id, id) {
					xs.sockEvents[evName] = append(xs.sockEvents[evName][:i], xs.sockEvents[evName][i+1:]...)
					break
				}
			}
		}
	}
	xs.LogSillyf("XS EventOff: sockEvents[\"%s\"]: len %d", evName, len(xs.sockEvents[evName]))
	return nil
}

// ProjectToFolder Convert Project structure to Folder structure
func (xs *XdsServer) ProjectToFolder(pPrj xaapiv1.ProjectConfig) *xsapiv1.FolderConfig {
	stID := ""
	if pPrj.Type == xsapiv1.TypeCloudSync {
		stID, _ = xs.SThg.IDGet()
	}
	// TODO: limit ClientData size and gzip it (see https://golang.org/pkg/compress/gzip/)
	fPrj := xsapiv1.FolderConfig{
		ID:         pPrj.ID,
		Label:      pPrj.Label,
		ClientPath: pPrj.ClientPath,
		Type:       xsapiv1.FolderType(pPrj.Type),
		Status:     pPrj.Status,
		IsInSync:   pPrj.IsInSync,
		DefaultSdk: pPrj.DefaultSdk,
		ClientData: pPrj.ClientData,
		DataPathMap: xsapiv1.PathMapConfig{
			ServerPath: pPrj.ServerPath,
		},
		DataCloudSync: xsapiv1.CloudSyncConfig{
			SyncThingID:   stID,
			STLocIsInSync: pPrj.IsInSync,
			STLocStatus:   pPrj.Status,
			STSvrIsInSync: pPrj.IsInSync,
			STSvrStatus:   pPrj.Status,
		},
	}

	return &fPrj
}

// FolderToProject Convert Folder structure to Project structure
func (xs *XdsServer) FolderToProject(fPrj xsapiv1.FolderConfig) xaapiv1.ProjectConfig {
	inSync := fPrj.IsInSync
	sts := fPrj.Status

	if fPrj.Type == xsapiv1.TypeCloudSync {
		inSync = fPrj.DataCloudSync.STSvrIsInSync && fPrj.DataCloudSync.STLocIsInSync

		sts = fPrj.DataCloudSync.STSvrStatus
		switch fPrj.DataCloudSync.STLocStatus {
		case xaapiv1.StatusErrorConfig, xaapiv1.StatusDisable, xaapiv1.StatusPause:
			sts = fPrj.DataCloudSync.STLocStatus
			break
		case xaapiv1.StatusSyncing:
			if sts != xaapiv1.StatusErrorConfig && sts != xaapiv1.StatusDisable && sts != xaapiv1.StatusPause {
				sts = xaapiv1.StatusSyncing
			}
			break
		case xaapiv1.StatusEnable:
			// keep STSvrStatus
			break
		}
	}

	pPrj := xaapiv1.ProjectConfig{
		ID:         fPrj.ID,
		ServerID:   xs.ID,
		Label:      fPrj.Label,
		ClientPath: fPrj.ClientPath,
		ServerPath: fPrj.DataPathMap.ServerPath,
		Type:       xaapiv1.ProjectType(fPrj.Type),
		Status:     sts,
		IsInSync:   inSync,
		DefaultSdk: fPrj.DefaultSdk,
		ClientData: fPrj.ClientData,
	}
	return pPrj
}

// CommandAdd Add a new command to the list of running commands
func (xs *XdsServer) CommandAdd(cmdID string, data interface{}) error {
	if xs.CommandGet(cmdID) != nil {
		return fmt.Errorf("command id already exist")
	}
	xs.cmdList[cmdID] = data
	return nil
}

// CommandDelete Delete a command from the command list
func (xs *XdsServer) CommandDelete(cmdID string) error {
	if xs.CommandGet(cmdID) == nil {
		return fmt.Errorf("unknown command id")
	}
	delete(xs.cmdList, cmdID)
	return nil
}

// CommandGet Retrieve a command data
func (xs *XdsServer) CommandGet(cmdID string) interface{} {
	d, exist := xs.cmdList[cmdID]
	if exist {
		return d
	}
	return nil
}

/***
** Private functions
***/

// Create HTTP client
func (xs *XdsServer) _CreateConnectHTTP() error {
	var err error
	xs.client, err = common.HTTPNewClient(xs.BaseURL,
		common.HTTPClientConfig{
			URLPrefix:           "/api/v1",
			HeaderClientKeyName: "Xds-Sid",
			CsrfDisable:         true,
			LogOut:              xs.logOut,
			LogPrefix:           "XDSSERVER: ",
			LogLevel:            common.HTTPLogLevelWarning,
		})

	xs.client.SetLogLevel(xs.Log.Level.String())

	if err != nil {
		msg := ": " + err.Error()
		if strings.Contains(err.Error(), "connection refused") {
			msg = fmt.Sprintf("(url: %s)", xs.BaseURL)
		}
		return fmt.Errorf("ERROR: cannot connect to XDS Server %s", msg)
	}
	if xs.client == nil {
		return fmt.Errorf("ERROR: cannot connect to XDS Server (null client)")
	}

	return nil
}

// _Reconnect Re-established connection
func (xs *XdsServer) _Reconnect() error {
	err := xs._Connect(true)
	if err == nil {
		// Reload projects list for this server
		err = xs.projects.Init(xs)
	}
	return err
}

// _Connect Established HTTP and WS connection and retrieve XDSServer config
func (xs *XdsServer) _Connect(reConn bool) error {

	xdsCfg := xsapiv1.APIConfig{}
	if err := xs.client.Get("/config", &xdsCfg); err != nil {
		xs.Connected = false
		if !reConn {
			xs._NotifyState()
		}
		return err
	}

	if reConn && xs.ID != xdsCfg.ServerUID {
		xs.Log.Warningf("Reconnected to server but ID differs: old=%s, new=%s", xs.ID, xdsCfg.ServerUID)
	}

	// Update local XDS config
	xs.ID = xdsCfg.ServerUID
	xs.ServerConfig = &xdsCfg

	// Establish WS connection and register listen
	if err := xs._SocketConnect(); err != nil {
		xs._Disconnected()
		return err
	}

	xs.Connected = true
	xs._NotifyState()
	return nil
}

// _SocketConnect Create WebSocket (io.socket) connection
func (xs *XdsServer) _SocketConnect() error {

	xs.Log.Infof("Connecting IO.socket for server %s (url %s)", xs.ID, xs.BaseURL)

	opts := &sio_client.Options{
		Transport: "websocket",
		Header:    make(map[string][]string),
	}
	opts.Header["XDS-SID"] = []string{xs.client.GetClientID()}

	iosk, err := sio_client.NewClient(xs.BaseURL, opts)
	if err != nil {
		return fmt.Errorf("IO.socket connection error for server %s: %v", xs.ID, err)
	}
	xs.ioSock = iosk

	// Register some listeners

	iosk.On("error", func(err error) {
		xs.Log.Infof("IO.socket Error server %s; err: %v", xs.ID, err)
		if xs.CBOnError != nil {
			xs.CBOnError(err)
		}
	})

	iosk.On("disconnection", func(err error) {
		xs.Log.Infof("IO.socket disconnection server %s", xs.ID)
		if xs.CBOnDisconnect != nil {
			xs.CBOnDisconnect(err)
		}
		xs._Disconnected()

		// Try to reconnect during 15min (or at least while not disabled)
		go func() {
			count := 0
			waitTime := 1
			for !xs.Disabled && !xs.Connected {
				count++
				if count%60 == 0 {
					waitTime *= 5
				}
				if waitTime > 15*60 {
					xs.Log.Infof("Stop reconnection to server url=%s id=%s !", xs.BaseURL, xs.ID)
					return
				}
				time.Sleep(time.Second * time.Duration(waitTime))
				xs.Log.Infof("Try to reconnect to server %s (%d)", xs.BaseURL, count)

				err := xs._Reconnect()
				if err != nil &&
					!(strings.Contains(err.Error(), "dial tcp") && strings.Contains(err.Error(), "connection refused")) {
					xs.Log.Errorf("ERROR while reconnecting: %v", err.Error())
				}

			}
		}()
	})

	// XXX - There is no connection event generated so, just consider that
	// we are connected when NewClient return successfully
	/* iosk.On("connection", func() { ... }) */
	xs.Log.Infof("IO.socket connected server url=%s id=%s", xs.BaseURL, xs.ID)

	return nil
}

// _Disconnected Set XDS Server as disconnected
func (xs *XdsServer) _Disconnected() error {
	// Clear all register events as socket is closed
	for k := range xs.sockEvents {
		delete(xs.sockEvents, k)
	}
	xs.Connected = false
	xs.ioSock = nil
	xs._NotifyState()
	return nil
}

// _NotifyState Send event to notify changes
func (xs *XdsServer) _NotifyState() {

	evSts := xaapiv1.ServerCfg{
		ID:         xs.ID,
		URL:        xs.BaseURL,
		APIURL:     xs.APIURL,
		PartialURL: xs.PartialURL,
		ConnRetry:  xs.ConnRetry,
		Connected:  xs.Connected,
	}
	if err := xs.events.Emit(xaapiv1.EVTServerConfig, evSts, ""); err != nil {
		xs.Log.Warningf("Cannot notify XdsServer state change: %v", err)
	}
}
